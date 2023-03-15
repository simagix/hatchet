/*
 * Copyright 2022-present Kuei-chun Chen. All rights reserved.
 * s3_client.go
 */

package hatchet

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Client provides methods to interact with an S3 service.
type S3Client struct {
	service *s3.S3
}

func NewS3Client(profile string, params...string) (*S3Client, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	})
	if err != nil {
		return nil, err
	}

	config := aws.Config{
		Region:           aws.String(*sess.Config.Region),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      sess.Config.Credentials,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
	if len(params) > 0 && params[0] != "" {
		config.Endpoint = &params[0]
	}
	if sess, err = session.NewSession(&config); err != nil {
		return nil, errors.New("error creating session")
	}

	// Create a new S3 service client
	service := s3.New(sess)
	return &S3Client{service}, nil
}

// CreateBucket creates a new S3 bucket.
func (c *S3Client) CreateBucket(bucket string) error {
	_, err := c.service.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// CreateBucket creates a new S3 bucket.
func (c *S3Client) DeleteBucket(bucket string) error {
	_, err := c.service.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// DeleteObject deletes an object from S3.
func (c *S3Client) DeleteObject(bucket, key string) error {
	_, err := c.service.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// PutObject uploads a file to S3.
func (c *S3Client) PutObject(bucket, key, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	_, err = c.service.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})

	return err
}

// GetObject retrieves an object from S3 and returns its contents as a byte slice.
func (c *S3Client) GetObject(bucket, key string) ([]byte, error) {
	resp, err := c.service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error retrieving S3 object: %v", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
