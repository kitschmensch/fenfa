package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Encode(input string, salt string, timestamp int64) string {
	var s string = fmt.Sprintf("%s%d%s", input, timestamp, salt)
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func ZipDirectory(dirPath string, maxDepth int) (string, error) {
	zipPath := dirPath + ".zip"
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("could not create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = addFilesToZip(zipWriter, dirPath, "", 0, maxDepth)
	if err != nil {
		return "", fmt.Errorf("could not zip directory: %v", err)
	}

	return zipPath, nil
}

func addFilesToZip(zipWriter *zip.Writer, basePath, relativePath string, currentDepth, maxDepth int) error {
	if maxDepth >= 0 && currentDepth > maxDepth {
		return nil
	}

	fullPath := filepath.Join(basePath, relativePath)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("could not stat file: %v", err)
	}

	if fileInfo.IsDir() {
		if relativePath != "" {
			_, err := zipWriter.Create(relativePath + "/")
			if err != nil {
				return fmt.Errorf("could not create directory in zip: %v", err)
			}
		}

		files, err := os.ReadDir(fullPath)
		if err != nil {
			return fmt.Errorf("could not read directory: %v", err)
		}

		for _, file := range files {
			err = addFilesToZip(zipWriter, basePath, filepath.Join(relativePath, file.Name()), currentDepth+1, maxDepth)
			if err != nil {
				return err
			}
		}
	} else {
		zipFileWriter, err := zipWriter.Create(relativePath)
		if err != nil {
			return fmt.Errorf("could not create file in zip: %v", err)
		}

		sourceFile, err := os.Open(fullPath)
		if err != nil {
			return fmt.Errorf("could not open file: %v", err)
		}
		defer sourceFile.Close()

		_, err = io.Copy(zipFileWriter, sourceFile)
		if err != nil {
			return fmt.Errorf("could not copy file to zip: %v", err)
		}
	}

	return nil
}

func EstimateZipSize(dirPath string, maxDepth int) (int64, error) {
	return estimateSizeHelper(dirPath, "", 0, maxDepth)
}

func estimateSizeHelper(basePath, relativePath string, currentDepth, maxDepth int) (int64, error) {
	if maxDepth >= 0 && currentDepth > maxDepth {
		return 0, nil
	}

	fullPath := filepath.Join(basePath, relativePath)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("could not stat file: %v", err)
	}

	totalSize := int64(0)

	if fileInfo.IsDir() {
		files, err := os.ReadDir(fullPath)
		if err != nil {
			return 0, fmt.Errorf("could not read directory: %v", err)
		}

		for _, file := range files {
			size, err := estimateSizeHelper(basePath, filepath.Join(relativePath, file.Name()), currentDepth+1, maxDepth)
			if err != nil {
				return 0, err
			}
			totalSize += size
		}
	} else {
		totalSize += fileInfo.Size()
	}

	return totalSize, nil
}
