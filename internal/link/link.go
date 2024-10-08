package link

import (
	"fenfa/internal/config"
	"fenfa/internal/store"
	"fenfa/pkg/utils"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func GenerateFileLink(path string) {
	path = filepath.Clean(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("File does not exist: %s", path)
		fmt.Printf("Error: File does not exist: %s\n", path)
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		log.Printf("Error checking file information: %s", path)
		fmt.Printf("Error: %v\n", err)
		return
	}

	if info.IsDir() {
		estimatedSize, err := utils.EstimateZipSize(path, config.MaxZipDepth)
		if err != nil {
			log.Printf("Error checking file information: %s", path)
			fmt.Printf("Error estimating zip size: %v\n", err)
			return
		}

		if estimatedSize > config.MaxZipSize {
			log.Printf("Error checking file information: %s", path)
			fmt.Printf("Error: Directory size exceeds the limit of %d bytes\n", config.MaxZipSize)
			return
		}

		zipPath, err := utils.ZipDirectory(path, config.MaxZipDepth)
		if err != nil {
			log.Printf("Error checking file information: %s", path)
			fmt.Printf("Error zipping directory: %v\n", err)
			return
		}

		err = os.MkdirAll(config.ZipDirectory, 0755)
		if err != nil {
			log.Printf("Error checking file information: %s", path)
			fmt.Printf("Error: Could not create directory: %v\n", err)
			return
		}

		finalZipPath := filepath.Join(config.ZipDirectory, filepath.Base(zipPath))
		err = os.Rename(zipPath, finalZipPath)
		if err != nil {
			log.Printf("Error checking file information: %s", path)
			fmt.Printf("Error: Could not move zip file: %v\n", err)
			return
		}

		path = finalZipPath
	}

	expiration := time.Now().Add(time.Duration(config.DefaultExpirationPeriod) * time.Second).Unix()
	hash := utils.Encode(path, config.Salt, expiration)
	store.Add(hash, expiration, path)
	var url string
	if config.TemplateIncludesPort {
		// Format with port
		url = fmt.Sprintf("%s:%d/%s", config.Host, config.Port, hash)
	} else {
		// Format without port
		url = fmt.Sprintf("%s/%s", config.Host, hash)
	}
	log.Printf("Generated link: %s for file: %s", url, path)
	fmt.Println(url)
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Error getting IP: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	failedAttempts, err := store.GetFailedAttempts(ip)
	if err != nil {
		log.Printf("Error getting failed attempts for IP %s: %v", ip, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if failedAttempts >= config.FailedAttemptLimit {
		log.Printf("Banned IP address: %s", ip)
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}
	_, hash := filepath.Split(r.URL.Path)
	entry, active, exists := store.Get(hash)
	if !exists {
		store.IncrementFailedAttempts(ip)
		log.Printf("Hash not found in map: %s", hash)
		http.NotFound(w, r)
		return
	}
	if !active {
		store.IncrementFailedAttempts(ip)
		log.Printf("Attempted access of expired link by %s: %s", ip, hash)
		http.Error(w, "Link Expired.", http.StatusGone)
		return
	}

	if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
		store.IncrementFailedAttempts(ip)
		log.Printf("File not found at path: %s", entry.Path)
		store.Delete(hash)
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Printf("Error accessing file at path: %s: %v", entry.Path, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Serving file: %s for hash: %s", entry.Path, hash)
	http.ServeFile(w, r, entry.Path)
}
