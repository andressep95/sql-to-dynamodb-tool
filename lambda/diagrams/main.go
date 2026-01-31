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

	// POST /api/v1/schemas -> validar SQL
	if method == "POST" && strings.HasSuffix(path, "/api/v1/schemas") {
		return handleValidateSQL(ctx, req)
	}

	// GET /api/v1/schemas -> health / info
	if method == "GET" && strings.HasSuffix(path, "/api/v1/schemas") {
		return jsonResponse(200, map[string]string{
			"service": "sql-to-nosql-parser",
			"status":  "healthy",
		})
	}

	return jsonResponse(404, ErrorResponse{
		Error:   "NOT_FOUND",
		Message: "Route not found",
	})
}

func handleValidateSQL(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 1. Parsear body
	var body ConvertRequest
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return jsonResponse(400, ErrorResponse{
			Error:   ErrInvalidJSON,
			Message: "Request body is not valid JSON",
		})
	}

	// 2. Validar sqlContent no vacio
	if strings.TrimSpace(body.SQLContent) == "" {
		return jsonResponse(400, ErrorResponse{
			Error:   ErrEmptySQLContent,
			Message: "Field sqlContent is required",
		})
	}

	// 3. Validar optimizationType si se envia
	if body.OptimizationType != "" && !validOptimizationTypes[body.OptimizationType] {
		return jsonResponse(400, ErrorResponse{
			Error:   ErrInvalidOptimizationType,
			Message: "Invalid optimization type. Valid values: read_heavy, write_heavy, balanced",
		})
	}

	if body.OptimizationType == "" {
		body.OptimizationType = "balanced"
	}

	// 4. Ejecutar validacion SQL
	result := ValidateSQL(body.SQLContent)

	if !result.IsValid {
		return jsonResponse(400, ErrorResponse{
			Error:   ErrInvalidSQLSyntax,
			Message: result.Errors[0].Message,
			Details: result.Errors,
		})
	}

	// 5. Schema valido -> crear registro PENDING en DynamoDB
	record, err := CreateConversionRecord(ctx, body.SQLContent, body.OptimizationType, len(result.Tables))
	if err != nil {
		log.Printf("ERROR: Failed to create DynamoDB record: %v", err)
		return jsonResponse(500, ErrorResponse{
			Error:   ErrInternalServerError,
			Message: "Failed to create conversion",
		})
	}

	// 6. Retornar 202 Accepted
	return jsonResponse(202, map[string]interface{}{
		"conversionId": record.ConversionID,
		"status":       record.Status,
		"createdAt":    record.CreatedAt,
		"expiresAt":    record.ExpiresAt,
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
