# Sistema de Invitaciones con Resend

## Descripción General

Sistema de invitaciones por código que permite a administradores invitar nuevos usuarios a la plataforma con envío automático de emails transaccionales mediante Resend.

## Arquitectura

```
Admin → Frontend → API Gateway → admin-handler Lambda
                                      ↓
                                  Secrets Manager (Resend API Key)
                                      ↓
                                  Resend API
                                      ↓
                                  Email al invitado
```

## Componentes

### 1. AWS Secrets Manager

**Secreto:** `prod-resend-api-key`

- Almacena el API key de Resend de forma segura
- Permite rotación sin redesplegar Lambda
- Auditoría completa con CloudTrail
- Políticas IAM de mínimo privilegio

**Configuración:**

```hcl
module "resend_secret" {
  source = "../../modules/secrets-manager"

  environment             = "prod"
  secret_name             = "resend-api-key"
  description             = "Resend API key for sending invitation emails"
  secret_value            = var.resend_api_key
  recovery_window_in_days = 7
}
```

### 2. Lambda admin-handler

**Variables de entorno:**

- `RESEND_SECRET_NAME`: Nombre del secreto en Secrets Manager
- `APP_URL`: URL base de la aplicación
- `COGNITO_USER_POOL_ID`: ID del User Pool
- `INVITATIONS_TABLE`: Nombre de la tabla DynamoDB

**Permisos IAM:**

```json
{
  "Effect": "Allow",
  "Action": ["secretsmanager:GetSecretValue"],
  "Resource": "arn:aws:secretsmanager:*:*:secret:prod-resend-api-key-*"
}
```

### 3. Resend Integration

**Cliente Go con cache:**

```go
var (
    secretsClient *secretsmanager.Client
    cachedAPIKey  string
    cacheMutex    sync.RWMutex
)

func getResendAPIKey(ctx context.Context) (string, error) {
    // Check cache first
    cacheMutex.RLock()
    if cachedAPIKey != "" {
        cacheMutex.RUnlock()
        return cachedAPIKey, nil
    }
    cacheMutex.RUnlock()

    // Get from Secrets Manager
    result, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    })
    
    // Cache the API key
    cacheMutex.Lock()
    cachedAPIKey = *result.SecretString
    cacheMutex.Unlock()

    return cachedAPIKey, nil
}
```

**Ventajas del cache:**

- Primera llamada: ~50ms (lectura de Secrets Manager)
- Llamadas subsecuentes: <1ms (cache en memoria)
- Reduce costos de Secrets Manager
- Mejora latencia de respuesta

## Flujo de Invitación

### 1. Creación de Invitación

```
POST /api/v1/invitations
Authorization: Bearer <jwt_token>

{
  "tenantId": "tenant-123",
  "role": "USER_TENANT",
  "email": "usuario@ejemplo.com"  // Opcional
}
```

**Validaciones:**

1. ✅ Usuario autenticado (JWT válido)
2. ✅ Permisos adecuados (SUPER_ADMIN o REALM_ADMIN)
3. ✅ REALM_ADMIN solo puede invitar a su tenant
4. ✅ Rol válido según permisos del invitador
5. ✅ **Email no existe en Cognito** (nuevo)

**Proceso:**

```go
// 1. Validar que el email no exista
listResult, err := cognitoClient.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
    UserPoolId: aws.String(userPoolID),
    Filter:     aws.String(fmt.Sprintf("email = \"%s\"", req.Email)),
    Limit:      aws.Int32(1),
})

if len(listResult.Users) > 0 {
    return jsonResponse(409, map[string]string{
        "error":   "EMAIL_EXISTS",
        "message": "Este email ya está registrado en la plataforma",
    })
}

// 2. Generar código de 6 dígitos
code := generateInvitationCode()

// 3. Guardar en DynamoDB
invitation := InvitationItem{
    InvitationCode: code,
    TenantID:       req.TenantID,
    Role:           req.Role,
    ExpiresAt:      now + (7 * 24 * 3600), // 7 días
    Used:           false,
}

// 4. Enviar email si se proporcionó
if req.Email != "" {
    sendInvitationEmail(req.Email, code, req.TenantID)
}
```

**Respuesta:**

```json
{
  "code": "123456",
  "tenantId": "tenant-123",
  "role": "USER_TENANT",
  "expiresAt": 1738972800,
  "createdBy": "admin@ejemplo.com"
}
```

### 2. Email Template

**Asunto:** "Invitación a SQL to NoSQL Parser"

**Contenido HTML:**

- Header con título
- Código de invitación destacado
- Botón CTA con link directo al registro
- Información de expiración (7 días)
- Footer con branding

**Variables:**

- `{{invitationCode}}`: Código de 6 dígitos
- `{{registerLink}}`: `https://app-sql.cloudcentinel.com/register?code=123456`

### 3. Validación de Código

```
GET /api/v1/invitations/{code}
```

**Validaciones:**

1. ✅ Código existe en DynamoDB
2. ✅ No ha sido usado (`used: false`)
3. ✅ No ha expirado (`expiresAt > now`)

