package pdf

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var pdfClient = &http.Client{
	Timeout: 1 * time.Minute,
}

// IsPDF checks if the URL points to a PDF file based on its extension.
func IsPDF(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return filepath.Ext(parsedURL.Path) == ".pdf"
}

// Download downloads a PDF file from the given URL and returns the path to the downloaded file.
func Download(urlStr string) (filePath string, err error) {
	logger := slog.With("url", urlStr)
	logger.Debug("Downloading file from URL")

	resp, err := pdfClient.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("expected success status code, was %s", resp.Status)
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "linkding-pdf-")
	if err != nil {
		return "", err
	}
	// Clean up temporary directory on error
	defer func() {
		if err != nil {
			os.RemoveAll(tempDir)
		}
	}()

	// Extract filename from URL path
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	filename := filepath.Base(parsedURL.Path)
	logger.Debug("Extracted filename from URL", "filename", filename, "ext", filepath.Ext(filename))
	// Use default filename if extraction fails or doesn't have a file extension
	if filename == "." || filename == "/" || filename == "" || filepath.Ext(filename) == "" {
		filename = "download.pdf"
		logger.Debug("Using default filename", "filename", filename)
	}

	// Create file in temporary directory
	filePath = filepath.Join(tempDir, filename)
	tempFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy response body to file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}

	logger.Debug("File downloaded successfully", "path", filePath)
	return filePath, nil
}
