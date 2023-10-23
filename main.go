package main

import (
	"fmt"
	"log"
	"runtime"
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
	backgroundImage := canvas.NewImageFromFile("assets/background.jpg") // Replace "background.jpg" with your image file path
	backgroundImage.FillMode = canvas.ImageFillStretch
	backgroundContainer := container.NewStack(backgroundImage)
	//// INPUT FIELDS
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
			ManifestURL: manifestURLEntry.Text,
			OutputDir:   outputDirLabel.Text,
		}
		err := cfg.SaveConfig()
		if err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation("Config Saved", "Configuration saved successfully!", myWindow)
		}
	})
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()
	var downloadButton *widget.Button
	downloadButton = widget.NewButton("Sync modpack files", func() {
		manifestURL := manifestURLEntry.Text
		if manifestURL == "" {
			dialog.ShowError(errors.New("Please enter a valid manifest URL"), myWindow)
			return
		}
		outputDir := outputDirLabel.Text
		if outputDir == "Output Directory: Not Selected" {
			dialog.ShowError(errors.New("Please choose an output directory"), myWindow)
			return
		}
		files, err := operations.DownloadManifestFiles(manifestURL)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		var wg sync.WaitGroup
		totalFiles := len(files)
		wg.Add(totalFiles)
		go func() {
			myWindow.SetContent(container.NewVBox(
				manifestURLEntry,
				outputDirLabel,
				outputDirButton,
				progressBar,
				downloadButton,
			))

			progressBar.Show()

			for _, file := range files {
				go func(file operations.File) {
					defer wg.Done()

					err := operations.DownloadFile(file, outputDir)
					if err != nil {
						dialog.ShowError(err, myWindow)
						return
					}
				}(file)
			}
			wg.Wait()
			err := operations.CleanupOutputDir(files, outputDir)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			progressBar.Hide()
			dialog.ShowInformation("Success", "Files downloaded successfully!", myWindow)
			myWindow.SetContent(container.NewVBox(
				manifestURLEntry,
				outputDirLabel,
				outputDirButton,
				downloadButton,
			))
		}()
	})
	///auto updater
	checkUpdate(myWindow)
	cfg, _ := config.LoadConfig()
	manifestURLEntry.SetText(cfg.ManifestURL)
	outputDirLabel.SetText(cfg.OutputDir)

	content := container.NewVBox(
		manifestURLEntry,
		outputDirLabel,
		outputDirButton,
		downloadButton,
		saveConfigButton,
	)

	mainContainer := container.NewBorder(nil, nil, nil, nil, backgroundContainer, content)
	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(400, 200))
	myWindow.ShowAndRun()
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
	confirmDialog := dialog.NewConfirm(fmt.Sprintf("New release %s is available to update", latestVersion), "A update was found. You can update now.", func(response bool) {
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
