package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/brutalzinn/manifest-downloader/config"
	"github.com/brutalzinn/manifest-downloader/operations"
	"github.com/brutalzinn/manifest-downloader/progress"
	"github.com/mouuff/go-rocket-update/pkg/provider"
	"github.com/mouuff/go-rocket-update/pkg/updater"
	"github.com/pkg/errors"
)

const (
	VERSION     = "v0.0.4"
	GITHUB_REPO = "github.com/brutalzinn/modpack-manifest-downloader"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow(fmt.Sprintf("Minecraft ModPack Manifest Downloader %s", VERSION))
	backgroundImage := canvas.NewImageFromFile("./assets/background.jpg") // Replace "background.jpg" with your image file path
	backgroundImage.FillMode = canvas.ImageFillStretch
	backgroundContainer := container.NewStack(backgroundImage)
	cfg, _ := config.LoadConfig()
	progressBar := widget.NewProgressBar()
	//// INPUT FIELDS
	ignoreFolders := widget.NewEntry()
	ignoreFolders.SetPlaceHolder("Ignore directories that doesnt need be verified. Separe dirs using ','")
	progressText := widget.NewLabel("...")
	manifestURLEntry := widget.NewEntry()
	manifestURLEntry.SetPlaceHolder("Enter Manifest URL")
	outputDirLabel := widget.NewLabel("Output Directory: Not Selected")
	outputDirButton := widget.NewButton("Choose the Minecraft Directory", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				outputDirLabel.SetText(uri.Path())
			}
		}, myWindow)
	})
	saveConfigButton := widget.NewButtonWithIcon("Save Config", theme.DocumentSaveIcon(), func() {
		cfg := &config.Config{
			ManifestURL:   manifestURLEntry.Text,
			OutputDir:     outputDirLabel.Text,
			IgnoreFolders: ignoreFolders.Text,
		}
		err := cfg.SaveConfig()
		if err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation("Config Saved", "Configuration saved successfully!", myWindow)
		}
	})

	downloadButton := widget.NewButton("Sync modpack files", func() {
		manifestURL := manifestURLEntry.Text
		outputDir := outputDirLabel.Text
		ignoreFolders := strings.Split(ignoreFolders.Text, ",")
		progressBar.Show()
		err := startSync(manifestURL, outputDir, ignoreFolders, func(progress *progress.Progress) {
			progressBar.Max = float64(progress.Max)
			progressBar.Min = 0
			progressBar.SetValue(float64(progress.Value))
			progressText.SetText(progress.Text)
			if progress.Completed {
				progressBar.Hide()
				dialog.ShowInformation("Success", "Files downloaded successfully!", myWindow)
			}
		})
		if err != nil {
			progressBar.Hide()
			dialog.ShowError(err, myWindow)
		}

	})
	///set loaded config to fields
	manifestURLEntry.SetText(cfg.ManifestURL)
	outputDirLabel.SetText(cfg.OutputDir)
	ignoreFolders.SetText(cfg.IgnoreFolders)
	///auto updater
	checkUpdate(myWindow)

	content := container.NewVBox(
		manifestURLEntry,
		ignoreFolders,
		outputDirLabel,
		outputDirButton,
		progressText,
		progressBar,
		downloadButton,
		saveConfigButton,
	)

	mainContainer := container.NewBorder(nil, nil, nil, nil, backgroundContainer, content)
	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(400, 200))
	myWindow.ShowAndRun()
}

func startSync(manifestUrl string, outputDir string, ignoreFolders []string, onProgress func(progress *progress.Progress)) error {
	if manifestUrl == "" {
		return errors.Errorf("Please enter a valid manifest URL")
	}
	if outputDir == "Output Directory: Not Selected" {
		return errors.Errorf("Please choose an output directory")
	}
	files, err := operations.ReadManifestFiles(manifestUrl)
	if err != nil {
		return errors.Errorf("Error on dowload manifest files")
	}
	var wg sync.WaitGroup
	totalFiles := len(files)
	wg.Add(totalFiles)
	progressDownload := progress.New(totalFiles)
	progressDownload.SetText("downloading files..")
	go func() error {
		for _, file := range files {
			go func(file operations.File) error {
				defer wg.Done()
				err := operations.DownloadFile(file, outputDir)
				if err != nil {
					return err
				}
				progressDownload.Done()
				onProgress(progressDownload)
				return nil
			}(file)
		}
		wg.Wait()
		err := operations.CleanupOutputDir(files, outputDir, ignoreFolders,
			func(progress *progress.Progress) {
				onProgress(progress)
			})
		if err != nil {
			return err
		}
		progressDownload.Complete()
		onProgress(progressDownload)
		return nil
	}()
	return nil
}

func checkUpdate(window fyne.Window) {
	u := &updater.Updater{
		Provider: &provider.Github{
			RepositoryURL: GITHUB_REPO,
			ArchiveName:   fmt.Sprintf("binaries_%s.zip", runtime.GOOS),
		},
		ExecutableName: fmt.Sprintf("modpack-manifest-downloader_%s_%s", runtime.GOOS, runtime.GOARCH),
		Version:        VERSION,
	}
	latestVersion, _ := u.GetLatestVersion()
	confirmDialog := dialog.NewConfirm(fmt.Sprintf("New release %s is available to download", latestVersion), "A update was found. You can update now.", func(response bool) {
		if response {
			log.Println("start update")
			u.Update()
		} else {
			log.Println("no update now")
		}
	}, window)
	if canUpdate, _ := u.CanUpdate(); canUpdate {
		confirmDialog.Show()
	}
}
