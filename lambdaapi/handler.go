package lambdaapi

import (
 	"github.com/aws/aws-lambda-go/events"

	"github.com/kennykarnama/video-color-palette-generator/source"
	"github.com/kennykarnama/video-color-palette-generator/processor"
	"github.com/kennykarnama/video-color-palette-generator/destination"

	"encoding/json"
	"net/http"
	"fmt"
	"context"
	"time"
	"os"
	"log"
)

func Handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
		case "GET":
			return GetHandler(req)
		case "POST":
			return PostHandler(req)
		default:
			return apiResponse(http.StatusInternalServerError, ErrorResponse{
				ErrorMessage: fmt.Errorf("unsupported HTTP method").Error(),
			})
	}
}

func GetHandler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusOK, "OK")
}

func PostHandler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var paletteGenReq ColorPaletteGenerationRequest
	if err := json.Unmarshal([]byte(req.Body), &paletteGenReq); err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
	}

	sourceProvider, err := source.GetProvider(paletteGenReq.SourceURL)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
	}
	// parse
	localURI, err := sourceProvider.LocalURI(context.Background(), paletteGenReq.SourceURL)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
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
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
	}
	// upload to destination source
	destinationHandler, err := destination.GetTarget(paletteGenReq.DestinationURI)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
	}
	csvFile, err := os.Open(csvOut)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
	}
	defer csvFile.Close()

	err = destinationHandler.Upload(context.Background(), csvFile) 
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorResponse{
			ErrorMessage: err.Error(),
		})
	}
	return apiResponse(http.StatusOK, struct{}{})
}

func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
      resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
      resp.StatusCode = status

      stringBody, _ := json.Marshal(body)
      resp.Body = string(stringBody)
      return &resp, nil
}
