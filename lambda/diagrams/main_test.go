package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func v2Request(method, path, body string) events.APIGatewayV2HTTPRequest {
	return events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: method,
				Path:   path,
			},
		},
		Body: body,
	}
}

func TestHandler_NotFound(t *testing.T) {
	resp, err := handler(context.Background(), v2Request("", "", ""))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404 for empty request, got %d", resp.StatusCode)
	}
}

func TestHandler_GET_Schemas(t *testing.T) {
	resp, err := handler(context.Background(), v2Request("GET", "/api/v1/schemas", ""))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandler_POST_ValidSQL(t *testing.T) {
	body, _ := json.Marshal(ConvertRequest{
		SQLContent: `CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email TEXT UNIQUE
		);`,
		OptimizationType: "balanced",
	})

	resp, err := handler(context.Background(), v2Request("POST", "/api/v1/schemas", string(body)))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, resp.Body)
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(resp.Body), &result)
	if result["isValid"] != true {
		t.Fatalf("expected isValid=true, got %v", result["isValid"])
	}
}

func TestHandler_POST_InvalidSQL(t *testing.T) {
	body, _ := json.Marshal(ConvertRequest{
		SQLContent: `SELECT * FROM users;`,
	})

	resp, err := handler(context.Background(), v2Request("POST", "/api/v1/schemas", string(body)))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	json.Unmarshal([]byte(resp.Body), &errResp)
	if errResp.Error != ErrInvalidSQLSyntax {
		t.Fatalf("expected error %s, got %s", ErrInvalidSQLSyntax, errResp.Error)
	}
}

func TestHandler_POST_EmptyBody(t *testing.T) {
	body, _ := json.Marshal(ConvertRequest{SQLContent: ""})

	resp, err := handler(context.Background(), v2Request("POST", "/api/v1/schemas", string(body)))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	json.Unmarshal([]byte(resp.Body), &errResp)
	if errResp.Error != ErrEmptySQLContent {
		t.Fatalf("expected error %s, got %s", ErrEmptySQLContent, errResp.Error)
	}
}

func TestHandler_POST_InvalidOptimizationType(t *testing.T) {
	body, _ := json.Marshal(ConvertRequest{
		SQLContent:       "CREATE TABLE t (id INT);",
		OptimizationType: "super_fast",
	})

	resp, err := handler(context.Background(), v2Request("POST", "/api/v1/schemas", string(body)))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	json.Unmarshal([]byte(resp.Body), &errResp)
	if errResp.Error != ErrInvalidOptimizationType {
		t.Fatalf("expected error %s, got %s", ErrInvalidOptimizationType, errResp.Error)
	}
}

func TestHandler_POST_InvalidJSON(t *testing.T) {
	resp, err := handler(context.Background(), v2Request("POST", "/api/v1/schemas", "not json"))
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
