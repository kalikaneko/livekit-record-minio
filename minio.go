package livekitminio

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func GetMinIOObject(objectName string) (io.Reader, error) {
	// Get MinIO server details
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_KEY")
	secretKey := os.Getenv("S3_SECRET")
	useSSL := false // Set to true if your MinIO server uses SSL
	bucketName := os.Getenv("S3_BUCKET")

	// Initialize a MinIO client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}

	// Fetch the object from MinIO and copy its content to the buffer
	object, err := minioClient.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(buffer, object)
	reader := bytes.NewReader(buffer.Bytes())
	return reader, nil
}
