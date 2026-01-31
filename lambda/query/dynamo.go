package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var dynamoClient *dynamodb.Client

func initDynamoClient() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("WARN: Failed to load AWS config: %v", err)
		return
	}

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint != "" {
		dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	} else {
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}
}

// GetConversionByID retrieves a single conversion record by its ID.
func GetConversionByID(ctx context.Context, conversionID string) (map[string]interface{}, error) {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME not set")
	}

	if dynamoClient == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"conversionId": &types.AttributeValueMemberS{Value: conversionID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDB GetItem failed: %w", err)
	}

	if result.Item == nil {
		return nil, nil
	}

	return unmarshalItem(result.Item), nil
}

// ListConversions retrieves all conversion records from DynamoDB via Scan.
func ListConversions(ctx context.Context) ([]map[string]interface{}, error) {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME not set")
	}

	if dynamoClient == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	result, err := dynamoClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDB Scan failed: %w", err)
	}

	records := make([]map[string]interface{}, 0, len(result.Items))
	for _, item := range result.Items {
		records = append(records, unmarshalItem(item))
	}

	log.Printf("Listed %d conversion records", len(records))
	return records, nil
}

// unmarshalItem converts a DynamoDB item to a simple map.
func unmarshalItem(item map[string]types.AttributeValue) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range item {
		switch val := v.(type) {
		case *types.AttributeValueMemberS:
			result[k] = val.Value
		case *types.AttributeValueMemberN:
			result[k] = val.Value
		case *types.AttributeValueMemberBOOL:
			result[k] = val.Value
		case *types.AttributeValueMemberNULL:
			result[k] = nil
		default:
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
