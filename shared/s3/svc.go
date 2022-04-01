package s3

import (
	"context"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	s3cli "github.com/aws/aws-sdk-go/service/s3"
)

var (
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	svc        *s3cli.S3
)

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	}))
	uploader = s3manager.NewUploader(sess)
	downloader = s3manager.NewDownloader(sess)
	svc = s3cli.New(sess)
}

func PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         body,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetWithContext(ctx context.Context, bucket, key string) (io.Reader, error) {
	input := &s3cli.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	result, err := svc.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func Download(ctx context.Context, bucket, key string, target string) error {
	file, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	_, err = downloader.DownloadWithContext(
		ctx,
		file,
		&s3cli.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
