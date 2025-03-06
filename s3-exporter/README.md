# S3 Exporter

A Go application for exporting segment data to JSON files and uploading them to Amazon S3.

## Overview

S3 Exporter is designed to process segment data in .sfm files, convert them to JSON format, and upload them to an S3 bucket. The process is optimized for reliability and performance, with features like:

- Tracking of exported segments to avoid duplicate processing
- Batch processing for efficient memory usage
- Compression to reduce storage and bandwidth costs
- Concurrent uploads for improved performance

## Requirements

- Go 1.18 or higher
- AWS account with S3 access
- AWS credentials with permissions to write to the destination bucket

## Directory Structure

```
S3-EXPORTER/
├── config/               # Configuration files
│   └── config.yaml       # Main configuration
├── data/                 # Data files
│   └── sample.sfm        # Sample segment file
├── exporter/             # Exporter logic
│   ├── config.go         # Configuration handling
│   ├── export.go         # Export functionality
│   └── utils.go          # Utility functions
├── logs/                 # Log files
│   └── app.log           # Application logs
├── scripts/              # Utility scripts
│   └── setup.sh          # Setup script
├── src/                  # Core functionality
│   ├── compression.go    # Compression utilities
│   └── s3_upload.go      # S3 upload functionality
├── tests/                # Tests
│   ├── exporter_tests.go # Exporter tests
│   └── s3_upload_test.go # S3 upload tests
├── .gitignore            # Git ignore file
├── go.mod                # Go module file
├── go.sum                # Go dependencies checksum
└── main.go               # Main application entry point
```

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/s3-exporter.git
   cd s3-exporter
   ```

2. Run the setup script:
   ```
   ./scripts/setup.sh
   ```

3. Update the config file with your AWS credentials and S3 bucket information:
   ```
   vim config/config.yaml
   ```

## Configuration

The `config.yaml` file contains all the configuration options:

```yaml
# S3 Configuration
s3:
  region: us-east-1
  bucket: your-s3-bucket
  access_key: YOUR_ACCESS_KEY
  secret_key: YOUR_SECRET_KEY

# Export Configuration
export:
  batch_size: 1000  # Number of JSON lines per file
  compression: true # Whether to compress files before upload
  temp_dir: ./temp

# Logging Configuration
logging:
  level: info
  format: text
```

## Usage

Run the application:

```
./s3-exporter
```

Options:

```
  -config string
        Path to configuration file (default "config/config.yaml")
  -data string
        Directory containing SFM files (default "data")
  -log string
        Path to log file (default "logs/app.log")
```

## Process Description

1. The application scans for `.sfm` files in the specified data directory.
2. For each file, it checks if it has already been exported (by looking for a `jsonS3Exported:true` flag).
3. If not exported, it reads the file and converts each record to JSON format.
4. The JSON records are batched into files based on the configured batch size.
5. Each batch file is compressed (if configured) and uploaded to S3.
6. After successful upload, the original `.sfm` file is marked as exported by setting the flag to `true`.

## Testing

Run tests:

```
go test ./tests/...
```

## License

[MIT License](LICENSE)