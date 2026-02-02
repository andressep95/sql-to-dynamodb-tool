package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Printf("Received %d SQS message(s) for access pattern generation", len(sqsEvent.Records))

	for _, record := range sqsEvent.Records {
		if err := processAccessPatternMessage(ctx, record); err != nil {
			log.Printf("ERROR processing access pattern message %s: %v", record.MessageId, err)
			return err
		}
	}

	return nil
}

func processAccessPatternMessage(ctx context.Context, record events.SQSMessage) error {
	var msg AccessPatternMessage
	if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
		log.Printf("ERROR: Failed to parse access pattern SQS message: %v", err)
		return err
	}

	log.Printf("[%s] Processing access patterns for DynamoDB design", msg.ConversionID)

	// Update status to PROCESSING_PATTERNS
	if err := UpdateStatusToProcessingPatterns(ctx, msg.ConversionID); err != nil {
		log.Printf("ERROR: Failed to update status: %v", err)
		return err
	}

	// Generate access patterns using Bedrock
	accessPatterns, err := GenerateAccessPatterns(ctx, msg.SQLContent, msg.DynamoDBDesign, msg.OptimizationType)
	if err != nil {
		log.Printf("[%s] Access pattern generation failed: %v", msg.ConversionID, err)
		if updateErr := UpdateStatusToPatternsFailed(ctx, msg.ConversionID, err.Error()); updateErr != nil {
			log.Printf("[%s] Failed to update status to PATTERNS_FAILED: %v", msg.ConversionID, updateErr)
		}
		return nil // Don't retry — already marked as failed
	}

	// Merge access patterns with existing design
	finalDesign := mergeAccessPatterns(msg.DynamoDBDesign, accessPatterns)

	// Store final result in DynamoDB
	if err := UpdateStatusToCompleted(ctx, msg.ConversionID, finalDesign); err != nil {
		log.Printf("[%s] Failed to update status to COMPLETED: %v", msg.ConversionID, err)
		return err
	}

	log.Printf("[%s] Access pattern generation completed successfully", msg.ConversionID)
	return nil
}

func main() {
	initDynamoClient()
	initBedrockClient()
	lambda.Start(handler)
}