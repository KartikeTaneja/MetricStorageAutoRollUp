package src

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// UploadToS3 uploads a file to an S3 bucket
func UploadToS3(filePath, s3Path, bucket, region, accessKey, secretKey string) error {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file stats
	// fileInfo, err := file.Stat()
	// if err != nil {
	// 	return fmt.Errorf("error getting file stats: %w", err)
	// }

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // 5MB part size
		u.Concurrency = 5            // 5 concurrent uploads
	})

	// Set content type based on file extension
	contentType := "application/octet-stream"
	ext := filepath.Ext(filePath)
	if ext == ".json" {
		contentType = "application/json"
	} else if ext == ".gz" {
		contentType = "application/gzip"
	}

	// Upload the file to S3
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(s3Path),
		Body:          file,
		// ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("error uploading file to S3: %w", err)
	}

	return nil
}

// DownloadFromS3 downloads a file from an S3 bucket
func DownloadFromS3(s3Path, localPath, bucket, region, accessKey, secretKey string) error {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	// Create a downloader
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the downloaded content
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Download the file from S3
	_, err = downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return fmt.Errorf("error downloading file from S3: %w", err)
	}

	return nil
}

// ListFilesInBucket lists files in an S3 bucket with a specified prefix
func ListFilesInBucket(bucket, prefix, region, accessKey, secretKey string) ([]string, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %w", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	// List objects in the bucket
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects in S3 bucket: %w", err)
	}

	// Extract the keys from the response
	var keys []string
	for _, item := range resp.Contents {
		keys = append(keys, *item.Key)
	}

	return keys, nil
}

// DeleteFileFromS3 deletes a file from an S3 bucket
func DeleteFileFromS3(s3Path, bucket, region, accessKey, secretKey string) error {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		return fmt.Errorf("error creating AWS session: %w", err)
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Delete the object
	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return fmt.Errorf("error deleting object from S3: %w", err)
	}

	// Wait until the deletion is complete
	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return fmt.Errorf("error waiting for object deletion: %w", err)
	}

	return nil
}