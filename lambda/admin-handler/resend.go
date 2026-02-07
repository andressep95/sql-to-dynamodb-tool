package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

type ResendEmailResponse struct {
	ID string `json:"id"`
}

var (
	secretsClient *secretsmanager.Client
	secretsOnce   sync.Once
	cachedAPIKey  string
	cacheMutex    sync.RWMutex
)

func initSecretsManager() {
	secretsOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Printf("Failed to load AWS config for Secrets Manager: %v", err)
			return
		}
		secretsClient = secretsmanager.NewFromConfig(cfg)
	})
}

func getResendAPIKey(ctx context.Context) (string, error) {
	// Check cache first
	cacheMutex.RLock()
	if cachedAPIKey != "" {
		cacheMutex.RUnlock()
		return cachedAPIKey, nil
	}
	cacheMutex.RUnlock()

	// Get from Secrets Manager
	initSecretsManager()
	if secretsClient == nil {
		return "", fmt.Errorf("secrets manager client not initialized")
	}

	secretName := os.Getenv("RESEND_SECRET_NAME")
	if secretName == "" {
		return "", fmt.Errorf("RESEND_SECRET_NAME not configured")
	}

	result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	if result.SecretString == nil {
		return "", fmt.Errorf("secret value is empty")
	}

	// Cache the API key
	cacheMutex.Lock()
	cachedAPIKey = *result.SecretString
	cacheMutex.Unlock()

	return cachedAPIKey, nil
}

func sendInvitationEmail(email, invitationCode, tenantID string) error {
	ctx := context.Background()
	
	apiKey, err := getResendAPIKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Resend API key: %w", err)
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "https://app-sql.cloudcentinel.com"
	}

	registerLink := fmt.Sprintf("%s/register?code=%s", appURL, invitationCode)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 10px; padding: 30px; margin-bottom: 20px;">
        <h1 style="color: #2c3e50; margin-bottom: 20px;">Invitación a SQL to NoSQL Parser</h1>
        <p style="font-size: 16px; margin-bottom: 20px;">Has sido invitado a unirte a la plataforma.</p>
        
        <div style="background-color: #fff; border-left: 4px solid #3498db; padding: 15px; margin: 20px 0;">
            <p style="margin: 0; font-weight: bold;">Código de invitación:</p>
            <p style="margin: 5px 0; font-size: 24px; color: #3498db; font-family: monospace;">%s</p>
        </div>
        
        <p style="margin: 20px 0;">Completa tu registro haciendo click en el siguiente enlace:</p>
        
        <a href="%s" style="display: inline-block; background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; font-weight: bold; margin: 10px 0;">Completar Registro</a>
        
        <p style="margin-top: 30px; font-size: 14px; color: #7f8c8d;">
            Este enlace expira en 7 días.<br>
            Si no solicitaste esta invitación, puedes ignorar este correo.
        </p>
    </div>
    
    <div style="text-align: center; color: #95a5a6; font-size: 12px;">
        <p>SQL to NoSQL Parser - Powered by AWS</p>
    </div>
</body>
</html>
`, invitationCode, registerLink)

	emailReq := ResendEmailRequest{
		From:    "SQL to NoSQL Parser <onboarding@resend.dev>",
		To:      []string{email},
		Subject: "Invitación a SQL to NoSQL Parser",
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(emailReq)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return fmt.Errorf("resend API error (status %d): %v", resp.StatusCode, errorBody)
	}

	return nil
}
