# Sistema de Roles y Multi-Tenancy

## Roles y Niveles

| Rol | Nivel | Permisos |
|-----|-------|----------|
| `SUPER_ADMIN` | 4 | Acceso total sin restricciones de tenant |
| `REALM_ADMIN` | 3 | Gestión completa dentro de su tenant. Puede crear/editar usuarios nivel ≤2 |
| `REALM_SUPERVISOR` | 2 | Solo lectura de usuarios nivel 1 dentro de su tenant |
| `USER_TENANT` | 1 | Acceso exclusivo a su propia información |

## Atributos Custom en Cognito

```
custom:tenant_id  → UUID del tenant (obligatorio excepto SUPER_ADMIN)
custom:role       → Uno de los 4 roles definidos
```

## Flujo de Autenticación

### 1. Registro de Usuario
```
Usuario → Cognito SignUp
       → Post-Confirmation Trigger
       → Asigna custom:tenant_id (default: "public")
       → Asigna custom:role (default: "USER_TENANT")
```

### 2. Login
```
Usuario → Cognito SignIn
       → Retorna JWT con claims:
         {
           "sub": "user-uuid",
           "email": "user@example.com",
           "custom:tenant_id": "tenant-uuid",
           "custom:role": "USER_TENANT"
         }
```

### 3. Request a API
```
Frontend → API Gateway (valida JWT)
        → Lambda (extrae claims del JWT)
        → Aplica filtros por tenant y rol
        → DynamoDB (query con tenant_id)
```

## Lógica de Autorización

### Acceso a Recursos (Read)

```go
// SUPER_ADMIN: ve todo
if role == "SUPER_ADMIN" {
    return allRecords
}

// Otros roles: solo su tenant
return recordsWhere(tenant_id == user.tenant_id)
```

### Gestión de Usuarios (Write)

```go
// SUPER_ADMIN: puede gestionar cualquier usuario
if manager.role == "SUPER_ADMIN" {
    return true
}

// Mismo tenant requerido
if manager.tenant_id != target.tenant_id {
    return false
}

// REALM_ADMIN: puede gestionar nivel ≤2
if manager.role == "REALM_ADMIN" && target.level <= 2 {
    return true
}

// REALM_SUPERVISOR: solo lectura (no write)
// USER_TENANT: solo su propia info
return false
```

## Implementación

### 1. Cognito (Terraform)

```hcl
# infra/terraform/modules/cognito/main.tf

# Atributos custom (mutable = true para permitir cambios via AdminUpdateUserAttributes)
schema {
  name                = "tenant_id"
  attribute_data_type = "String"
  mutable             = true
}

schema {
  name                = "role"
  attribute_data_type = "String"
  mutable             = true
}
```

#### Permisos de atributos en App Client

Los atributos `custom:tenant_id` y `custom:role` son **read-only** a nivel de app client.
Esto impide que usuarios modifiquen estos valores desde el SPA.

Solo SUPER_ADMIN y REALM_ADMIN pueden cambiarlos via `AdminUpdateUserAttributes` en el Lambda admin-handler (bypasea restricciones del client).

Ref: [Multi-tenancy Security Recommendations](https://docs.aws.amazon.com/cognito/latest/developerguide/multi-tenancy-security-recommendations.html)

```hcl
# App Client - Restricción de escritura
read_attributes = [
  "email",
  "email_verified",
  "custom:tenant_id",
  "custom:role",
]

write_attributes = [
  "email",  # Solo email es escribible por el usuario
]
```

#### Deletion Protection

Habilitado en producción (`deletion_protection = true`) para prevenir eliminación accidental del User Pool via Terraform.

### 2. Trigger Lambda

```go
// lambda/cognito-trigger/main.go
// Asigna tenant_id y role por defecto al confirmar usuario
UserAttributes: []types.AttributeType{
    {Name: "custom:tenant_id", Value: "public"},
    {Name: "custom:role", Value: "USER_TENANT"},
}
```

### 3. Paquete de Autorización

```go
// lambda/pkg/auth/auth.go
func GetTenantFilter(claims *UserClaims) string
func CanAccessTenant(claims *UserClaims, targetTenant string) bool
func CanManageUser(manager *UserClaims, targetRole, targetTenant string) bool
func CanReadUser(reader *UserClaims, targetUserID, targetTenant string) bool
```

### 4. Lambdas de API

```go
// Extraer claims del JWT
claims := extractClaims(req.RequestContext.Authorizer.JWT)

// Aplicar filtro de tenant
tenantFilter := auth.GetTenantFilter(claims)

// Query DynamoDB con filtro
records := queryWithTenant(tenantFilter)
```

## Queries DynamoDB

### Tabla `schemas`

```
PK: conversionId (UUID)
SK: tenant_id#timestamp
GSI: tenant_id-createdAt-index
```

### Query por Tenant

```go
// SUPER_ADMIN (sin filtro)
QueryInput{
    IndexName: "tenant_id-createdAt-index",
    // Sin KeyConditionExpression = retorna todo
}

// Otros roles (con filtro)
QueryInput{
    IndexName: "tenant_id-createdAt-index",
    KeyConditionExpression: "tenant_id = :tid",
    ExpressionAttributeValues: {
        ":tid": userClaims.TenantID,
    },
}
```

## Variables de Entorno

### Cognito Trigger Lambda

```bash
DEFAULT_TENANT_ID=public    # Tenant por defecto para nuevos usuarios
DEFAULT_ROLE=USER_TENANT    # Rol por defecto para nuevos usuarios
```

### API Lambdas

```bash
SCHEMAS_TABLE_NAME=prod-schemas
```

## Testing

### Crear SUPER_ADMIN

```bash
aws cognito-idp admin-create-user \
  --user-pool-id <POOL_ID> \
  --username admin@example.com \
  --user-attributes \
    Name=email,Value=admin@example.com \
    Name=custom:role,Value=SUPER_ADMIN \
  --temporary-password TempPass123!

aws cognito-idp admin-set-user-password \
  --user-pool-id <POOL_ID> \
  --username admin@example.com \
  --password AdminPass123! \
  --permanent
```

### Crear REALM_ADMIN

```bash
aws cognito-idp admin-create-user \
  --user-pool-id <POOL_ID> \
  --username realm-admin@example.com \
  --user-attributes \
    Name=email,Value=realm-admin@example.com \
    Name=custom:tenant_id,Value=tenant-123 \
    Name=custom:role,Value=REALM_ADMIN \
  --temporary-password TempPass123!
```

### Verificar JWT

```bash
# Login
TOKEN=$(aws cognito-idp initiate-auth \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id <CLIENT_ID> \
  --auth-parameters USERNAME=admin@example.com,PASSWORD=AdminPass123! \
  --query 'AuthenticationResult.IdToken' \
  --output text)

# Decodificar
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq
```

## Próximos Pasos

1. ✅ Agregar `custom:role` a Cognito
2. ✅ Actualizar trigger para asignar rol
3. ✅ Crear paquete `lambda/pkg/auth`
4. ⏳ Actualizar Query Handler con autorización
5. ⏳ Actualizar Process Handler con autorización
6. ⏳ Actualizar frontend para mostrar rol del usuario
7. ⏳ Crear scripts de gestión de usuarios admin
