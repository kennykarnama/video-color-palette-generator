package source

import (
	sharedS3Internal "github.com/kennykarnama/video-color-palette-generator/shared/s3"
	
	"fmt"
	"path/filepath"
	"context"
)

type S3Source struct {
	Data *sharedS3Internal.S3URI
}

func NewS3SourceFromURI(uri string) (*S3Source, error) {
	data, err := sharedS3Internal.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("action=newS3SourceFromURI uri=%v err=%v", uri, err)
	}
	return &S3Source {
		Data: data,
	}, nil
}

func (s *S3Source) LocalURI(ctx context.Context, uri string) (string, error) {
	localUri := filepath.Join("/tmp", s.Data.Key)
	err := sharedS3Internal.Download(ctx, s.Data.Bucket, s.Data.Key, localUri)
	if err != nil {
		return "", fmt.Errorf("action=s3Source.LocalURI uri=%v target=%v err=%v", uri, localUri, err)
	}
	return localUri, nil
}


