# SQL to DynamoDB Converter

Sistema serverless que automatiza la migración de esquemas SQL relacionales a diseños optimizados de DynamoDB usando inteligencia artificial (AWS Bedrock - Claude). Incluye autenticación multi-tenant con AWS Cognito y gestión de usuarios con roles.

## Resumen del Proyecto

| Aspecto           | Tecnología                              |
| ----------------- | --------------------------------------- |
| **Frontend**      | Vue 3 + TypeScript + PrimeVue           |
| **Backend**       | 6 Lambda Functions (Go 1.x, ARM64)      |
| **IA**            | Claude 3.5 Sonnet v2 + Claude 3.5 Haiku |
| **Autenticación** | AWS Cognito (Multi-tenant)              |
| **Base de datos** | DynamoDB (TTL 24h)                      |
| **Mensajería**    | SQS (2 colas + 2 DLQ)                   |
| **API**           | HTTP API Gateway v2                     |
| **CDN**           | CloudFront + Cloudflare                 |
| **Email**         | Resend (invitaciones automáticas)       |
| **Secretos**      | AWS Secrets Manager                     |
| **IaC**           | Terraform (11 módulos)                  |

## Estructura del Proyecto

```
sql-to-nosql-parser/
├── lambda/                    # Backend - 6 funciones Lambda (Go)
│   ├── diagrams/              # Process Handler - validación SQL
│   ├── conversion-worker/     # SQL → DynamoDB (Bedrock Sonnet)
│   ├── access-pattern-worker/ # Access patterns (Bedrock Haiku)
│   ├── query/                 # Query Handler - obtener resultados
│   ├── admin-handler/         # Admin Handler - gestión usuarios/tenants
│   └── dlq-handler/           # Manejo de mensajes fallidos
│
├── web/db-parser/             # Frontend - Vue 3 SPA
│   └── src/
│       ├── views/             # HomeView, HistoryView, TenantsView, LoginView
│       ├── components/        # Sidebar
│       ├── services/          # API client (schemasApi, adminApi, authService)
│       ├── stores/            # Pinia stores (auth)
│       └── models/            # TypeScript interfaces
│
├── infra/terraform/           # Infraestructura como código
│   ├── modules/               # 11 módulos reutilizables
│   │   ├── lambda/            # Configuración Lambda genérica
│   │   ├── gateway/           # API Gateway (HTTP v2 + REST v1)
│   │   ├── dynamodb/          # Tabla con TTL y GSI
│   │   ├── sqs/               # Colas + DLQ
│   │   ├── s3/                # Bucket frontend
│   │   ├── cloudfront/        # CDN + OAC + ACM + CloudFront Function
│   │   ├── cognito/           # User Pool + Client + Groups
│   │   ├── bedrock/           # Model access
│   │   ├── secrets-manager/   # Gestión segura de API keys
│   │   └── iam/               # Roles y políticas
│   ├── environments/dev/      # LocalStack (desarrollo)
│   └── environments/prod/     # AWS (producción)
│
├── docs/                      # Documentación detallada
│   ├── infrastructure/        # Guía de Terraform
│   ├── backend/               # Guía de Lambda/Go
│   ├── frontend/              # Guía de Vue
│   └── architecture/          # Diagramas y flujos
│
└── Makefile                   # Comandos de automatización
```

## Flujo de Procesamiento

```
Usuario (SQL) → Frontend → API Gateway → Process Handler
                                              ↓
                                         DynamoDB (PENDING)
                                              ↓
                                         SQS Queue
                                              ↓
                          Conversion Worker (Bedrock Sonnet)
                                              ↓
                                         SQS Queue
                                              ↓
                       Access Pattern Worker (Bedrock Haiku)
                                              ↓
                                    DynamoDB (COMPLETED)
                                              ↓
                        Frontend ← API Gateway ← Query Handler
```

## Sistema de Autenticación Multi-Tenant

### Roles y Permisos

| Rol                | Descripción                                      | Acceso                          |
| ------------------ | ------------------------------------------------ | ------------------------------- |
| **SUPER_ADMIN**    | Administrador global (creado manualmente)        | Todos los tenants y usuarios    |
| **REALM_ADMIN**    | Administrador de tenant                          | Usuarios de su tenant           |
| **REALM_SUPERVISOR** | Supervisor con permisos de lectura             | Conversiones de su tenant       |
| **USER_TENANT**    | Usuario estándar                                 | Conversiones de su tenant       |

### Flujo de Autenticación

```
Login → Cognito → JWT Token → API Gateway Authorizer → Lambda
                                      ↓
                              Validación de tenant
                                      ↓
                              Filtrado por tenantId
```

### Gestión de Usuarios y Tenants

- **Vista unificada**: Click en tenant → Modal con usuarios
- **Creación de usuarios**: Directamente desde el modal del tenant
- **Roles configurables**: SUPER_ADMIN puede asignar REALM_ADMIN
- **Aislamiento**: Cada tenant solo ve sus propios datos

## Comandos Principales

```bash
# Desarrollo Local (LocalStack)
make docker-up        # Iniciar LocalStack
make lambda           # Compilar Lambdas
make localstack       # Deploy infraestructura local
make frontend         # Build SPA

# Producción
make prod-plan        # Plan de cambios
make prod             # Aplicar cambios
make deploy-frontend  # Deploy SPA + invalidar cache
```

## Estados de Conversión

```
PENDING → PROCESSING → DESIGN_COMPLETED → PROCESSING_PATTERNS → COMPLETED
              ↓                                    ↓
           FAILED                               FAILED
```

