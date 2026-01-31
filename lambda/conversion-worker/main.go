package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Printf("Received %d SQS message(s)", len(sqsEvent.Records))

	for _, record := range sqsEvent.Records {
		if err := processMessage(ctx, record); err != nil {
			log.Printf("ERROR processing message %s: %v", record.MessageId, err)
			return err
		}
	}

	return nil
}

func processMessage(ctx context.Context, record events.SQSMessage) error {
	var msg SQSMessageBody
	if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
		log.Printf("ERROR: Failed to parse SQS message body: %v", err)
		return err
	}

	log.Printf("[%s] Processing conversion (optimization: %s, tables: %d)",
		msg.ConversionID, msg.OptimizationType, msg.TablesExtracted)

	// Update DynamoDB status to PROCESSING
	if err := UpdateStatusToProcessing(ctx, msg.ConversionID); err != nil {
		log.Printf("ERROR: Failed to update status: %v", err)
		return err
	}

	// Invoke Bedrock for conversion
	result, err := InvokeConversion(ctx, msg.SQLContent, msg.OptimizationType)
	if err != nil {
		log.Printf("[%s] Bedrock conversion failed: %v", msg.ConversionID, err)
		if updateErr := UpdateStatusToFailed(ctx, msg.ConversionID, err.Error()); updateErr != nil {
			log.Printf("[%s] Failed to update status to FAILED: %v", msg.ConversionID, updateErr)
		}
		return nil // Don't retry — already marked as FAILED
	}

	// Store result in DynamoDB
	if err := UpdateStatusToCompleted(ctx, msg.ConversionID, result); err != nil {
		log.Printf("[%s] Failed to update status to COMPLETED: %v", msg.ConversionID, err)
		return err
	}

	log.Printf("[%s] Conversion completed successfully", msg.ConversionID)
	return nil
}

func main() {
	initDynamoClient()
	initBedrockClient()
	lambda.Start(handler)
}
