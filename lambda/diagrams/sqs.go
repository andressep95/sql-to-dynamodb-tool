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

// SQSMessage is the message body sent to the conversion queue.
type SQSMessage struct {
	ConversionID     string `json:"conversionId"`
	SQLContent       string `json:"sqlContent"`
	OptimizationType string `json:"optimizationType"`
	TablesExtracted  int    `json:"tablesExtracted"`
}

// SendToQueue sends a conversion record to the SQS queue for async processing.
func SendToQueue(ctx context.Context, record *ConversionRecord) error {
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		return fmt.Errorf("SQS_QUEUE_URL not set")
	}

	if sqsClient == nil {
		return fmt.Errorf("SQS client not initialized")
	}

	msg := SQSMessage{
		ConversionID:     record.ConversionID,
		SQLContent:       record.SQLContent,
		OptimizationType: record.OptimizationType,
		TablesExtracted:  record.TablesExtracted,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal SQS message: %w", err)
	}

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("SQS SendMessage failed: %w", err)
	}

	log.Printf("[%s] Message sent to SQS queue", record.ConversionID)
	return nil
}
