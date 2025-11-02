package pdf

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsPDF(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/document.pdf", true},
		{"https://example.com/document.PDF", false},
		{"https://example.com/document.txt", false},
		{"https://example.com/document", false},
		{"https://example.com/path/to/file.pdf", true},
		{"https://example.com/file.pdf?query=param", true}, // Query params should not affect URL path parsing
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			actual := IsPDF(tt.url)
			if actual != tt.expected {
				t.Errorf("IsPDF(%q) = %v, want %v", tt.url, actual, tt.expected)
			}
		})
	}
}

func TestDownload(t *testing.T) {
	// Create a test HTTP server that serves a PDF
	pdfContent := []byte("%PDF-1.4\nTest PDF content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(pdfContent)
	}))
	defer server.Close()

	url := server.URL + "/test-document.pdf"
	path, err := Download(url)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(filepath.Dir(path))
	})

	// Verify the file was created
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("File not found at %s: %v", path, err)
	}

	// Verify the filename
	expectedName := "test-document.pdf"
	actualName := filepath.Base(stat.Name())
	if actualName != expectedName {
		t.Errorf("Unexpected filename: got %s, want %s", actualName, expectedName)
	}

	// Verify the content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(pdfContent) {
		t.Errorf("Unexpected content: got %q, want %q", content, pdfContent)
	}
}

func TestDownloadWithHTTPError(t *testing.T) {
	// Create a test HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	url := server.URL + "/nonexistent.pdf"
	path, err := Download(url)

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if path != "" {
		t.Errorf("Expected empty path, got %s", path)
	}

	// Verify temp directory was cleaned up
	if path != "" {
		tempDir := filepath.Dir(path)
		if _, err := os.Stat(tempDir); err == nil {
			t.Errorf("Expected temp directory to be cleaned up, but it still exists: %s", tempDir)
		}
	}
}

func TestDownloadWithNoFilename(t *testing.T) {
	// Create a test HTTP server with a URL that has no clear filename
	pdfContent := []byte("%PDF-1.4\nTest PDF content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write(pdfContent)
	}))
	defer server.Close()

	// Use the server URL with a trailing slash to ensure no filename
	url := server.URL + "/"
	path, err := Download(url)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(filepath.Dir(path))
	})

	// Verify the default filename was used
	expectedName := "download.pdf"
	actualName := filepath.Base(path)
	if actualName != expectedName {
		t.Errorf("Unexpected filename: got %s, want %s", actualName, expectedName)
	}
}
