package exporter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"s3-exporter/src"
)

// CheckIfExported checks if a segment file has already been exported
func CheckIfExported(sfmFile string) (bool, error) {
	// Open the SFM file
	file, err := os.Open(sfmFile)
	if err != nil {
		return false, fmt.Errorf("error opening SFM file: %w", err)
	}
	defer file.Close()

	// Look for the jsonS3Exported flag
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "jsonS3Exported") {
			// Check if the flag is true
			if strings.Contains(line, "jsonS3Exported:true") || 
               strings.Contains(line, "jsonS3Exported: true") {
				return true, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading SFM file: %w", err)
	}

	return false, nil
}

// MarkAsExported updates the SFM file to mark it as exported
func MarkAsExported(sfmFile string) error {
	// Read the entire file
	data, err := os.ReadFile(sfmFile)
	if err != nil {
		return fmt.Errorf("error reading SFM file: %w", err)
	}

	content := string(data)

	// Check if the jsonS3Exported flag exists
	if strings.Contains(content, "jsonS3Exported:false") {
		content = strings.Replace(content, "jsonS3Exported:false", "jsonS3Exported:true", 1)
	} else if strings.Contains(content, "jsonS3Exported: false") {
		content = strings.Replace(content, "jsonS3Exported: false", "jsonS3Exported: true", 1)
	} else {
		// If the flag doesn't exist, add it
		if strings.Contains(content, ".sfm") || strings.Contains(content, "segmeta.json") {
			// Find a suitable location to add the flag
			// This is a simplification - you might need more robust logic
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.Contains(line, ".sfm") || strings.Contains(line, "segmeta.json") {
					lines[i] = line + "\njsonS3Exported:true"
					break
				}
			}
			content = strings.Join(lines, "\n")
		}
	}

	// Write the modified content back to the file
	err = os.WriteFile(sfmFile, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error updating SFM file: %w", err)
	}

	return nil
}

// ConvertAndUpload converts an SFM file to JSON and uploads it to S3
func ConvertAndUpload(sfmFile string, config *Config) error {
	// Create temp directory if it doesn't exist
	err := os.MkdirAll(config.Export.TempDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}

	// Create a temporary file for the JSON output
	baseFileName := filepath.Base(sfmFile)
	baseFileName = strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))
	timeStamp := time.Now().Format("20060102-150405")
	jsonFileName := fmt.Sprintf("%s/%s-%s.json", config.Export.TempDir, baseFileName, timeStamp)
	
	jsonFile, err := os.Create(jsonFileName)
	if err != nil {
		return fmt.Errorf("error creating JSON file: %w", err)
	}
	defer jsonFile.Close()

	// Open the SFM file
	sfmReader, err := os.Open(sfmFile)
	if err != nil {
		return fmt.Errorf("error opening SFM file: %w", err)
	}
	defer sfmReader.Close()

	// Read column names from the SFM file
	columnNames, err := readColumnNames(sfmReader)
	if err != nil {
		return fmt.Errorf("error reading column names: %w", err)
	}

	// Reset file pointer to beginning
	_, err = sfmReader.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error resetting file pointer: %w", err)
	}

	scanner := bufio.NewScanner(sfmReader)
	writer := bufio.NewWriter(jsonFile)
	recordCount := 0
	batchCount := 0

	// Process each record
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header lines or non-data lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Process data line and convert to JSON
		record := strings.Split(line, ",")
		if len(record) != len(columnNames) {
			continue // Skip malformed records
		}

		jsonRecord := make(map[string]string)
		for i, value := range record {
			jsonRecord[columnNames[i]] = strings.TrimSpace(value)
		}

		// Convert to JSON
		jsonData, err := json.Marshal(jsonRecord)
		if err != nil {
			return fmt.Errorf("error marshaling to JSON: %w", err)
		}

		// Write to file
		_, err = writer.WriteString(string(jsonData) + "\n")
		if err != nil {
			return fmt.Errorf("error writing to JSON file: %w", err)
		}

		recordCount++
		
		// Flush every N records to avoid memory issues
		if recordCount%1000 == 0 {
			err = writer.Flush()
			if err != nil {
				return fmt.Errorf("error flushing to file: %w", err)
			}
		}

		// Check if we need to start a new batch
		if config.Export.BatchSize > 0 && recordCount >= config.Export.BatchSize {
			err = writer.Flush()
			if err != nil {
				return fmt.Errorf("error flushing to file: %w", err)
			}
			
			// Close current file
			jsonFile.Close()
			
			// Compress if needed
			finalFile := jsonFileName
			if config.Export.Compression {
				compressedFile, err := src.CompressFile(jsonFileName)
				if err != nil {
					return fmt.Errorf("error compressing file: %w", err)
				}
				finalFile = compressedFile
			}
			
			// Upload to S3
			s3Path := fmt.Sprintf("%s/batch-%d.json", baseFileName, batchCount)
			if strings.HasSuffix(finalFile, ".gz") {
				s3Path += ".gz"
			}
			
			err = src.UploadToS3(finalFile, s3Path, config.S3.Bucket, config.S3.Region, 
                            config.S3.AccessKey, config.S3.SecretKey)
			if err != nil {
				return fmt.Errorf("error uploading to S3: %w", err)
			}
			
			// Start a new batch
			batchCount++
			recordCount = 0
			jsonFileName = fmt.Sprintf("%s/%s-%s-batch-%d.json", 
                            config.Export.TempDir, baseFileName, timeStamp, batchCount)
			
			jsonFile, err = os.Create(jsonFileName)
			if err != nil {
				return fmt.Errorf("error creating JSON file: %w", err)
			}
			writer = bufio.NewWriter(jsonFile)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading SFM file: %w", err)
	}

	// Flush any remaining records
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing to file: %w", err)
	}
	
	// Close the final file
	jsonFile.Close()

	// Process the final batch if there's any data
	if recordCount > 0 {
		// Compress if needed
		finalFile := jsonFileName
		if config.Export.Compression {
			compressedFile, err := src.CompressFile(jsonFileName)
			if err != nil {
				return fmt.Errorf("error compressing file: %w", err)
			}
			finalFile = compressedFile
		}
		
		// Upload to S3
		s3Path := fmt.Sprintf("%s/batch-%d.json", baseFileName, batchCount)
		if strings.HasSuffix(finalFile, ".gz") {
			s3Path += ".gz"
		}
		
		err = src.UploadToS3(finalFile, s3Path, config.S3.Bucket, config.S3.Region, 
                        config.S3.AccessKey, config.S3.SecretKey)
		if err != nil {
			return fmt.Errorf("error uploading to S3: %w", err)
		}
	}

	return nil
}

// readColumnNames reads column names from the SFM file header
func readColumnNames(file *os.File) ([]string, error) {
	scanner := bufio.NewScanner(file)
	
	// Look for the header line (typically starts with # or similar)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") && strings.Contains(line, ",") {
			// Remove the # prefix and split by comma
			headerLine := strings.TrimPrefix(line, "#")
			columns := strings.Split(headerLine, ",")
			
			// Trim whitespace from column names
			for i, col := range columns {
				columns[i] = strings.TrimSpace(col)
			}
			
			return columns, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %w", err)
	}

	return nil, fmt.Errorf("column names not found in file header")
}