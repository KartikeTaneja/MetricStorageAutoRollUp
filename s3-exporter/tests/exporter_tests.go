package tests

import (
	"os"
	"path/filepath"
	"testing"

	"s3-exporter/exporter"
)

// TestLoadConfig tests the LoadConfig function
func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	
	configContent := `
s3:
  region: test-region
  bucket: test-bucket
  access_key: test-access-key
  secret_key: test-secret-key
export:
  batch_size: 500
  compression: true
  temp_dir: /tmp/test
logging:
  level: debug
  format: json
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Load the config
	config, err := exporter.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// Verify config values
	if config.S3.Region != "test-region" {
		t.Errorf("Expected region to be 'test-region', got '%s'", config.S3.Region)
	}
	
	if config.S3.Bucket != "test-bucket" {
		t.Errorf("Expected bucket to be 'test-bucket', got '%s'", config.S3.Bucket)
	}
	
	if config.Export.BatchSize != 500 {
		t.Errorf("Expected batch_size to be 500, got %d", config.Export.BatchSize)
	}
	
	if !config.Export.Compression {
		t.Errorf("Expected compression to be true")
	}
}

// TestCheckIfExported tests the CheckIfExported function
func TestCheckIfExported(t *testing.T) {
	// Create a temporary SFM file with the exported flag set to true
	tempDir := t.TempDir()
	exportedFile := filepath.Join(tempDir, "exported.sfm")
	
	exportedContent := `
# header
name,age,location
jsonS3Exported:true
John,30,New York
`
	
	err := os.WriteFile(exportedFile, []byte(exportedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test exported file: %v", err)
	}
	
	// Create a temporary SFM file with the exported flag set to false
	notExportedFile := filepath.Join(tempDir, "not_exported.sfm")
	
	notExportedContent := `
# header
name,age,location
jsonS3Exported:false
Jane,25,Los Angeles
`
	
	err = os.WriteFile(notExportedFile, []byte(notExportedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test not exported file: %v", err)
	}
	
	// Test the exported file
	exported, err := exporter.CheckIfExported(exportedFile)
	if err != nil {
		t.Fatalf("CheckIfExported failed: %v", err)
	}
	
	if !exported {
		t.Errorf("Expected exported to be true for '%s'", exportedFile)
	}
	
	// Test the not exported file
	exported, err = exporter.CheckIfExported(notExportedFile)
	if err != nil {
		t.Fatalf("CheckIfExported failed: %v", err)
	}
	
	if exported {
		t.Errorf("Expected exported to be false for '%s'", notExportedFile)
	}
}

// TestMarkAsExported tests the MarkAsExported function
func TestMarkAsExported(t *testing.T) {
	// Create a temporary SFM file with the exported flag set to false
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.sfm")
	
	testContent := `
# header
name,age,location
jsonS3Exported:false
Jane,25,Los Angeles
`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Mark the file as exported
	err = exporter.MarkAsExported(testFile)
	if err != nil {
		t.Fatalf("MarkAsExported failed: %v", err)
	}
	
	// Verify the file was updated
	exported, err := exporter.CheckIfExported(testFile)
	if err != nil {
		t.Fatalf("CheckIfExported failed: %v", err)
	}
	
	if !exported {
		t.Errorf("Expected file to be marked as exported")
	}
}