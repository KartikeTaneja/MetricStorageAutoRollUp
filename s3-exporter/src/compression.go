package src

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"path/filepath"
)

// CompressFile compresses a file using gzip and returns the compressed file path
func CompressFile(filePath string) (string, error) {
	// Open the source file
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Create the destination file
	destPath := filePath + ".gz"
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	// Copy data from source to gzip writer
	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		return "", fmt.Errorf("error compressing file: %w", err)
	}

	// Flush and close the gzip writer explicitly before returning
	err = gzipWriter.Close()
	if err != nil {
		return "", fmt.Errorf("error closing gzip writer: %w", err)
	}

	// Check if the compression was successful by comparing file sizes
	sourceInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting source file info: %w", err)
	}

	destInfo, err := os.Stat(destPath)
	if err != nil {
		return "", fmt.Errorf("error getting destination file info: %w", err)
	}

	// If compressed file is larger than source, use the original file
	if destInfo.Size() >= sourceInfo.Size() {
		os.Remove(destPath)
		return filePath, nil
	}

	return destPath, nil
}

// DecompressFile decompresses a gzip file and returns the path to the decompressed file
func DecompressFile(filePath string) (string, error) {
	// Check if the file is a gzip file
	if filepath.Ext(filePath) != ".gz" {
		return "", fmt.Errorf("file is not a gzip file: %s", filePath)
	}

	// Open the source file
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(sourceFile)
	if err != nil {
		return "", fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create the destination file (without .gz extension)
	destPath := strings.TrimSuffix(filePath, ".gz")
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	// Copy data from gzip reader to destination file
	_, err = io.Copy(destFile, gzipReader)
	if err != nil {
		return "", fmt.Errorf("error decompressing file: %w", err)
	}

	return destPath, nil
}

// CompressDirectory compresses all files in a directory
func CompressDirectory(dirPath string) ([]string, error) {
	// Get a list of all files in the directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	compressedFiles := make([]string, 0)

	// Compress each file
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		
		// Skip already compressed files
		if filepath.Ext(filePath) == ".gz" {
			compressedFiles = append(compressedFiles, filePath)
			continue
		}

		compressedPath, err := CompressFile(filePath)
		if err != nil {
			return compressedFiles, fmt.Errorf("error compressing %s: %w", filePath, err)
		}

		compressedFiles = append(compressedFiles, compressedPath)
	}

	return compressedFiles, nil
}