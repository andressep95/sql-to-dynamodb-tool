package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type CreateInvitationRequest struct {
	TenantID string `json:"tenantId"`
	Role     string `json:"role"`
	Email    string `json:"email,omitempty"` // Opcional: para enviar por correo
}

type InvitationResponse struct {
	Code      string `json:"code"`
	TenantID  string `json:"tenantId"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expiresAt"`
	CreatedBy string `json:"createdBy"`
}

type RegisterRequest struct {
	InvitationCode string `json:"invitationCode"`
	Email          string `json:"email"`
	Password       string `json:"password"`
}

type InvitationItem struct {
	InvitationCode string `dynamodbav:"invitationCode"`
	TenantID       string `dynamodbav:"tenantId"`
	Role           string `dynamodbav:"role"`
	CreatedBy      string `dynamodbav:"createdBy"`
	CreatedAt      int64  `dynamodbav:"createdAt"`
	ExpiresAt      int64  `dynamodbav:"expiresAt"`
	Used           bool   `dynamodbav:"used"`
}

var dynamoClient *dynamodb.Client

func initDynamoDB() {
	if dynamoClient == nil {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("Failed to load AWS config for DynamoDB: %v", err)
		}
		dynamoClient = dynamodb.NewFromConfig(cfg)
	}
}

// Genera código de 6 dígitos
func generateInvitationCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

func handleCreateInvitation(ctx context.Context, claims *UserClaims, body string) (V2Response, error) {
	// Solo SUPER_ADMIN y REALM_ADMIN pueden crear invitaciones
	if claims.Role != "SUPER_ADMIN" && claims.Role != "REALM_ADMIN" {
		return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Insufficient permissions"})
	}

	var req CreateInvitationRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, map[string]string{"error": "INVALID_REQUEST", "message": "Invalid JSON"})
	}

	// REALM_ADMIN solo puede crear invitaciones para su tenant
	if claims.Role == "REALM_ADMIN" && req.TenantID != claims.TenantID {
		return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Cannot create invitations for other tenants"})
	}

	// Validar rol
	validRoles := map[string]bool{
		"USER_TENANT":      true,
		"REALM_SUPERVISOR": true,
		"REALM_ADMIN":      claims.Role == "SUPER_ADMIN", // Solo SUPER_ADMIN puede invitar REALM_ADMIN
	}
	if !validRoles[req.Role] {
		return jsonResponse(400, map[string]string{"error": "INVALID_ROLE", "message": "Invalid role"})
	}

	// Si se proporciona email, validar que no exista en Cognito
	if req.Email != "" {
		userPoolID := os.Getenv("COGNITO_USER_POOL_ID")
		if userPoolID == "" {
			return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Cognito not configured"})
		}
		
		// Intentar obtener el usuario por email
		listResult, err := cognitoClient.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
			UserPoolId: aws.String(userPoolID),
			Filter:     aws.String(fmt.Sprintf("email = \"%s\"", req.Email)),
			Limit:      aws.Int32(1),
		})
		
		if err != nil {
			log.Printf("Error checking if user exists: %v", err)
			return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to validate email"})
		}

		// Si el usuario ya existe, rechazar
		if len(listResult.Users) > 0 {
			return jsonResponse(409, map[string]string{
				"error":   "EMAIL_EXISTS",
				"message": "Este email ya está registrado en la plataforma",
			})
		}
	}

	initDynamoDB()

	// Generar código único
	code, err := generateInvitationCode()
	if err != nil {
		log.Printf("Error generating code: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to generate code"})
	}

	now := time.Now().Unix()
	expiresAt := now + (7 * 24 * 3600) // 7 días

	invitation := InvitationItem{
		InvitationCode: code,
		TenantID:       req.TenantID,
		Role:           req.Role,
		CreatedBy:      claims.Email,
		CreatedAt:      now,
		ExpiresAt:      expiresAt,
		Used:           false,
	}

	item, err := attributevalue.MarshalMap(invitation)
	if err != nil {
		log.Printf("Error marshaling invitation: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to create invitation"})
	}

	tableName := os.Getenv("INVITATIONS_TABLE")
	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		log.Printf("Error saving invitation: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to save invitation"})
	}

	// Enviar email si se proporcionó un email
	if req.Email != "" {
		if err := sendInvitationEmail(req.Email, code, req.TenantID); err != nil {
			log.Printf("Warning: Failed to send invitation email to %s: %v", req.Email, err)
			// No fallar la request si el email falla, solo loguear
		} else {
			log.Printf("Invitation email sent successfully to %s", req.Email)
		}
	}

	response := InvitationResponse{
		Code:      code,
		TenantID:  req.TenantID,
		Role:      req.Role,
		ExpiresAt: expiresAt,
		CreatedBy: claims.Email,
	}

	return jsonResponse(201, response)
}

