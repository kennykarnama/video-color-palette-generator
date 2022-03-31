package lambdaapi

import (
 	"github.com/aws/aws-lambda-go/events"
	
	"encoding/json"
	"net/http"
	"fmt"
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

func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
      resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
      resp.StatusCode = status

      stringBody, _ := json.Marshal(body)
      resp.Body = string(stringBody)
      return &resp, nil
}
