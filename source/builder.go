package source

import (
	"log"
	"fmt"
)

func GetProvider(sourceURI string) (Provider, error) {
	if s3Source, err := NewS3SourceFromURI(sourceURI); err == nil {
		log.Printf("GetProvider.s3 err=%v", err)
		return s3Source, nil
	}
	return nil, fmt.Errorf("GetProvider sourceURI=%v err=undefined sourceURI", sourceURI)
}
