package utils

import (
	"archive/zip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GenerateRandomSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}
	return hex.EncodeToString(salt), nil
}

func Encode(input string) (string, error) {
	salt, err := GenerateRandomSalt(16)
	if err != nil {
		return "", err
	}

	key := []byte(salt)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(input))

	return hex.EncodeToString(h.Sum(nil)), nil
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ResolveToAbsolutePath(path string) (string, error) {
	if filepath.IsAbs(path) || fileExists(path) {
		return filepath.Abs(path)
	}

	pathEnv := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(pathEnv) {
		fullPath := filepath.Join(dir, path)
		if fileExists(fullPath) {
			return filepath.Abs(fullPath)
		}
	}

	return "", fmt.Errorf("file not found: %s", path)
}
