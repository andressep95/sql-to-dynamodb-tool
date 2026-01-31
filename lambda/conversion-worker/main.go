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

	// TODO: Invoke Bedrock for actual conversion
	log.Printf("[%s] Bedrock conversion placeholder - not implemented yet", msg.ConversionID)

	return nil
}

func main() {
	initDynamoClient()
	lambda.Start(handler)
}
