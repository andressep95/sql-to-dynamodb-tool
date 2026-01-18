package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	resp, err := handler(context.Background(), events.APIGatewayProxyRequest{})
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	fmt.Println("Response body:")
	fmt.Println(resp.Body)

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
