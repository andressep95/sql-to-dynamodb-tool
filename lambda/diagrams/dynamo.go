package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
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

// ConversionRecord represents a record to be stored in DynamoDB
type ConversionRecord struct {
	ConversionID   string `json:"conversionId"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt"`
	ExpiresAt      int64  `json:"expiresAt"`
	ConversionDate string `json:"conversionDate"`
	SQLContent     string `json:"sqlContent"`
	OptimizationType string `json:"optimizationType"`
	TablesExtracted  int    `json:"tablesExtracted"`
}

// CreateConversionRecord generates a UUID, builds the record, and stores it in DynamoDB.
// Returns the record on success or an error.
func CreateConversionRecord(ctx context.Context, sqlContent, optimizationType string, tablesExtracted int) (*ConversionRecord, error) {
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE_NAME not set")
	}

	if dynamoClient == nil {
		return nil, fmt.Errorf("DynamoDB client not initialized")
	}

	now := time.Now().UTC()
	record := &ConversionRecord{
		ConversionID:     uuid.New().String(),
		Status:           "PENDING",
		CreatedAt:        now.Format(time.RFC3339),
		ExpiresAt:        now.Add(24 * time.Hour).Unix(),
		ConversionDate:   now.Format("2006-01-02"),
		SQLContent:       sqlContent,
		OptimizationType: optimizationType,
		TablesExtracted:  tablesExtracted,
	}

	item := map[string]types.AttributeValue{
		"conversionId":     &types.AttributeValueMemberS{Value: record.ConversionID},
		"status":           &types.AttributeValueMemberS{Value: record.Status},
		"createdAt":        &types.AttributeValueMemberS{Value: record.CreatedAt},
		"expiresAt":        &types.AttributeValueMemberN{Value: strconv.FormatInt(record.ExpiresAt, 10)},
		"conversionDate":   &types.AttributeValueMemberS{Value: record.ConversionDate},
		"sqlContent":       &types.AttributeValueMemberS{Value: record.SQLContent},
		"optimizationType": &types.AttributeValueMemberS{Value: record.OptimizationType},
		"tablesExtracted":  &types.AttributeValueMemberN{Value: strconv.Itoa(record.TablesExtracted)},
	}

	_, err := dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("DynamoDB PutItem failed: %w", err)
	}

	log.Printf("[%s] DynamoDB record created (status: PENDING)", record.ConversionID)
	return record, nil
}
