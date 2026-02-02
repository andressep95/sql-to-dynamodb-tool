package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var sqsClient *sqs.Client

func initSQSClient() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("WARN: Failed to load AWS config for SQS: %v", err)
		return
	}

	endpoint := os.Getenv("SQS_ENDPOINT")
	if endpoint != "" {
		sqsClient = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	} else {
		sqsClient = sqs.NewFromConfig(cfg)
	}
}

// AccessPatternMessage represents the message sent to access-pattern-worker
type AccessPatternMessage struct {
	ConversionID     string          `json:"conversionId"`
	SQLContent       string          `json:"sqlContent"`
	DynamoDBDesign   *DynamoDBDesign `json:"dynamodbDesign"`
	OptimizationType string          `json:"optimizationType"`
}

// SendToAccessPatternQueue sends a message to the access pattern processing queue
func SendToAccessPatternQueue(ctx context.Context, conversionID, sqlContent string, design *DynamoDBDesign, optimizationType string) error {
	queueURL := os.Getenv("ACCESS_PATTERN_QUEUE_URL")
	if queueURL == "" {
		return fmt.Errorf("ACCESS_PATTERN_QUEUE_URL not set")
	}

	if sqsClient == nil {
		return fmt.Errorf("SQS client not initialized")
	}

	message := AccessPatternMessage{
		ConversionID:     conversionID,
		SQLContent:       sqlContent,
		DynamoDBDesign:   design,
		OptimizationType: optimizationType,
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal access pattern message: %w", err)
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:     aws.String(queueURL),
		MessageBody:  aws.String(string(messageBody)),
		DelaySeconds: 15, // Avoid Bedrock throttling between conversion and access pattern generation
		MessageAttributes: map[string]types.MessageAttributeValue{
			"ConversionID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(conversionID),
			},
			"OptimizationType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(optimizationType),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send message to access pattern queue: %w", err)
	}

	log.Printf("[%s] Message sent to access pattern queue", conversionID)
	return nil
}
