package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/research-data-analysis/config"
)

// UploadFile mengupload file ke Google Cloud Storage
func UploadFile(ctx context.Context, fileName string, data io.Reader, contentType string) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(config.GetGCSBucket())
	obj := bucket.Object(fileName)

	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType

	if _, err := io.Copy(wc, data); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}

	// Return public URL
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", config.GetGCSBucket(), fileName)
	return publicURL, nil
}

// GetSignedURL menghasilkan signed URL untuk akses sementara
func GetSignedURL(ctx context.Context, fileName string, expiration time.Duration) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := client.Bucket(config.GetGCSBucket()).SignedURL(fileName, opts)
	if err != nil {
		return "", fmt.Errorf("Bucket.SignedURL: %v", err)
	}

	return url, nil
}

// DeleteFile menghapus file dari GCS
func DeleteFile(ctx context.Context, fileName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	obj := client.Bucket(config.GetGCSBucket()).Object(fileName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("Object.Delete: %v", err)
	}

	return nil
}

// DownloadFile mengunduh file dari GCS
func DownloadFile(ctx context.Context, fileName string) ([]byte, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	rc, err := client.Bucket(config.GetGCSBucket()).Object(fileName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("Object.NewReader: %v", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %v", err)
	}

	return data, nil
}
