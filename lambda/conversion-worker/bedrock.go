package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

var bedrockClient *bedrockruntime.Client

func initBedrockClient() {
	if os.Getenv("USE_MOCK_BEDROCK") == "true" {
		log.Println("Mock Bedrock enabled — skipping client initialization")
		return
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("WARN: Failed to load AWS config for Bedrock: %v", err)
		return
	}

	endpoint := os.Getenv("BEDROCK_ENDPOINT")
	if endpoint != "" {
		bedrockClient = bedrockruntime.NewFromConfig(cfg, func(o *bedrockruntime.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	} else {
		bedrockClient = bedrockruntime.NewFromConfig(cfg)
	}
}

// InvokeConversion calls Bedrock (or returns mock) to convert SQL schema to DynamoDB JSON.
func InvokeConversion(ctx context.Context, sqlContent, optimizationType string) (string, error) {
	if os.Getenv("USE_MOCK_BEDROCK") == "true" {
		return mockBedrockResponse(), nil
	}

	if bedrockClient == nil {
		return "", fmt.Errorf("Bedrock client not initialized")
	}

	modelID := os.Getenv("BEDROCK_MODEL_ID")
	if modelID == "" {
		return "", fmt.Errorf("BEDROCK_MODEL_ID not set")
	}

	prompt := fmt.Sprintf(`Analiza el siguiente esquema SQL y conviértelo a un diseño óptimo de DynamoDB.

Tipo de optimización: %s

SQL Schema:
%s

Responde ÚNICAMENTE con un JSON válido con esta estructura:
{
  "tables": [
    {
      "tableName": "...",
      "partitionKey": {"name": "...", "type": "S|N|B"},
      "sortKey": {"name": "...", "type": "S|N|B"} | null,
      "attributes": [{"name": "...", "type": "S|N|B"}],
      "globalSecondaryIndexes": [],
      "billingMode": "PAY_PER_REQUEST"
    }
  ]
}`, optimizationType, sqlContent)

	requestBody, err := json.Marshal(map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	output, err := bedrockClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Body:        requestBody,
	})
	if err != nil {
		return "", fmt.Errorf("Bedrock InvokeModel failed: %w", err)
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return "", fmt.Errorf("failed to parse Bedrock response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from Bedrock")
	}

	return response.Content[0].Text, nil
}

func mockBedrockResponse() string {
	mock := map[string]interface{}{
		"tables": []map[string]interface{}{
			{
				"tableName":    "users",
				"partitionKey": map[string]string{"name": "userId", "type": "S"},
				"sortKey":      nil,
				"attributes": []map[string]string{
					{"name": "userId", "type": "S"},
					{"name": "email", "type": "S"},
					{"name": "username", "type": "S"},
				},
				"globalSecondaryIndexes": []map[string]interface{}{
					{
						"indexName":    "email-index",
						"partitionKey": map[string]string{"name": "email", "type": "S"},
						"sortKey":      nil,
						"projection":   "ALL",
					},
				},
				"billingMode": "PAY_PER_REQUEST",
			},
		},
	}
	b, _ := json.Marshal(mock)
	return string(b)
}
