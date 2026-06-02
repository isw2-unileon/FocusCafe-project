package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadPdf_FileNotFound(t *testing.T) {
	_, err := ReadPdf("non_existent_file_path_12345.pdf")
	if err == nil {
		t.Fatalf("Expected an error when opening a non-existent file, got nil")
	}

	if !strings.Contains(err.Error(), "error trying to open pdf") {
		t.Errorf("Unexpected error context message: %v", err)
	}
}

func TestReadPdf_InvalidPDFStructure(t *testing.T) {
	// Create a temporary directory for the mock file lifecycle
	tmpDir, err := os.MkdirTemp("", "pdf_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFilePath := filepath.Join(tmpDir, "corrupted.pdf")
	corruptedContent := []byte("this is not a valid pdf binary structure header data")

	err = os.WriteFile(tmpFilePath, corruptedContent, 0o644)
	if err != nil {
		t.Fatalf("Failed to write mock corrupted PDF asset file: %v", err)
	}

	_, err = ReadPdf(tmpFilePath)
	if err == nil {
		t.Fatalf("Expected parsing error on structural validation constraints, got nil")
	}

	// The library fails to open or parse corrupted payloads internally, returning wrapped context
	if !strings.Contains(err.Error(), "error trying to open pdf") && !strings.Contains(err.Error(), "error trying to read plain text") {
		t.Errorf("Unexpected error wrapping execution track layout: %v", err)
	}
}
