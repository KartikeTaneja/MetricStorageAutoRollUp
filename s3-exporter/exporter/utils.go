package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateOutputFileName generates a filename for the exported JSON
func GenerateOutputFileName(sfmFile string, batchNum int) string {
	baseName := filepath.Base(sfmFile)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	timestamp := time.Now().Format("20060102-150405")
	
	if batchNum > 0 {
		return fmt.Sprintf("%s-%s-batch-%d.json", baseName, timestamp, batchNum)
	}
	return fmt.Sprintf("%s-%s.json", baseName, timestamp)
}

// EnsureDirectoryExists ensures that a directory exists
func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// CleanupTempFiles removes temporary files created during processing
func CleanupTempFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}

// ParseSfmLine parses a line from an SFM file and returns a map of values
func ParseSfmLine(line string, columnNames []string) (map[string]string, error) {
	// Skip comments and empty lines
	if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
		return nil, nil
	}
	
	// Split the line into fields
	fields := strings.Split(line, ",")
	
	// Check if we have the correct number of fields
	if len(fields) != len(columnNames) {
		return nil, fmt.Errorf("field count mismatch: expected %d, got %d", len(columnNames), len(fields))
	}
	
	// Create a map of field name to value
	result := make(map[string]string)
	for i, name := range columnNames {
		result[name] = strings.TrimSpace(fields[i])
	}
	
	return result, nil
}

// FormatBytes formats a byte count as a human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// IsCompressible checks if a file should be compressed based on its extension
func IsCompressible(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	// Don't compress already compressed files
	if ext == ".gz" || ext == ".zip" || ext == ".bz2" || ext == ".xz" {
		return false
	}
	return true
}