**Respuesta:**

```json
{
  "code": "123456",
  "tenantId": "tenant-123",
  "role": "USER_TENANT",
  "expiresAt": 1738972800
}
```

### 4. Registro con Código

```
POST /api/v1/register

{
  "invitationCode": "123456",
  "email": "usuario@ejemplo.com",
  "password": "SecurePass123!"
}
```

**Proceso:**

1. Validar código (existe, no usado, no expirado)
2. Crear usuario en Cognito con atributos del tenant
3. Marcar invitación como usada
4. Retornar confirmación

## Seguridad

### Validación de Email Único

**Problema:** Prevenir invitaciones duplicadas a usuarios existentes

**Solución:**

```go
// Consultar Cognito antes de crear invitación
listResult, err := cognitoClient.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
    UserPoolId: aws.String(userPoolID),
    Filter:     aws.String(fmt.Sprintf("email = \"%s\"", email)),
    Limit:      aws.Int32(1),
})

if len(listResult.Users) > 0 {
    return error409("EMAIL_EXISTS")
}
```

**Beneficios:**

- ✅ Previene spam de invitaciones
- ✅ Mejor experiencia de usuario
- ✅ Mensaje de error claro
- ✅ No se envían emails innecesarios

### Rotación de API Key

**Sin downtime:**

```bash
# 1. Actualizar secreto en Secrets Manager
aws secretsmanager update-secret \
  --secret-id prod-resend-api-key \
  --secret-string "nuevo_api_key_aqui"

# 2. Esperar al próximo cold start de Lambda
# O forzar actualización desplegando nueva versión
```

**La Lambda automáticamente:**

- Lee el nuevo valor en el próximo cold start
- Actualiza el cache en memoria
- No requiere cambios de código

### Políticas IAM

**Mínimo privilegio:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["secretsmanager:GetSecretValue"],
      "Resource": "arn:aws:secretsmanager:us-east-1:*:secret:prod-resend-api-key-*"
    },
    {
      "Effect": "Allow",
      "Action": ["cognito-idp:ListUsers", "cognito-idp:AdminCreateUser"],
      "Resource": "arn:aws:cognito-idp:us-east-1:*:userpool/us-east-1_*"
    },
    {
      "Effect": "Allow",
      "Action": ["dynamodb:PutItem", "dynamodb:GetItem", "dynamodb:UpdateItem"],
      "Resource": "arn:aws:dynamodb:us-east-1:*:table/prod-invitations"
    }
  ]
}
```

## Monitoreo

### CloudWatch Metrics

**Métricas clave:**

- `InvitationsCreated`: Total de invitaciones generadas
- `EmailsSent`: Emails enviados exitosamente
- `EmailsFailed`: Fallos en envío de email
- `DuplicateEmailAttempts`: Intentos de invitar emails existentes

### CloudWatch Logs

**Logs importantes:**

```
[INFO] Invitation created: code=123456, tenant=tenant-123, role=USER_TENANT
[INFO] Invitation email sent successfully to usuario@ejemplo.com
[WARN] Email already exists: usuario@ejemplo.com
[ERROR] Failed to send invitation email: <error details>
```

### CloudTrail

**Auditoría de acceso a secretos:**

- Quién accedió al secreto
- Cuándo se accedió
- Desde qué Lambda
- IP de origen

## Costos

### Resend

- **Tier gratuito:** 100 emails/día, 3,000 emails/mes
- **Tier pagado:** $20/mes por 50,000 emails

### AWS Secrets Manager

- **Almacenamiento:** $0.40/secreto/mes
- **API calls:** $0.05 por 10,000 requests
- **Con cache:** ~1 request por cold start

### Estimación mensual

**Escenario: 50 invitaciones/día**

- Resend: $0 (dentro del tier gratuito)
- Secrets Manager: $0.40/mes
- API calls: ~$0.01/mes (200 cold starts)

**Total: ~$0.41/mes**

## Troubleshooting

### Email no llega

1. Verificar logs de Lambda: `aws logs tail /aws/lambda/prod-admin_handler --follow`
2. Verificar respuesta de Resend API
3. Revisar spam folder del destinatario
4. Verificar dominio verificado en Resend

### Error "EMAIL_EXISTS"

- Usuario ya está registrado en la plataforma
- No se puede invitar al mismo email dos veces
- Solución: Usar otro email o contactar al usuario existente

### Error al leer secreto

1. Verificar permisos IAM de la Lambda
2. Verificar que el secreto existe: `aws secretsmanager describe-secret --secret-id prod-resend-api-key`
3. Revisar logs de CloudWatch

## Mejoras Futuras

- [ ] Reintentos automáticos para emails fallidos
- [ ] Dashboard de invitaciones pendientes
- [ ] Notificaciones cuando una invitación es usada
- [ ] Templates de email personalizables por tenant
- [ ] Límite de invitaciones por admin/día
- [ ] Métricas de conversión (invitaciones → registros)
