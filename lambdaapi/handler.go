package lambdaapi

import (
 	"github.com/aws/aws-lambda-go/events"

	"github.com/kennykarnama/video-color-palette-generator/source"
	"github.com/kennykarnama/video-color-palette-generator/processor"

	"encoding/json"
	"net/http"
	"fmt"
	"context"
)

func Handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch req.HTTPMethod {
		case "GET":
			return GetHandler(req)
		default:
			return nil, fmt.Errorf("unsupported http method")
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
	param := processor.Parameter{}
	param.InputFile = localURI
	param.InputSerial = paletteGenReq.SourceSerial
	param.PeriodDuration = paletteGenReq.PeriodSeconds
	param.PaletteSize = paletteGenReq.PaletteSize
	param.FunctionType = paletteGenReq.FunctionType

	err = processor.Run(param)
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
