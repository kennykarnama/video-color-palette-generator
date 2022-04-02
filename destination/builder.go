package destination

import (
	"fmt"
)

func GetTarget(destinationURI string) (Target, error) {
	if s3Dest, err := NewS3DestinationFromURI(destinationURI); err == nil {
		return s3Dest, nil
	}
	if destinationURI == "" {
		return NewNoop(""), nil
	}
	return nil, fmt.Errorf("GetTarget destURI=%v err=%v", destinationURI, "unknown target")
}
