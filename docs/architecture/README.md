# Arquitectura General

Documentación de la arquitectura completa del sistema SQL to DynamoDB Converter.

## Tabla de Contenidos

1. [Visión General](#visión-general)
2. [Diagrama de Arquitectura](#diagrama-de-arquitectura)
3. [Flujo de Datos](#flujo-de-datos)
4. [Componentes del Sistema](#componentes-del-sistema)
5. [Patrones de Diseño](#patrones-de-diseño)
6. [Seguridad](#seguridad)
7. [Escalabilidad](#escalabilidad)
8. [Manejo de Errores](#manejo-de-errores)
9. [Costos](#costos)

---

## Visión General

SQL to DynamoDB Converter es una plataforma **serverless** que automatiza la migración de esquemas SQL relacionales a diseños optimizados de DynamoDB usando inteligencia artificial.

### Principios de Arquitectura

1. **Serverless-First**: Sin servidores que administrar, pago por uso
2. **Event-Driven**: Comunicación asíncrona entre componentes
3. **Single-Table Design**: DynamoDB optimizado con un solo table
4. **AI-Powered**: Claude para generación inteligente de diseños
5. **Infrastructure as Code**: Todo definido en Terraform

### Características Clave

| Característica | Implementación |
|----------------|----------------|
| Alta disponibilidad | Multi-AZ automático (AWS managed) |
| Escalabilidad | Auto-scaling Lambda + on-demand DynamoDB |
| Seguridad | IAM least privilege, TLS everywhere |
| Observabilidad | CloudWatch Logs/Metrics/Alarms |
| Costo optimizado | ARM64 Graviton, TTL automático |

---

## Diagrama de Arquitectura

### Vista de Alto Nivel

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                  USUARIOS                                   │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLOUDFLARE                                     │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐                   │
│  │  DNS          │  │  WAF          │  │  DDoS         │                   │
│  │  Routing      │  │  Protection   │  │  Mitigation   │                   │
│  └───────────────┘  └───────────────┘  └───────────────┘                   │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLOUDFRONT CDN                                 │
│  ┌───────────────────────────┐  ┌───────────────────────────────────────┐  │
│  │  Edge Locations           │  │  CloudFront Function                  │  │
│  │  Global Caching           │  │  (x-origin-secret validation)         │  │
│  └───────────────────────────┘  └───────────────────────────────────────┘  │
│                                                                             │
│  Origins:                                                                   │
│  ┌─────────────────────────┐    ┌─────────────────────────────────────┐    │
│  │  S3 (Frontend)          │    │  API Gateway (Backend)              │    │
│  │  Path: /*               │    │  Path: /prod/*                      │    │
│  │  OAC Protected          │    │  Custom Header Auth                 │    │
│  └─────────────────────────┘    └─────────────────────────────────────┘    │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
              ┌───────────────────────┴───────────────────────┐
              │                                               │
              ▼                                               ▼
┌─────────────────────────────┐             ┌─────────────────────────────────┐
│         S3 BUCKET           │             │        API GATEWAY HTTP v2      │
│  ┌───────────────────────┐  │             │  ┌─────────────────────────┐    │
│  │  Vue 3 SPA            │  │             │  │  Routes                 │    │
│  │  - index.html         │  │             │  │  POST /api/v1/schemas   │    │
│  │  - assets/            │  │             │  │  GET  /api/v1/schemas   │    │
│  │  - js/css bundles     │  │             │  │  GET  /api/v1/schemas/* │    │
│  └───────────────────────┘  │             │  └─────────────────────────┘    │
└─────────────────────────────┘             └──────────────┬──────────────────┘
                                                           │
                            ┌──────────────────────────────┴─────────────────┐
                            │                                                │
                            ▼                                                ▼
              ┌─────────────────────────────┐            ┌───────────────────────────┐
              │      PROCESS HANDLER        │            │      QUERY HANDLER        │
              │        (Lambda)             │            │        (Lambda)           │
              │  ┌───────────────────────┐  │            │  ┌───────────────────┐    │
              │  │  Validate SQL         │  │            │  │  GetItem          │    │
              │  │  Create PENDING       │  │            │  │  Query by Date    │    │
              │  │  Enqueue message      │  │            │  │  Scan             │    │
              │  └───────────────────────┘  │            │  └───────────────────┘    │
              └──────────────┬──────────────┘            └─────────────┬─────────────┘
                             │                                         │
                             ▼                                         │
              ┌─────────────────────────────┐                          │
              │         SQS QUEUE           │                          │
              │    (conversion-queue)       │                          │
              │  ┌───────────────────────┐  │                          │
              │  │  Visibility: 180s    │  │                          │
              │  │  Max Receive: 3      │  │                          │
              │  │  Retention: 24h      │  │                          │
              │  └───────────────────────┘  │                          │
              └──────────────┬──────────────┘                          │
                             │                                         │
                             ▼                                         │
              ┌─────────────────────────────┐                          │
              │    CONVERSION WORKER        │                          │
              │        (Lambda)             │                          │
              │  ┌───────────────────────┐  │                          │
              │  │  Update PROCESSING   │  │                          │
              │  │  Invoke Bedrock      │  │                          │
              │  │  Save Design         │  │                          │
              │  │  Enqueue AP message  │  │                          │
              │  └───────────────────────┘  │                          │
              └──────────────┬──────────────┘                          │
                             │                                         │
              ┌──────────────┴──────────────┐                          │
              │                             │                          │
              ▼                             ▼                          │
┌─────────────────────────┐   ┌─────────────────────────────┐         │
│      AWS BEDROCK        │   │         SQS QUEUE           │         │
│  ┌───────────────────┐  │   │   (access-pattern-queue)    │         │
│  │  Claude 3.5       │  │   │  ┌───────────────────────┐  │         │
│  │  Sonnet v2        │  │   │  │  Visibility: 120s    │  │         │
│  │                   │  │   │  │  Max Receive: 3      │  │         │
│  │  SQL → DynamoDB   │  │   │  └───────────────────────┘  │         │
│  │  Design           │  │   └──────────────┬──────────────┘         │
│  └───────────────────┘  │                  │                        │
└─────────────────────────┘                  ▼                        │
                            ┌─────────────────────────────┐           │
                            │   ACCESS PATTERN WORKER     │           │
                            │        (Lambda)             │           │
                            │  ┌───────────────────────┐  │           │
                            │  │  Update PROCESSING_   │  │           │
                            │  │  PATTERNS             │  │           │
                            │  │  Invoke Bedrock       │  │           │
                            │  │  Merge & Save         │  │           │
                            │  │  Update COMPLETED     │  │           │
                            │  └───────────────────────┘  │           │
                            └──────────────┬──────────────┘           │
                                           │                          │
              ┌────────────────────────────┴──────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 DYNAMODB                                    │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Table: schemas                                                       │  │
│  │                                                                       │  │
│  │  Primary Key: conversionId (S)                                        │  │
│  │  GSI: byDate (conversionDate)                                         │  │
│  │  TTL: expiresAt (24 horas)                                            │  │
│  │                                                                       │  │
│  │  Attributes:                                                          │  │
│  │  - status, sqlContent, noSqlSchema, optimizationType, etc.            │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           ERROR HANDLING                                    │
│                                                                             │
│  ┌───────────────────────┐         ┌─────────────────────────────────────┐  │
│  │  conversion-dlq       │────────►│       DLQ HANDLER (Lambda)          │  │
│  │  (Dead Letter Queue)  │         │  Mark as FAILED after 3 retries     │  │
│  └───────────────────────┘         └─────────────────────────────────────┘  │
│                                                                             │
│  ┌───────────────────────┐         ┌─────────────────────────────────────┐  │
│  │  access-pattern-dlq   │────────►│       DLQ HANDLER (Lambda)          │  │
│  │  (Dead Letter Queue)  │         │  Mark as PATTERNS_FAILED            │  │
│  └───────────────────────┘         └─────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           OBSERVABILITY                                     │
│                                                                             │
│  ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────────┐   │
│  │  CloudWatch Logs  │  │  CloudWatch       │  │  CloudWatch Alarms    │   │
│  │  - Lambda logs    │  │  Metrics          │  │  - Error rate         │   │
│  │  - API Gateway    │  │  - Invocations    │  │  - Duration           │   │
│  │  - VPC Flow       │  │  - Duration       │  │  - DLQ messages       │   │
│  └───────────────────┘  │  - Errors         │  │  - Throttling         │   │
│                         └───────────────────┘  └───────────┬───────────┘   │
│                                                            │               │
│                                                            ▼               │
│                                              ┌─────────────────────────┐   │
│                                              │  SNS Topic (Alerts)     │   │
│                                              │  → Email notifications  │   │
│                                              └─────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Flujo de Datos

### Flujo Completo de Conversión

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   1. USUARIO ENVÍA SQL                                                       │
│   ═══════════════════                                                        │
│                                                                              │
│   POST /api/v1/schemas                                                       │
│   {                                                                          │
│     "sqlContent": "CREATE TABLE users (...)",                                │
│     "optimizationType": "balanced"                                           │
│   }                                                                          │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   2. PROCESS HANDLER                                                         │
│   ══════════════════                                                         │
│                                                                              │
│   a) Valida sintaxis SQL (regex, tipos de datos)                             │
│   b) Extrae información de tablas y columnas                                 │
│   c) Genera conversionId (UUID)                                              │
│   d) Crea registro en DynamoDB:                                              │
│      {                                                                       │
│        conversionId: "abc-123",                                              │
│        status: "PENDING",                                                    │
│        createdAt: 1705320000,                                                │
│        expiresAt: 1705406400  // +24h                                        │
│      }                                                                       │
│   e) Encola mensaje en SQS                                                   │
│   f) Retorna 202 Accepted                                                    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   3. SQS CONVERSION QUEUE                                                    │
│   ═══════════════════════                                                    │
│                                                                              │
│   Message:                                                                   │
│   {                                                                          │
│     "conversionId": "abc-123",                                               │
│     "sqlContent": "CREATE TABLE...",                                         │
│     "optimizationType": "balanced"                                           │
│   }                                                                          │
│                                                                              │
│   Config:                                                                    │
│   - Visibility Timeout: 180s (3 min)                                         │
│   - Max Receive Count: 3 (before DLQ)                                        │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   4. CONVERSION WORKER                                                       │
│   ════════════════════                                                       │
│                                                                              │
│   a) Update status → PROCESSING                                              │
│   b) Construye prompt para Claude:                                           │
│      "Convert SQL to single-table DynamoDB design..."                        │
│   c) Invoca Bedrock (Claude 3.5 Sonnet v2)                                   │
│   d) Recibe diseño JSON:                                                     │
│      {                                                                       │
│        "design": {                                                           │
│          "tableName": "MyAppTable",                                          │
│          "primaryKey": { "pk": "S", "sk": "S" },                             │
│          "gsi": [...]                                                        │
│        }                                                                     │
│      }                                                                       │
│   e) Valida diseño (estructura, campos requeridos)                           │
│   f) Update status → DESIGN_COMPLETED                                        │
│   g) Guarda diseño en noSqlSchema                                            │
│   h) Encola mensaje para access patterns                                     │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   5. SQS ACCESS PATTERN QUEUE                                                │
│   ═══════════════════════════                                                │
│                                                                              │
│   Message:                                                                   │
│   {                                                                          │
│     "conversionId": "abc-123",                                               │
│     "sqlContent": "...",                                                     │
│     "dynamoDBDesign": { ... }                                                │
│   }                                                                          │
│                                                                              │
│   Config:                                                                    │
│   - Visibility Timeout: 120s (2 min)                                         │
│   - Max Receive Count: 3                                                     │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   6. ACCESS PATTERN WORKER                                                   │
│   ════════════════════════                                                   │
│                                                                              │
│   a) Update status → PROCESSING_PATTERNS                                     │
│   b) Construye prompt con SQL + diseño                                       │
│   c) Invoca Bedrock (Claude 3.5 Haiku - más rápido)                          │
│   d) Recibe access patterns:                                                 │
│      [                                                                       │
│        {                                                                     │
│          "patternId": "AP1",                                                 │
│          "description": "Get user by ID",                                    │
│          "operation": "GetItem",                                             │
│          "keyCondition": "pk = USER#<id>"                                    │
│        }                                                                     │
│      ]                                                                       │
│   e) Mergea patterns con diseño existente                                    │
│   f) Update status → COMPLETED                                               │
│   g) Guarda resultado final                                                  │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   7. USUARIO CONSULTA RESULTADO                                              │
│   ═════════════════════════════                                              │
│                                                                              │
│   GET /api/v1/schemas/abc-123                                                │
│                                                                              │
│   Response:                                                                  │
│   {                                                                          │
│     "conversionId": "abc-123",                                               │
│     "status": "COMPLETED",                                                   │
│     "sqlContent": "CREATE TABLE...",                                         │
│     "noSqlSchema": {                                                         │
│       "design": {...},                                                       │
│       "accessPatternImplementation": [...],                                  │
│       "sampleData": [...]                                                    │
│     }                                                                        │
│   }                                                                          │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Estados del Sistema

```
                 ┌─────────┐
                 │ PENDING │
                 └────┬────┘
                      │ Process Handler crea registro
                      ▼
              ┌──────────────┐
              │  PROCESSING  │
              └──────┬───────┘
                     │ Conversion Worker procesa
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            │
   ┌─────────┐  ┌─────────────────┴──┐
   │  FAILED │  │  DESIGN_COMPLETED  │
   └─────────┘  └─────────┬──────────┘
                          │ Access Pattern Worker procesa
                          ▼
              ┌────────────────────────┐
              │  PROCESSING_PATTERNS   │
              └───────────┬────────────┘
                          │
         ┌────────────────┼────────────────┐
         │                │                │
         ▼                ▼                │
┌─────────────────┐  ┌───────────┐         │
│ PATTERNS_FAILED │  │ COMPLETED │         │
└─────────────────┘  └───────────┘         │
         │                                 │
         └─────────────────────────────────┘
                    (via DLQ Handler)
```

---

## Componentes del Sistema

### Frontend (Vue 3 SPA)

| Componente | Responsabilidad |
|------------|-----------------|
| HomeView | Entrada de SQL, selección de optimización |
| HistoryView | Lista de conversiones, detalle modal |
| Sidebar | Navegación entre vistas |
| schemasApi | Cliente HTTP para backend |

### Backend (Lambda Functions)

| Lambda | Trigger | Responsabilidad |
|--------|---------|-----------------|
| process-handler | API Gateway | Validar SQL, crear record, encolar |
| conversion-worker | SQS | Generar diseño con Bedrock |
| access-pattern-worker | SQS | Generar access patterns |
| query-handler | API Gateway | Consultar conversiones |
| dlq-handler | SQS DLQ | Marcar fallidos |

### Infraestructura

| Servicio | Propósito |
|----------|-----------|
| CloudFront | CDN, SSL, routing |
| API Gateway | HTTP API, CORS, throttling |
| S3 | Hosting SPA, logs |
| DynamoDB | Persistencia de datos |
| SQS | Mensajería asíncrona |
| Bedrock | IA generativa |
| CloudWatch | Logs, métricas, alarmas |
| SNS | Notificaciones |
| ACM | Certificados SSL |

---

## Patrones de Diseño

### 1. Event-Driven Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Producer      │────►│      Queue      │────►│   Consumer      │
│  (API Handler)  │     │    (SQS)        │     │   (Worker)      │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

**Beneficios:**
- Desacoplamiento entre componentes
- Procesamiento asíncrono
- Manejo de picos de carga
- Reintentos automáticos

### 2. Saga Pattern (Simplified)

```
┌──────────────────────────────────────────────────────────────────┐
│                        SAGA: Conversión                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Step 1: Create Record (PENDING)                                 │
│     ↓                                                            │
│  Step 2: Generate Design (PROCESSING → DESIGN_COMPLETED)         │
│     ↓                                                            │
│  Step 3: Generate Patterns (PROCESSING_PATTERNS → COMPLETED)     │
│                                                                  │
│  Compensación: DLQ Handler marca como FAILED                     │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 3. Single-Table Design (DynamoDB)

```
┌─────────────────────────────────────────────────────────────────┐
│  Tabla: schemas                                                 │
├─────────────────┬───────────────────────────────────────────────┤
│  PK             │  conversionId (UUID)                          │
│  GSI1-PK        │  conversionDate (YYYY-MM-DD)                  │
│  TTL            │  expiresAt (24h auto-delete)                  │
└─────────────────┴───────────────────────────────────────────────┘
```

### 4. API Gateway Pattern

```
┌─────────────────────────────────────────────────────────────────┐
│                      API GATEWAY                                │
├─────────────────────────────────────────────────────────────────┤
│  - Routing                                                      │
│  - CORS                                                         │
│  - Throttling                                                   │
│  - Request validation                                           │
│  - Authentication (optional)                                    │
│  - Logging                                                      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Seguridad

### Capas de Seguridad

```
                     ┌─────────────────────────────────────┐
                     │         Cloudflare WAF             │  Layer 1
                     │  - DDoS protection                 │
                     │  - Bot mitigation                  │
                     │  - Rate limiting                   │
                     └─────────────────────────────────────┘
                                      │
                     ┌─────────────────────────────────────┐
                     │       CloudFront + ACM             │  Layer 2
                     │  - TLS 1.2+ encryption             │
                     │  - Custom header validation        │
                     │  - Origin access control           │
                     └─────────────────────────────────────┘
                                      │
                     ┌─────────────────────────────────────┐
                     │       API Gateway                  │  Layer 3
                     │  - Throttling (rate limit)         │
                     │  - Request validation              │
                     │  - CORS configuration              │
                     └─────────────────────────────────────┘
                                      │
                     ┌─────────────────────────────────────┐
                     │       Lambda + IAM                 │  Layer 4
                     │  - Least privilege roles           │
                     │  - Resource-based policies         │
                     │  - VPC (optional)                  │
                     └─────────────────────────────────────┘
                                      │
                     ┌─────────────────────────────────────┐
                     │     DynamoDB + SQS                 │  Layer 5
                     │  - Encryption at rest              │
                     │  - IAM access control              │
                     │  - SSE-SQS encryption              │
                     └─────────────────────────────────────┘
```

### IAM Least Privilege

Cada Lambda tiene permisos mínimos:

| Lambda | DynamoDB | SQS | Bedrock |
|--------|----------|-----|---------|
| process-handler | PutItem | SendMessage | - |
| conversion-worker | UpdateItem | ReceiveMessage, DeleteMessage, SendMessage | InvokeModel |
| access-pattern-worker | UpdateItem | ReceiveMessage, DeleteMessage | InvokeModel |
| query-handler | GetItem, Query | - | - |
| dlq-handler | UpdateItem | ReceiveMessage, DeleteMessage | - |

### Origin Validation

```
Cloudflare → CloudFront → API Gateway

1. CloudFront Function valida header x-origin-secret
2. Solo requests de Cloudflare pasan
3. Requests directos a CloudFront son rechazados
```

---

## Escalabilidad

### Características de Auto-scaling

| Componente | Escalabilidad |
|------------|---------------|
| Lambda | Automático (1000+ concurrent) |
| API Gateway | Automático (10K+ RPS) |
| DynamoDB | On-demand (automático) |
| SQS | Ilimitado (messages) |
| CloudFront | Global edge network |

### Límites y Throttling

```
API Gateway:
- Burst: 5000 requests
- Rate: 10000 requests/second

Lambda:
- Concurrent executions: 1000 (default)
- Timeout: 30s (process), 180s (workers)

DynamoDB:
- On-demand: scales automatically
- Read/Write: unlimited (pay per request)

SQS:
- In-flight messages: 120,000
- Message size: 256 KB
```

### Bottlenecks Potenciales

1. **Bedrock Rate Limits**: Tokens per minute por modelo
2. **Lambda Cold Starts**: ~100ms para Go
3. **DynamoDB Hot Partitions**: Mitigado con UUID como PK

---

## Manejo de Errores

### Estrategia de Reintentos

```
┌─────────────────────────────────────────────────────────────────┐
│                    RETRY STRATEGY                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  SQS Message Processing:                                        │
│  ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐   │
│  │ Try 1   │────►│ Try 2   │────►│ Try 3   │────►│  DLQ    │   │
│  │         │     │         │     │         │     │         │   │
│  │ Fail    │     │ Fail    │     │ Fail    │     │ Handler │   │
│  └─────────┘     └─────────┘     └─────────┘     └─────────┘   │
│       │              │              │                 │         │
│       ▼              ▼              ▼                 ▼         │
│   Visibility     Visibility     Visibility       FAILED       │
│   Timeout        Timeout        Timeout          Status       │
│   (180s)         (180s)         (180s)                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Tipos de Errores

| Tipo | Acción | Ejemplo |
|------|--------|---------|
| Transient | Reintento automático | Timeout de red |
| Bedrock Error | No reintentar | Rate limit, invalid response |
| Validation Error | No reintentar | SQL inválido |
| Fatal | DLQ → FAILED | 3 reintentos fallidos |

### Alarmas y Notificaciones

```
CloudWatch Alarm → SNS Topic → Email

Alarmas configuradas:
- Lambda Error Rate > threshold
- Lambda Duration > timeout * 0.8
- SQS DLQ Messages > 0
- SQS Message Age > 1 hour
- DynamoDB Throttled Requests > 0
```

---

## Costos

### Estimación Mensual (1000 conversiones/mes)

| Servicio | Uso | Costo Estimado |
|----------|-----|----------------|
| Lambda | 5000 invocaciones × 5 funciones | $0.10 |
| API Gateway | 3000 requests | $0.01 |
| DynamoDB | 1000 writes, 3000 reads | $0.05 |
| SQS | 5000 messages | $0.01 |
| Bedrock | 2000 invocaciones Claude | $5-20 |
| CloudFront | 10GB transfer | $0.85 |
| S3 | 100MB storage | $0.02 |
| **Total** | | **~$10-25/mes** |

### Optimizaciones de Costo

1. **ARM64 (Graviton2)**: 20% más barato que x86
2. **On-demand DynamoDB**: Sin provisioning
3. **TTL 24h**: Datos eliminados automáticamente
4. **Haiku para patterns**: Más barato que Sonnet
5. **CloudFront caching**: Reduce requests a origen

### Cost Alerts

```hcl
resource "aws_budgets_budget" "monthly" {
  name         = "sql-to-nosql-monthly"
  budget_type  = "COST"
  limit_amount = "50"
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  notification {
    comparison_operator = "GREATER_THAN"
    threshold           = 80
    threshold_type      = "PERCENTAGE"
    notification_type   = "ACTUAL"
    subscriber_email_addresses = ["admin@example.com"]
  }
}
```
