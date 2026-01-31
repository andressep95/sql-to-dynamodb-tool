package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Request: %s %s", req.HTTPMethod, req.Path)

	path := req.Path
	method := req.HTTPMethod

	// GET /api/v1/schemas/{id} -> obtener por ID
	if method == "GET" && req.PathParameters["id"] != "" {
		return handleGetByID(ctx, req.PathParameters["id"])
	}

	// GET /api/v1/schemas -> listar todos
	if method == "GET" && strings.HasSuffix(path, "/api/v1/schemas") {
		return handleListAll(ctx)
	}

	return jsonResponse(404, map[string]string{
		"error":   "NOT_FOUND",
		"message": "Route not found",
	})
}

func handleGetByID(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	record, err := GetConversionByID(ctx, id)
	if err != nil {
		log.Printf("ERROR: GetConversionByID failed: %v", err)
		return jsonResponse(500, map[string]string{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Failed to retrieve conversion",
		})
	}

	if record == nil {
		return jsonResponse(404, map[string]string{
			"error":   "NOT_FOUND",
			"message": "Conversion not found",
		})
	}

	return jsonResponse(200, record)
}

func handleListAll(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	records, err := ListConversions(ctx)
	if err != nil {
		log.Printf("ERROR: ListConversions failed: %v", err)
		return jsonResponse(500, map[string]string{
			"error":   "INTERNAL_SERVER_ERROR",
			"message": "Failed to list conversions",
		})
	}

	return jsonResponse(200, map[string]interface{}{
		"conversions": records,
		"count":       len(records),
	})
}

func jsonResponse(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(b),
	}, nil
}

func main() {
	initDynamoClient()
	lambda.Start(handler)
}
