package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/google/uuid"
)

type V2Response = events.APIGatewayV2HTTPResponse

var (
	cognitoClient *cognitoidentityprovider.Client
	userPoolID    string
)

type UserClaims struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	TenantID string `json:"custom:tenant_id"`
	Role     string `json:"custom:role"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	TenantID string `json:"tenantId"`
}

type CreateTenantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TenantResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserCount   int    `json:"userCount"`
}

type UserResponse struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	TenantID string `json:"tenantId"`
	Enabled  bool   `json:"enabled"`
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	cognitoClient = cognitoidentityprovider.NewFromConfig(cfg)
	userPoolID = os.Getenv("COGNITO_USER_POOL_ID")
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (V2Response, error) {
	method := req.RequestContext.HTTP.Method
	path := req.RequestContext.HTTP.Path

	// Handle OPTIONS for CORS preflight
	if method == "OPTIONS" {
		return V2Response{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization",
			},
		}, nil
	}

	claims := extractClaims(req)
	if claims == nil {
		return jsonResponse(401, map[string]string{"error": "UNAUTHORIZED", "message": "Invalid token"})
	}

	log.Printf("Request: %s %s by %s (role: %s, tenant: %s)", method, path, claims.Email, claims.Role, claims.TenantID)

	// GET /api/v1/users
	if method == "GET" && strings.HasSuffix(path, "/api/v1/users") {
		return handleListUsers(ctx, claims)
	}

	// POST /api/v1/users
	if method == "POST" && strings.HasSuffix(path, "/api/v1/users") {
		return handleCreateUser(ctx, claims, req.Body)
	}

	// GET /api/v1/tenants
	if method == "GET" && strings.HasSuffix(path, "/api/v1/tenants") {
		return handleListTenants(ctx, claims)
	}

	// POST /api/v1/tenants
	if method == "POST" && strings.HasSuffix(path, "/api/v1/tenants") {
		return handleCreateTenant(ctx, claims, req.Body)
	}

	// POST /api/v1/invitations
	if method == "POST" && strings.HasSuffix(path, "/api/v1/invitations") {
		return handleCreateInvitation(ctx, claims, req.Body)
	}

	// GET /api/v1/invitations/{code}
	if method == "GET" && strings.Contains(path, "/api/v1/invitations/") {
		parts := strings.Split(path, "/")
		code := parts[len(parts)-1]
		return handleGetInvitation(ctx, code)
	}

	// POST /api/v1/register (público, sin autenticación)
	if method == "POST" && strings.HasSuffix(path, "/api/v1/register") {
		return handleRegister(ctx, req.Body)
	}

	return jsonResponse(404, map[string]string{"error": "NOT_FOUND", "message": "Route not found"})
}

func extractClaims(req events.APIGatewayV2HTTPRequest) *UserClaims {
	// Permitir registro sin autenticación
	if req.RequestContext.HTTP.Method == "POST" && strings.HasSuffix(req.RequestContext.HTTP.Path, "/api/v1/register") {
		return &UserClaims{} // Claims vacíos para registro público
	}

	// Permitir consulta de invitación sin autenticación
	if req.RequestContext.HTTP.Method == "GET" && strings.Contains(req.RequestContext.HTTP.Path, "/api/v1/invitations/") {
		return &UserClaims{} // Claims vacíos para consulta pública
	}

	if req.RequestContext.Authorizer == nil || req.RequestContext.Authorizer.JWT == nil {
		return nil
	}

	claims := req.RequestContext.Authorizer.JWT.Claims
	return &UserClaims{
		Sub:      claims["sub"],
		Email:    claims["email"],
		TenantID: claims["custom:tenant_id"],
		Role:     claims["custom:role"],
	}
}

func handleListUsers(ctx context.Context, claims *UserClaims) (V2Response, error) {
	if claims.Role != "SUPER_ADMIN" && claims.Role != "REALM_ADMIN" {
		return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Insufficient permissions"})
	}

	output, err := cognitoClient.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(userPoolID),
		Limit:      aws.Int32(60),
	})
	if err != nil {
		log.Printf("ERROR: ListUsers failed: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to list users"})
	}

	users := []UserResponse{}
	for _, user := range output.Users {
		// Get username (sub is not in attributes, it's the Username field)
		username := ""
		if user.Username != nil {
			username = *user.Username
		}

		userResp := UserResponse{
			Sub:      username,
			Email:    getAttributeValue(user.Attributes, "email"),
			Role:     getAttributeValue(user.Attributes, "custom:role"),
			TenantID: getAttributeValue(user.Attributes, "custom:tenant_id"),
			Enabled:  user.Enabled,
		}

		// Filter by tenant for REALM_ADMIN
		if claims.Role == "REALM_ADMIN" && userResp.TenantID != claims.TenantID {
			continue
		}

		users = append(users, userResp)
	}

	return jsonResponse(200, map[string]interface{}{"users": users, "count": len(users)})
}

func handleCreateUser(ctx context.Context, claims *UserClaims, body string) (V2Response, error) {
	if claims.Role != "SUPER_ADMIN" && claims.Role != "REALM_ADMIN" {
		return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Insufficient permissions"})
	}

	var req CreateUserRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, map[string]string{"error": "INVALID_REQUEST", "message": "Invalid JSON"})
	}

	// Validate role permissions
	if claims.Role == "REALM_ADMIN" {
		if req.Role == "SUPER_ADMIN" || req.Role == "REALM_ADMIN" {
			return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Cannot create users with higher or equal role"})
		}
		if req.TenantID != claims.TenantID {
			return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Cannot create users in other tenants"})
		}
	}

	// Create user in Cognito
	_, err := cognitoClient.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(req.Email),
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(req.Email)},
			{Name: aws.String("email_verified"), Value: aws.String("true")},
			{Name: aws.String("custom:role"), Value: aws.String(req.Role)},
			{Name: aws.String("custom:tenant_id"), Value: aws.String(req.TenantID)},
		},
		MessageAction: types.MessageActionTypeSuppress,
	})
	if err != nil {
		log.Printf("ERROR: AdminCreateUser failed: %v", err)
		return jsonResponse(500, map[string]string{"error": "CREATE_FAILED", "message": "Failed to create user"})
	}

	// Set permanent password
	_, err = cognitoClient.AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(req.Email),
		Password:   aws.String(req.Password),
		Permanent:  true,
	})
	if err != nil {
		log.Printf("ERROR: AdminSetUserPassword failed: %v", err)
		return jsonResponse(500, map[string]string{"error": "PASSWORD_FAILED", "message": "User created but password failed"})
	}

	return jsonResponse(201, map[string]interface{}{
		"message": "User created successfully",
		"email":   req.Email,
		"role":    req.Role,
		"tenantId": req.TenantID,
	})
}

func handleListTenants(ctx context.Context, claims *UserClaims) (V2Response, error) {
	if claims.Role != "SUPER_ADMIN" {
		return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Only SUPER_ADMIN can list tenants"})
	}

	// Get all users and extract unique tenant_ids
	output, err := cognitoClient.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(userPoolID),
		Limit:      aws.Int32(60),
	})
	if err != nil {
		log.Printf("ERROR: ListUsers failed: %v", err)
		return jsonResponse(500, map[string]string{"error": "INTERNAL_ERROR", "message": "Failed to list tenants"})
	}

	tenantMap := make(map[string]int)
	for _, user := range output.Users {
		tenantID := getAttributeValue(user.Attributes, "custom:tenant_id")
		if tenantID != "" {
			tenantMap[tenantID]++
		}
	}

	tenants := []TenantResponse{}
	for tenantID, count := range tenantMap {
		tenants = append(tenants, TenantResponse{
			ID:          tenantID,
			Name:        tenantID,
			Description: "Auto-generated from users",
			UserCount:   count,
		})
	}

	return jsonResponse(200, map[string]interface{}{"tenants": tenants, "count": len(tenants)})
}

func handleCreateTenant(ctx context.Context, claims *UserClaims, body string) (V2Response, error) {
	if claims.Role != "SUPER_ADMIN" {
		return jsonResponse(403, map[string]string{"error": "FORBIDDEN", "message": "Only SUPER_ADMIN can create tenants"})
	}

	var req CreateTenantRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, map[string]string{"error": "INVALID_REQUEST", "message": "Invalid JSON"})
	}

	tenantID := uuid.New().String()

	return jsonResponse(201, map[string]interface{}{
		"message":     "Tenant created successfully",
		"id":          tenantID,
		"name":        req.Name,
		"description": req.Description,
	})
}

func getAttributeValue(attrs []types.AttributeType, name string) string {
	for _, attr := range attrs {
		if attr.Name != nil && *attr.Name == name && attr.Value != nil {
			return *attr.Value
		}
	}
	return ""
}

func jsonResponse(statusCode int, body interface{}) (V2Response, error) {
	b, _ := json.Marshal(body)
	return V2Response{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
		},
		Body: string(b),
	}, nil
}

func main() {
	lambda.Start(handler)
}
