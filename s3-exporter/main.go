package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"s3-exporter/exporter"
)

func main() {
	// Setup command-line flags
	configFile := flag.String("config", "config/config.yaml", "Path to configuration file")
	dataDir := flag.String("data", "data", "Directory containing SFM files")
	logFile := flag.String("log", "logs/app.log", "Path to log file")
	flag.Parse()

	// Set up logging
	f, err := os.OpenFile(*logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("S3 Exporter started")

	// Load configuration
	config, err := exporter.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Find all SFM files
	var sfmFiles []string
	err = filepath.Walk(*dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".sfm" {
			sfmFiles = append(sfmFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error finding SFM files: %v", err)
	}

	// Process each SFM file
	for _, sfmFile := range sfmFiles {
		log.Printf("Processing SFM file: %s", sfmFile)
		
		// Check if the file has already been exported
		exported, err := exporter.CheckIfExported(sfmFile)
		if err != nil {
			log.Printf("Error checking export status for %s: %v", sfmFile, err)
			continue
		}
		
		if exported {
			log.Printf("File %s already exported, skipping", sfmFile)
			continue
		}
		
		// Start the conversion process
		err = exporter.ConvertAndUpload(sfmFile, config)
		if err != nil {
			log.Printf("Error processing %s: %v", sfmFile, err)
			continue
		}
		
		// Mark as exported
		err = exporter.MarkAsExported(sfmFile)
		if err != nil {
			log.Printf("Error marking %s as exported: %v", sfmFile, err)
		}
	}

	fmt.Println("S3 Export process completed. Check logs for details.")
}