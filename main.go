package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/mouuff/go-rocket-update/pkg/provider"
	"github.com/mouuff/go-rocket-update/pkg/updater"
	"github.com/pkg/errors"
)

const (
	VERSION     = "1.0.0"
	GITHUB_REPO = "https://github.com/brutalzinn/modpack-manifest-downloader"
)

type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
	Hash string `json:"hash"`
}
type Config struct {
	ManifestURL string `json:"manifest_url"`
	OutputDir   string `json:"output_dir"`
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Minecraft ModPack Manifest Downloader")
	config := &Config{}
	loadConfig(config)

	updater := &updater.Updater{
		Provider: &provider.Github{
			RepositoryURL: GITHUB_REPO,
			ArchiveName:   fmt.Sprintf("binaries_%s.zip", runtime.GOOS),
		},
		ExecutableName: fmt.Sprintf("modpack-manifest-downloader_%s_%s", runtime.GOOS, runtime.GOARCH),
		Version:        VERSION,
	}
	if _, err := updater.Update(); err != nil {
		log.Println(err)
	}

	manifestURLEntry := widget.NewEntry()
	manifestURLEntry.SetPlaceHolder("Enter Manifest URL")
	manifestURLEntry.SetText(config.ManifestURL)

	outputDirLabel := widget.NewLabel("Output Directory: Not Selected")
	outputDirButton := widget.NewButton("Choose Directory", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				outputDirLabel.SetText(uri.Path())
			}
		}, myWindow)
	})
	outputDirLabel.SetText(config.OutputDir)

	saveConfigButton := widget.NewButtonWithIcon("Save Config", theme.DocumentSaveIcon(), func() {
		config := &Config{
			ManifestURL: manifestURLEntry.Text,
			OutputDir:   outputDirLabel.Text,
		}

		err := saveConfig(config)
		if err != nil {
			dialog.ShowError(err, myWindow)
		} else {
			dialog.ShowInformation("Config Saved", "Configuration saved successfully!", myWindow)
		}
	})
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	var downloadButton *widget.Button
	downloadButton = widget.NewButton("Download", func() {
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

		files, err := downloadManifestFiles(manifestURL)
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
				go func(file File) {
					defer wg.Done()

					err := downloadFile(file, outputDir)
					if err != nil {
						dialog.ShowError(err, myWindow)
						return
					}
				}(file)
			}

			wg.Wait()

			err := cleanupOutputDir(files, outputDir)
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

	myWindow.SetContent(container.NewVBox(
		manifestURLEntry,
		outputDirLabel,
		outputDirButton,
		downloadButton,
		saveConfigButton,
	))

	myWindow.Resize(fyne.NewSize(400, 500))
	myWindow.ShowAndRun()
}
