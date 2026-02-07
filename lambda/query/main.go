package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (V2Response, error) {
	method := req.RequestContext.HTTP.Method
	path := req.RequestContext.HTTP.Path
	log.Printf("Request: %s %s", method, path)

	// Extract tenantId from JWT claims
	tenantID := "public"
	if jwtClaims := req.RequestContext.Authorizer.JWT; jwtClaims != nil {
		if tid, ok := jwtClaims.Claims["custom:tenant_id"]; ok && tid != "" {
			tenantID = tid
		}
	}

	// GET /api/v1/schemas/{id} -> obtener por ID
	if method == "GET" && req.PathParameters["id"] != "" {
		return handleGetByID(ctx, req.PathParameters["id"], tenantID)
	}

	// GET /api/v1/schemas -> listar todos
	if method == "GET" && strings.HasSuffix(path, "/api/v1/schemas") {
		return handleListAll(ctx, tenantID)
	}

	return jsonResponse(404, map[string]string{
		"error":   "NOT_FOUND",
		"message": "Route not found",
	})
}

func handleGetByID(ctx context.Context, id string, tenantID string) (V2Response, error) {
	record, err := GetConversionByID(ctx, id, tenantID)
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

	// Parse noSqlSchema from string to JSON object
	parseNoSqlSchema(record)

	return jsonResponse(200, record)
}

func handleListAll(ctx context.Context, tenantID string) (V2Response, error) {
	records, err := ListConversions(ctx, tenantID)
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

func jsonResponse(statusCode int, body interface{}) (V2Response, error) {
	b, _ := json.Marshal(body)
	return V2Response{
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

// parseNoSqlSchema converts the noSqlSchema string to a JSON object
func parseNoSqlSchema(record map[string]interface{}) {
	if noSqlSchemaStr, ok := record["noSqlSchema"].(string); ok && noSqlSchemaStr != "" {
		var noSqlSchema interface{}
		if err := json.Unmarshal([]byte(noSqlSchemaStr), &noSqlSchema); err == nil {
			record["noSqlSchema"] = noSqlSchema
		} else {
			log.Printf("WARN: Failed to parse noSqlSchema JSON: %v", err)
		}
	}
}
