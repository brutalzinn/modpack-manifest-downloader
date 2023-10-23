package operations

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"url"`
	Hash string `json:"hash"`
}

func ReadManifestFiles(url string) ([]File, error) {
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

func DownloadFile(file File, outputDir string) error {
	outFilePath := filepath.Join(outputDir, file.Path)
	dirName, _ := filepath.Split(outFilePath)
	err := os.MkdirAll(dirName, 0744)
	if err != nil {
		return errors.Wrapf(err, "failed to create dir: %s", outFilePath)
	}
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
func CleanupOutputDir(manifestFiles []File, outputDir string, ignoreFolders []string) error {
	manifestFileSet := make(map[string]string)
	for _, file := range manifestFiles {
		manifestFileSet[file.Path] = file.Hash
	}

	err := filepath.WalkDir(outputDir, func(filePath string, fileInfo os.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error accessing path %s", filePath)
		}
		if slices.Contains(ignoreFolders, fileInfo.Name()) {
			log.Println("Ignore ", fileInfo.Name(), "because is ignored by ignore folders", ignoreFolders)
			return filepath.SkipDir
		}
		if !fileInfo.IsDir() {
			relativePath, err := filepath.Rel(outputDir, filePath)
			if err != nil {
				return errors.Wrapf(err, "failed to calculate relative path for %s", filePath)
			}

			if expectedHash, exists := manifestFileSet[relativePath]; !exists {
				err := os.Remove(filePath)
				log.Println("remove filepath", filePath, relativePath)
				if err != nil {
					return errors.Wrapf(err, "failed to delete file: %s", filePath)
				}
			} else {
				fileHash, err := CalculateFileHash(filePath)
				if err != nil {
					return errors.Wrapf(err, "failed to calculate hash for file: %s", filePath)
				}
				if fileHash != expectedHash {
					err := os.Remove(filePath)
					if err != nil {
						return errors.Wrapf(err, "failed to delete file: %s", filePath)
					}
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

func CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	hashCalc := hex.EncodeToString(hash.Sum(nil))
	return hashCalc, nil
}
