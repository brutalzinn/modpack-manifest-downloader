package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func downloadManifestFiles(url string) ([]File, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch manifest")
	}
	defer response.Body.Close()

	var files []File
	err = json.NewDecoder(response.Body).Decode(&files)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse manifest")
	}

	return files, nil
}

func downloadFile(file File, outputDir string) error {
	outFilePath := filepath.Join(outputDir, file.Name)
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create file: %s", outFilePath)
	}
	defer outFile.Close()

	response, err := http.Get(file.URL)
	if err != nil {
		return errors.Wrapf(err, "failed to download file: %s", file.Name)
	}
	defer response.Body.Close()
	_, err = io.Copy(io.MultiWriter(outFile), response.Body)
	if err != nil {
		return errors.Wrapf(err, "failed to save file: %s", file.Name)
	}
	return nil
}
func cleanupOutputDir(manifestFiles []File, outputDir string) error {
	manifestFileSet := make(map[string]bool)
	for _, file := range manifestFiles {
		manifestFileSet[file.Path] = true
	}

	err := filepath.WalkDir(outputDir, func(filePath string, fileInfo os.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error accessing path %s", filePath)
		}

		if !fileInfo.IsDir() {
			relativePath, err := filepath.Rel(outputDir, filePath)
			if err != nil {
				return errors.Wrapf(err, "failed to calculate relative path for %s", filePath)
			}

			if _, exists := manifestFileSet[relativePath]; !exists {
				// File not found in manifest, remove it
				err := os.Remove(filePath)
				if err != nil {
					return errors.Wrapf(err, "failed to delete file: %s", filePath)
				}
			}
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "cleanup failed")
	}

	return nil
}

func calculateFileHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}
