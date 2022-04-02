package destination


import (
	sharedS3Internal "github.com/kennykarnama/video-color-palette-generator/shared/s3"
	
	"fmt"
	"io"
	"context"
)

type S3Destination struct {
	Data *sharedS3Internal.S3URI
}

func NewS3DestinationFromURI(destinationURI string) (*S3Destination, error) {
	data, err := sharedS3Internal.ParseURL(destinationURI)
	if err != nil {
		return nil, fmt.Errorf("action=newS3SourceFromURI uri=%v err=%v", destinationURI, err)
	}
	return &S3Destination {
		Data: data,
	}, nil
}

func (s *S3Destination) Upload(ctx context.Context, data io.Reader) error {
	err  := sharedS3Internal.PutObject(ctx, s.Data.Bucket, s.Data.Key, data)
	if err != nil {
		return fmt.Errorf("s3.destination target_bucket=%v key=%v err=%v", s.Data.Bucket, s.Data.Key, err)
	}
	return nil

}