func handleGetInvitation(ctx context.Context, code string) (V2Response, error) {
	initDynamoDB()

	tableName := os.Getenv("INVITATIONS_TABLE")
	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]dynamodbtypes.AttributeValue{
			"invitationCode": &dynamodbtypes.AttributeValueMemberS{Value: code},
		},
	})
	if err != nil {
		log.Printf("Error getting invitation: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to get invitation"})
	}

	if result.Item == nil {
		return jsonResponse(404, map[string]string{"error": "NOT_FOUND", "message": "Invitation not found or expired"})
	}

	var invitation InvitationItem
	if err := attributevalue.UnmarshalMap(result.Item, &invitation); err != nil {
		log.Printf("Error unmarshaling invitation: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to parse invitation"})
	}

	if invitation.Used {
		return jsonResponse(400, map[string]string{"error": "ALREADY_USED", "message": "Invitation already used"})
	}

	if time.Now().Unix() > invitation.ExpiresAt {
		return jsonResponse(400, map[string]string{"error": "EXPIRED", "message": "Invitation expired"})
	}

	response := InvitationResponse{
		Code:      invitation.InvitationCode,
		TenantID:  invitation.TenantID,
		Role:      invitation.Role,
		ExpiresAt: invitation.ExpiresAt,
		CreatedBy: invitation.CreatedBy,
	}

	return jsonResponse(200, response)
}

func handleRegister(ctx context.Context, body string) (V2Response, error) {
	var req RegisterRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, map[string]string{"error": "INVALID_REQUEST", "message": "Invalid JSON"})
	}

	// Validar invitación
	initDynamoDB()
	tableName := os.Getenv("INVITATIONS_TABLE")
	result, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]dynamodbtypes.AttributeValue{
			"invitationCode": &dynamodbtypes.AttributeValueMemberS{Value: req.InvitationCode},
		},
	})
	if err != nil || result.Item == nil {
		return jsonResponse(400, map[string]string{"error": "INVALID_CODE", "message": "Invalid or expired invitation code"})
	}

	var invitation InvitationItem
	if err := attributevalue.UnmarshalMap(result.Item, &invitation); err != nil {
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to parse invitation"})
	}

	if invitation.Used {
		return jsonResponse(400, map[string]string{"error": "ALREADY_USED", "message": "Invitation already used"})
	}

	if time.Now().Unix() > invitation.ExpiresAt {
		return jsonResponse(400, map[string]string{"error": "EXPIRED", "message": "Invitation expired"})
	}

	// Crear usuario en Cognito
	_, err = cognitoClient.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:        aws.String(userPoolID),
		Username:          aws.String(req.Email),
		MessageAction:     cognitotypes.MessageActionTypeSuppress,
		TemporaryPassword: aws.String(req.Password),
		UserAttributes: []cognitotypes.AttributeType{
			{Name: aws.String("email"), Value: aws.String(req.Email)},
			{Name: aws.String("email_verified"), Value: aws.String("true")},
			{Name: aws.String("custom:tenant_id"), Value: aws.String(invitation.TenantID)},
			{Name: aws.String("custom:role"), Value: aws.String(invitation.Role)},
		},
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return jsonResponse(500, map[string]string{"error": "USER_CREATION_FAILED", "message": fmt.Sprintf("Failed to create user: %v", err)})
	}

	// Establecer contraseña permanente
	_, err = cognitoClient.AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(req.Email),
		Password:   aws.String(req.Password),
		Permanent:  true,
	})
	if err != nil {
		log.Printf("Error setting password: %v", err)
	}

	// Marcar invitación como usada
	_, err = dynamoClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]dynamodbtypes.AttributeValue{
			"invitationCode": &dynamodbtypes.AttributeValueMemberS{Value: req.InvitationCode},
		},
		UpdateExpression: aws.String("SET used = :used"),
		ExpressionAttributeValues: map[string]dynamodbtypes.AttributeValue{
			":used": &dynamodbtypes.AttributeValueMemberBOOL{Value: true},
		},
	})
	if err != nil {
		log.Printf("Error marking invitation as used: %v", err)
	}

	return jsonResponse(201, map[string]string{
		"message": "User registered successfully",
		"email":   req.Email,
	})
}
