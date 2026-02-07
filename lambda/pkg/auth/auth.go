package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// Role levels
const (
	RoleSuperAdmin      = "SUPER_ADMIN"
	RoleRealmAdmin      = "REALM_ADMIN"
	RoleRealmSupervisor = "REALM_SUPERVISOR"
	RoleUserTenant      = "USER_TENANT"
)

// Role hierarchy levels
const (
	LevelSuperAdmin      = 4
	LevelRealmAdmin      = 3
	LevelRealmSupervisor = 2
	LevelUserTenant      = 1
)

// UserClaims represents the JWT claims from Cognito
type UserClaims struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	TenantID string `json:"custom:tenant_id"`
	Role     string `json:"custom:role"`
}

// GetRoleLevel returns the numeric level for a role
func GetRoleLevel(role string) int {
	switch role {
	case RoleSuperAdmin:
		return LevelSuperAdmin
	case RoleRealmAdmin:
		return LevelRealmAdmin
	case RoleRealmSupervisor:
		return LevelRealmSupervisor
	case RoleUserTenant:
		return LevelUserTenant
	default:
		return 0
	}
}

// ParseJWT extracts claims from a JWT token (simple base64 decode, no verification)
// API Gateway already verified the token signature
func ParseJWT(token string) (*UserClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims UserClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

// CanAccessTenant checks if user can access resources from a specific tenant
func CanAccessTenant(userClaims *UserClaims, targetTenantID string) bool {
	// SUPER_ADMIN can access any tenant
	if userClaims.Role == RoleSuperAdmin {
		return true
	}

	// All other roles must match tenant
	return userClaims.TenantID == targetTenantID
}

// CanManageUser checks if user can manage another user based on roles and tenant
func CanManageUser(managerClaims *UserClaims, targetRole string, targetTenantID string) bool {
	managerLevel := GetRoleLevel(managerClaims.Role)
	targetLevel := GetRoleLevel(targetRole)

	// SUPER_ADMIN can manage anyone
	if managerClaims.Role == RoleSuperAdmin {
		return true
	}

	// Must be same tenant
	if managerClaims.TenantID != targetTenantID {
		return false
	}

	// REALM_ADMIN can manage REALM_SUPERVISOR and USER_TENANT
	if managerClaims.Role == RoleRealmAdmin {
		return targetLevel <= LevelRealmSupervisor
	}

	// REALM_SUPERVISOR can only read USER_TENANT (no write)
	// USER_TENANT cannot manage others
	return false
}

// CanReadUser checks if user can read another user's data
func CanReadUser(readerClaims *UserClaims, targetUserID string, targetTenantID string) bool {
	// SUPER_ADMIN can read anyone
	if readerClaims.Role == RoleSuperAdmin {
		return true
	}

	// Must be same tenant
	if readerClaims.TenantID != targetTenantID {
		return false
	}

	// REALM_ADMIN and REALM_SUPERVISOR can read within their tenant
	if readerClaims.Role == RoleRealmAdmin || readerClaims.Role == RoleRealmSupervisor {
		return true
	}

	// USER_TENANT can only read their own data
	return readerClaims.Sub == targetUserID
}

// GetTenantFilter returns the tenant filter for DynamoDB queries
func GetTenantFilter(userClaims *UserClaims) string {
	// SUPER_ADMIN sees all tenants (no filter)
	if userClaims.Role == RoleSuperAdmin {
		return ""
	}

	// All other roles are restricted to their tenant
	return userClaims.TenantID
}
