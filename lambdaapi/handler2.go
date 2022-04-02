package lambdaapi

import (

	"github.com/kennykarnama/video-color-palette-generator/source"
	"github.com/kennykarnama/video-color-palette-generator/processor"
	"github.com/kennykarnama/video-color-palette-generator/destination"

	"fmt"
	"context"
	"time"
	"os"
	"log"
)

func ColorPaletteHandler(paletteGenReq ColorPaletteGenerationRequest) (GenericResponse, error) {
	sourceProvider, err := source.GetProvider(paletteGenReq.SourceURL)
	if err != nil {
		return GenericResponse{}, err
	}
	// parse
	localURI, err := sourceProvider.LocalURI(context.Background(), paletteGenReq.SourceURL)
	if err != nil {
		return GenericResponse{}, err
	}
	csvOut := fmt.Sprintf("/tmp/%v.csv", time.Now().UTC().Unix())
	param := processor.Parameter{}
	param.InputFile = localURI
	param.InputSerial = paletteGenReq.SourceSerial
	param.PeriodDuration = paletteGenReq.PeriodSeconds
	param.PaletteSize = paletteGenReq.PaletteSize
	param.FunctionType = paletteGenReq.FunctionType
	param.CsvResult = csvOut

	defer func() {
		for _, f := range []string{csvOut, localURI} {
			log.Printf("Remove file: %v", f)
			os.Remove(f)
		}
	}()

	err = processor.Run(param)
	if err != nil {
		return GenericResponse{}, err
	}

	// upload to destination source
	destinationHandler, err := destination.GetTarget(paletteGenReq.DestinationURI)
	if err != nil {
		return GenericResponse{}, err
	}
	csvFile, err := os.Open(csvOut)
	if err != nil {
		return GenericResponse{}, err
	}
	defer csvFile.Close()

	err = destinationHandler.Upload(context.Background(), csvFile) 
	if err != nil {
		return GenericResponse{}, err
	}
	return GenericResponse{}, nil
}