## Documentación Detallada

- [Infraestructura Terraform](docs/infrastructure/README.md)
- [Backend Lambda/Go](docs/backend/README.md)
- [Frontend Vue](docs/frontend/README.md)
- [Arquitectura General](docs/architecture/README.md)

## Variables de Entorno

```bash
# .env
LOCALSTACK_AUTH_TOKEN=...           # LocalStack Pro license
AWS_ACCESS_KEY_ID=...               # AWS credentials
AWS_SECRET_ACCESS_KEY=...
TF_VAR_cloudflare_secret_header_value=...  # Header de seguridad

# Frontend (.env.production)
VITE_BASE_PATH_URL=https://app-sql.cloudcentinel.com
VITE_ENDPOINT_URL=prod/api/v1/schemas
VITE_COGNITO_USER_POOL_ID=...
VITE_COGNITO_CLIENT_ID=...
```

## Endpoints API

### Conversiones SQL

| Método | Ruta                   | Lambda          | Descripción         |
| ------ | ---------------------- | --------------- | ------------------- |
| POST   | `/api/v1/schemas`      | process-handler | Iniciar conversión  |
| GET    | `/api/v1/schemas`      | query-handler   | Listar conversiones |
| GET    | `/api/v1/schemas/{id}` | query-handler   | Obtener por ID      |

### Administración (requiere autenticación)

| Método | Ruta                          | Lambda        | Roles                      | Descripción                  |
| ------ | ----------------------------- | ------------- | -------------------------- | ---------------------------- |
| GET    | `/api/v1/users`               | admin-handler | SUPER_ADMIN, REALM_ADMIN   | Listar usuarios              |
| POST   | `/api/v1/users`               | admin-handler | SUPER_ADMIN, REALM_ADMIN   | Crear usuario                |
| GET    | `/api/v1/tenants`             | admin-handler | SUPER_ADMIN, REALM_ADMIN   | Listar tenants               |
| POST   | `/api/v1/tenants`             | admin-handler | SUPER_ADMIN                | Crear tenant                 |
| POST   | `/api/v1/invitations`         | admin-handler | SUPER_ADMIN, REALM_ADMIN   | Crear invitación (con email) |
| GET    | `/api/v1/invitations/{code}`  | admin-handler | Público                    | Validar código               |
| POST   | `/api/v1/register`            | admin-handler | Público                    | Registro con código          |

## Sistema de Invitaciones

### Flujo Completo

1. **Admin genera invitación**
   - Selecciona rol (USER_TENANT, REALM_SUPERVISOR, REALM_ADMIN*)
   - Opcionalmente ingresa email del invitado
   - Sistema valida que el email no exista en Cognito
   - Genera código de 6 dígitos (expira en 7 días)

2. **Envío automático de email** (si se proporciona email)
   - Lambda lee API key desde AWS Secrets Manager (con cache)
   - Envía email via Resend con template HTML profesional
   - Incluye código y link directo: `https://app-sql.cloudcentinel.com/register?code=123456`

3. **Usuario completa registro**
   - Abre link o ingresa código manualmente
   - Frontend valida código con backend
   - Usuario ingresa email y password
   - Backend crea usuario en Cognito con atributos del tenant
   - Marca invitación como usada

### Características

- ✅ **Validación de email único**: Previene invitaciones a usuarios existentes
- ✅ **Envío automático**: Resend API (100 emails/día gratis)
- ✅ **Seguridad**: API key en AWS Secrets Manager
- ✅ **Performance**: Cache en memoria del API key
- ✅ **Rotación sin downtime**: Actualizar secreto sin redesplegar
- ✅ **Auditoría**: CloudTrail registra acceso a secretos

### Resend Integration

**Configuración:**

```go
// Lambda lee secreto con cache
apiKey, err := getResendAPIKey(ctx)

// Envía email
emailReq := ResendEmailRequest{
    From:    "SQL to NoSQL Parser <onboarding@resend.dev>",
    To:      []string{email},
    Subject: "Invitación a SQL to NoSQL Parser",
    HTML:    htmlTemplate,
}
```

**Costos estimados (50 invitaciones/día):**

- Resend: $0 (tier gratuito)
- Secrets Manager: $0.40/mes
- Total: ~$0.41/mes

## Seguridad

### Cloudflare + CloudFront

- **Full (strict) SSL**: Certificado ACM validado por DNS
- **Header secreto**: CloudFront Function valida `x-origin-secret`
- **Transform Rule**: Cloudflare inyecta el header automáticamente
- **Resultado**: Acceso directo a CloudFront bloqueado (403)

### Cognito

- **JWT Tokens**: Validación en API Gateway
- **Custom Claims**: `tenantId` y `role` en el token
- **Password Policy**: Mínimo 8 caracteres, mayúsculas, números
- **MFA**: Opcional por usuario
- **Deletion Protection**: Habilitado en producción (previene eliminación accidental del User Pool)
- **Attribute Permissions**: `custom:tenant_id` y `custom:role` son **read-only** para el app client (SPA). Solo modificables via `AdminUpdateUserAttributes` por SUPER_ADMIN y REALM_ADMIN desde el backend

## Tema Visual

- **Color primario**: `#3b82f6` (azul)
- **Framework UI**: PrimeVue Aura preset personalizado
- **Responsive**: Mobile-first con sidebar colapsable
- **Accesibilidad**: ARIA labels y navegación por teclado
