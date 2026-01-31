package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Printf("DLQ Handler: received %d message(s)", len(sqsEvent.Records))

	for _, record := range sqsEvent.Records {
		var msg SQSMessageBody
		if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
			log.Printf("ERROR: Failed to parse SQS message body: %v", err)
			return err
		}

		log.Printf("[%s] Marking conversion as FAILED (max retries exceeded)", msg.ConversionID)

		if err := UpdateStatusToFailed(ctx, msg.ConversionID, "Max retries exceeded"); err != nil {
			log.Printf("[%s] ERROR: Failed to update status to FAILED: %v", msg.ConversionID, err)
			return err
		}

		log.Printf("[%s] Successfully marked as FAILED", msg.ConversionID)
	}

	return nil
}

func main() {
	initDynamoClient()
	lambda.Start(handler)
}
