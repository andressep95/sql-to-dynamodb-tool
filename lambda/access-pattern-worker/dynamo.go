package main

import (
	"context"
	"encoding/json"
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

// UpdateStatusToProcessingPatterns sets the status to PROCESSING_PATTERNS
func UpdateStatusToProcessingPatterns(ctx context.Context, conversionID string) error {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return fmt.Errorf("DYNAMODB_TABLE_NAME not set")
	}

	if dynamoClient == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"conversionId": &types.AttributeValueMemberS{Value: conversionID},
		},
		UpdateExpression: aws.String("SET #s = :status"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "PROCESSING_PATTERNS"},
		},
	})
	if err != nil {
		return fmt.Errorf("DynamoDB UpdateItem failed: %w", err)
	}

	log.Printf("[%s] Status updated to PROCESSING_PATTERNS", conversionID)
	return nil
}

// UpdateStatusToCompleted sets status to COMPLETED and stores the final design with access patterns
func UpdateStatusToCompleted(ctx context.Context, conversionID string, design *DynamoDBDesign) error {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return fmt.Errorf("DYNAMODB_TABLE_NAME not set")
	}
	if dynamoClient == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	// Serialize design to JSON
	designJSON, err := json.Marshal(design)
	if err != nil {
		return fmt.Errorf("failed to marshal design: %w", err)
	}

	_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"conversionId": &types.AttributeValueMemberS{Value: conversionID},
		},
		UpdateExpression: aws.String("SET #s = :status, #r = :result"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
			"#r": "noSqlSchema",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "COMPLETED"},
			":result": &types.AttributeValueMemberS{Value: string(designJSON)},
		},
	})
	if err != nil {
		return fmt.Errorf("DynamoDB UpdateItem failed: %w", err)
	}

	log.Printf("[%s] Status updated to COMPLETED with access patterns", conversionID)
	return nil
}

// UpdateStatusToPatternsFailed sets status to PATTERNS_FAILED and stores the error message
func UpdateStatusToPatternsFailed(ctx context.Context, conversionID, errorMsg string) error {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return fmt.Errorf("DYNAMODB_TABLE_NAME not set")
	}
	if dynamoClient == nil {
		return fmt.Errorf("DynamoDB client not initialized")
	}

	_, err := dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"conversionId": &types.AttributeValueMemberS{Value: conversionID},
		},
		UpdateExpression: aws.String("SET #s = :status, #e = :errMsg"),
		ExpressionAttributeNames: map[string]string{
			"#s": "status",
			"#e": "errorMessage",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: "PATTERNS_FAILED"},
			":errMsg": &types.AttributeValueMemberS{Value: errorMsg},
		},
	})
	if err != nil {
		return fmt.Errorf("DynamoDB UpdateItem failed: %w", err)
	}

	log.Printf("[%s] Status updated to PATTERNS_FAILED: %s", conversionID, errorMsg)
	return nil
}