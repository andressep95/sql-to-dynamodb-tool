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
	design, err := InvokeConversion(ctx, msg.SQLContent, msg.OptimizationType)
	if err != nil {
		log.Printf("[%s] Bedrock conversion failed: %v", msg.ConversionID, err)
		if updateErr := UpdateStatusToFailed(ctx, msg.ConversionID, err.Error()); updateErr != nil {
			log.Printf("[%s] Failed to update status to FAILED: %v", msg.ConversionID, updateErr)
		}
		return nil // Don't retry — already marked as FAILED
	}

	// Validate design
	validation := ValidateDesignFull(design, msg.SQLContent)
	if !validation.IsValid {
		log.Printf("[%s] Design validation failed: %v", msg.ConversionID, validation.Errors)
		if updateErr := UpdateStatusToFailed(ctx, msg.ConversionID, "Design validation failed"); updateErr != nil {
			log.Printf("[%s] Failed to update status to FAILED: %v", msg.ConversionID, updateErr)
		}
		return nil
	}

	// Store intermediate result and trigger access pattern generation
	if err := UpdateStatusToDesignCompleted(ctx, msg.ConversionID, design); err != nil {
		log.Printf("[%s] Failed to update status to DESIGN_COMPLETED: %v", msg.ConversionID, err)
		return err
	}

	// Send message to access pattern queue
	if err := SendToAccessPatternQueue(ctx, msg.ConversionID, msg.SQLContent, design, msg.OptimizationType); err != nil {
		log.Printf("[%s] Failed to send to access pattern queue: %v", msg.ConversionID, err)
		// Don't fail the conversion, just log the error
		// The design is already saved, user can still see basic structure
	}

	log.Printf("[%s] Conversion completed successfully", msg.ConversionID)
	return nil
}

func main() {
	initDynamoClient()
	initBedrockClient()
	initSQSClient()
	lambda.Start(handler)
}
