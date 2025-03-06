#!/bin/bash

# S3 Exporter Setup Script
# This script sets up the S3 Exporter environment

set -e

# Create necessary directories
mkdir -p config
mkdir -p data
mkdir -p logs
mkdir -p temp

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Go is not installed. Please install Go and try again."
    exit 1
fi

# Install required dependencies
echo "Installing required Go dependencies..."
go mod tidy

# Create default config file if it doesn't exist
if [ ! -f config/config.yaml ]; then
    echo "Creating default config file..."
    cat > config/config.yaml << EOF
# S3 Configuration
s3:
  region: us-east-1
  bucket: sigmenbucket


# Export Configuration
export:
  batch_size: 1000  # Number of JSON lines per file
  compression: true # Whether to compress files before upload
  temp_dir: ./temp

# Logging Configuration
logging:
  level: info
  format: text
EOF
    echo "Default config created at config/config.yaml"
    echo "Please update the configuration with your AWS credentials and S3 bucket information."
fi

# Create a simple sample.sfm file for testing
if [ ! -f data/sample.sfm ]; then
    echo "Creating sample SFM file..."
    cat > data/sample.sfm << EOF
# id,name,value,timestamp
jsonS3Exported:false
1,item1,100,2023-01-01T12:00:00Z
2,item2,200,2023-01-02T12:00:00Z
3,item3,300,2023-01-03T12:00:00Z
EOF
    echo "Sample SFM file created at data/sample.sfm"
fi

# Build the application
echo "Building S3 Exporter..."
go build -o s3-exporter main.go

echo "S3 Exporter setup complete."
echo "Run './s3-exporter' to start the application."
echo "Use './s3-exporter -config path/to/config.yaml' to use a custom config file."