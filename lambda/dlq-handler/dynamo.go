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

// UpdateStatusToFailed sets status to FAILED and stores the error message.
func UpdateStatusToFailed(ctx context.Context, conversionID, errorMsg string) error {
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
			":status": &types.AttributeValueMemberS{Value: "FAILED"},
			":errMsg": &types.AttributeValueMemberS{Value: errorMsg},
		},
	})
	if err != nil {
		return fmt.Errorf("DynamoDB UpdateItem failed: %w", err)
	}

	log.Printf("[%s] Status updated to FAILED: %s", conversionID, errorMsg)
	return nil
}
