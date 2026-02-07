package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

var cognitoClient *cognitoidentityprovider.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	cognitoClient = cognitoidentityprovider.NewFromConfig(cfg)
}

func handler(ctx context.Context, event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	userPoolID := event.UserPoolID
	username := event.UserName
	tenantID := os.Getenv("DEFAULT_TENANT_ID")
	if tenantID == "" {
		tenantID = "public"
	}
	defaultRole := os.Getenv("DEFAULT_ROLE")
	if defaultRole == "" {
		defaultRole = "USER_TENANT"
	}

	log.Printf("Post-confirmation trigger for user %s in pool %s, assigning tenant: %s, role: %s", username, userPoolID, tenantID, defaultRole)

	_, err := cognitoClient.AdminUpdateUserAttributes(ctx, &cognitoidentityprovider.AdminUpdateUserAttributesInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(username),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("custom:tenant_id"),
				Value: aws.String(tenantID),
			},
			{
				Name:  aws.String("custom:role"),
				Value: aws.String(defaultRole),
			},
		},
	})
	if err != nil {
		log.Printf("ERROR: Failed to set attributes for user %s: %v", username, err)
		return event, err
	}

	log.Printf("Successfully set custom:tenant_id=%s and custom:role=%s for user %s", tenantID, defaultRole, username)
	return event, nil
}

func main() {
	lambda.Start(handler)
}
