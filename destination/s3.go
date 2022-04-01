package destination


import (
	sharedS3Internal "github.com/kennykarnama/video-color-palette-generator/shared/s3"
	
	"fmt"
	"io"
)

type S3Destination struct {
	Data *sharedS3Internal.S3URI
}

func NewS3DestinationFromURI(destinationURI string) (*S3Destination, error) {
	data, err := sharedS3Internal.ParseURL(uri)
	if err != nil {
		return nil fmt.Errorf("action=newS3SourceFromURI uri=%v err=%v", uri, err)
	}
	return &S3Destination {
		Data: data,
	}, nil
}

func (s *S3Destination) Upload(ctx context.Context, data io.Reader) error {
	return nil
}
