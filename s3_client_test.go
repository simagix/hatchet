/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * s3_client_test.go
 */

package hatchet

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

const (
	S3Profile     = "default"
	testEndpoint  = "http://localhost:9090"
	testRegion    = "us-east-1"
	testAccessKey = "test-access-key"
	testSecretKey = "test-secret-key"
	testBucket    = "test-bucket"
	testKey       = "test-key"
)

func TestAWS(t *testing.T) {
	_, err := NewS3Client(S3Profile)
	if err != nil {
		t.Fatalf("failed to create S3 client: %v", err)
	}
}

func TestNewS3Client(t *testing.T) {
	s3client, err := NewS3Client(S3Profile, testEndpoint)
	if err != nil {
		t.Fatalf("failed to create S3 client: %v", err)
	}

	// create a new S3 bucket
	bucketName := "test-bucket"
	s3client.DeleteBucket(bucketName) // just in case
	err = s3client.CreateBucket(bucketName)
	if err != nil {
		t.Fatalf("failed to create S3 bucket: %v", err)
	}

	// upload a file to S3
	fileName := "test-file.txt"
	filePath := "./testdata/" + fileName
	err = s3client.PutObject(bucketName, fileName, filePath)
	if err != nil {
		t.Fatalf("failed to upload file to S3: %v", err)
	}

	// download the file from S3
	var buf []byte
	buf, err = s3client.GetObject(fmt.Sprintf("%v/%v", bucketName, fileName))
	if err != nil {
		t.Fatalf("failed to download file from S3: %v", err)
	}

	// check that the contents of the downloaded file match the original file
	originalBytes, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read original file: %v", err)
	}
	if !bytes.Equal(buf, originalBytes) {
		t.Fatalf("file contents do not match: expected '%s', got '%s'", string(originalBytes), string(buf))
	}
	t.Log(string(buf))

	// delete the file from S3
	err = s3client.DeleteObject(bucketName, fileName)
	if err != nil {
		t.Fatalf("failed to delete file from S3: %v", err)
	}

	err = s3client.DeleteBucket(bucketName)
	if err != nil {
		t.Fatalf("failed to delete S3 bucket: %v", err)
	}
}
