# SQL to DynamoDB Converter

Sistema serverless que automatiza la migración de esquemas SQL relacionales a diseños optimizados de DynamoDB usando inteligencia artificial (AWS Bedrock - Claude).

## Resumen del Proyecto

| Aspecto | Tecnología |
|---------|------------|
| **Frontend** | Vue 3 + TypeScript + PrimeVue |
| **Backend** | 5 Lambda Functions (Go 1.x, ARM64) |
| **IA** | Claude 3.5 Sonnet v2 + Claude 3.5 Haiku |
| **Base de datos** | DynamoDB (TTL 24h) |
| **Mensajería** | SQS (2 colas + 2 DLQ) |
| **API** | HTTP API Gateway v2 |
| **CDN** | CloudFront + Cloudflare |
| **IaC** | Terraform (10 módulos) |

## Estructura del Proyecto

```
sql-to-nosql-parser/
├── lambda/                    # Backend - 5 funciones Lambda (Go)
│   ├── diagrams/              # Process Handler - validación SQL
│   ├── conversion-worker/     # SQL → DynamoDB (Bedrock Sonnet)
│   ├── access-pattern-worker/ # Access patterns (Bedrock Haiku)
│   ├── query/                 # Query Handler - obtener resultados
│   └── dlq-handler/           # Manejo de mensajes fallidos
│
├── web/db-parser/             # Frontend - Vue 3 SPA
│   └── src/
│       ├── views/             # HomeView, HistoryView
│       ├── components/        # Sidebar
│       ├── services/          # API client
│       └── models/            # TypeScript interfaces
│
├── infra/terraform/           # Infraestructura como código
│   ├── modules/               # 10 módulos reutilizables
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
```

## Endpoints API

| Método | Ruta | Lambda | Descripción |
|--------|------|--------|-------------|
| POST | `/api/v1/schemas` | process-handler | Iniciar conversión |
| GET | `/api/v1/schemas` | query-handler | Listar conversiones |
| GET | `/api/v1/schemas/{id}` | query-handler | Obtener por ID |
