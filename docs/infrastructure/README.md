# Infraestructura Terraform

Documentación detallada de la infraestructura como código del proyecto SQL to DynamoDB Converter.

## Tabla de Contenidos

1. [Visión General](#visión-general)
2. [Estructura de Módulos](#estructura-de-módulos)
3. [Ambientes](#ambientes)
4. [Módulo Lambda](#módulo-lambda)
5. [Módulo API Gateway](#módulo-api-gateway)
6. [Módulo DynamoDB](#módulo-dynamodb)
7. [Módulo SQS](#módulo-sqs)
8. [Módulo S3](#módulo-s3)
9. [Módulo CloudFront](#módulo-cloudfront)
10. [Módulo Bedrock](#módulo-bedrock)
11. [Gestión de IAM](#gestión-de-iam)
12. [Backend Remoto](#backend-remoto)
13. [Monitoreo y Alarmas](#monitoreo-y-alarmas)
14. [Decisiones de Arquitectura](#decisiones-de-arquitectura)

---

## Visión General

La infraestructura está diseñada con los siguientes principios:

- **Modularidad**: 8 módulos reutilizables e independientes
- **Multi-ambiente**: Desarrollo (LocalStack) y Producción (AWS)
- **Seguridad**: Least privilege IAM, encriptación en tránsito y reposo
- **Observabilidad**: CloudWatch Logs, Metrics y Alarms
- **Costo optimizado**: ARM64, on-demand DynamoDB, TTL automático

### Diagrama de Recursos

```
                         ┌─────────────────┐
                         │   Cloudflare    │
                         │   (DNS + WAF)   │
                         └────────┬────────┘
                                  │
                         ┌────────▼────────┐
                         │   CloudFront    │
                         │   Distribution  │
                         └────────┬────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
     ┌────────▼────────┐ ┌────────▼────────┐ ┌────────▼────────┐
     │   S3 Bucket     │ │  API Gateway    │ │  CloudFront     │
     │   (Frontend)    │ │  HTTP v2        │ │  Function       │
     └─────────────────┘ └────────┬────────┘ │  (Security)     │
                                  │          └─────────────────┘
              ┌───────────────────┼───────────────────┐
              │                   │                   │
     ┌────────▼────────┐ ┌────────▼────────┐ ┌────────▼────────┐
     │   Process       │ │   Query         │ │   DLQ           │
     │   Handler       │ │   Handler       │ │   Handler       │
     └────────┬────────┘ └────────┬────────┘ └─────────────────┘
              │                   │
     ┌────────▼────────┐ ┌────────▼────────┐
     │   SQS Queue     │ │   DynamoDB      │
     │   (Conversion)  │ │   (Schemas)     │
     └────────┬────────┘ └─────────────────┘
              │
     ┌────────▼────────┐
     │   Conversion    │
     │   Worker        │
     └────────┬────────┘
              │
     ┌────────▼────────┐
     │   Bedrock       │
     │   (Claude AI)   │
     └────────┬────────┘
              │
     ┌────────▼────────┐
     │   SQS Queue     │
     │   (Access Pat.) │
     └────────┬────────┘
              │
     ┌────────▼────────┐
     │   Access Pat.   │
     │   Worker        │
     └─────────────────┘
```

---

## Estructura de Módulos

```
infra/terraform/
├── modules/
│   ├── lambda/           # Template genérico para Lambda
│   ├── gateway/
│   │   ├── http-v2/      # API Gateway HTTP v2 (prod)
│   │   ├── rest-v1/      # API Gateway REST v1 (dev)
│   │   └── wrapper/      # Abstracción para ambos
│   ├── dynamodb/         # Tabla con TTL y GSI
│   ├── sqs/              # Colas + DLQ
│   ├── s3/               # Bucket con encryption
│   ├── cloudfront/       # CDN + OAC + ACM
│   └── bedrock/          # Acceso a modelos IA + políticas
│
├── environments/
│   ├── dev/              # LocalStack
│   │   ├── main.tf
│   │   └── variables.tf
│   └── prod/             # AWS Real
│       ├── main.tf
│       ├── variables.tf
│       ├── components/   # Lambdas individuales
│       └── shared/       # Recursos compartidos
│
└── backend/              # S3 backend config
    └── main.tf
```

---

## Ambientes

### Desarrollo (LocalStack)

```hcl
# environments/dev/main.tf
provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    apigateway     = "http://localhost:4566"
    dynamodb       = "http://localhost:4566"
    lambda         = "http://localhost:4566"
    sqs            = "http://localhost:4566"
    s3             = "http://s3.localhost.localstack.cloud:4566"
  }
}
```

**Características:**
- LocalStack Pro (emulación AWS completa)
- API Gateway REST v1 (más compatible)
- Sin costos de AWS
- Desarrollo rápido e iterativo

### Producción (AWS)

```hcl
# environments/prod/main.tf
terraform {
  backend "s3" {
    bucket       = "sql-to-nosql-terraform-state"
    key          = "prod/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
  }
}

provider "aws" {
  region = "us-east-1"
}
```

**Características:**
- AWS real con state remoto en S3
- API Gateway HTTP v2 (más eficiente)
- CloudFront + ACM para HTTPS
- Monitoreo completo con CloudWatch

---

## Módulo Lambda

**Ubicación:** `modules/lambda/`

### Propósito

Provisiona funciones Lambda con configuración estandarizada:
- Runtime Go 1.x en arquitectura ARM64
- CloudWatch Logs con retención configurable
- Alarmas de errores y duración
- Variables de entorno configurables

### Variables de Entrada

```hcl
variable "function_name" {
  description = "Nombre de la función Lambda"
  type        = string
}

variable "handler" {
  description = "Nombre del handler (default: bootstrap)"
  type        = string
  default     = "bootstrap"
}

variable "runtime" {
  description = "Runtime de Lambda"
  type        = string
  default     = "provided.al2"  # Go custom runtime
}

variable "architecture" {
  description = "Arquitectura del procesador"
  type        = string
  default     = "arm64"  # Graviton2 - más económico
}

variable "memory_size" {
  description = "Memoria asignada en MB"
  type        = number
  default     = 128
}

variable "timeout" {
  description = "Timeout en segundos"
  type        = number
  default     = 30
}

variable "environment_variables" {
  description = "Variables de entorno"
  type        = map(string)
  default     = {}
}

variable "source_path" {
  description = "Path al archivo ZIP del código"
  type        = string
}

variable "role_arn" {
  description = "ARN del rol IAM"
  type        = string
}
```

### Recursos Creados

```hcl
# Función Lambda
resource "aws_lambda_function" "this" {
  function_name    = var.function_name
  role             = var.role_arn
  handler          = var.handler
  runtime          = var.runtime
  architectures    = [var.architecture]
  memory_size      = var.memory_size
  timeout          = var.timeout
  filename         = var.source_path
  source_code_hash = filebase64sha256(var.source_path)

  environment {
    variables = var.environment_variables
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "this" {
  name              = "/aws/lambda/${var.function_name}"
  retention_in_days = var.log_retention_days
}

# Alarma de errores
resource "aws_cloudwatch_metric_alarm" "errors" {
  alarm_name          = "${var.function_name}-high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = var.error_threshold
  alarm_actions       = [var.sns_topic_arn]

  dimensions = {
    FunctionName = aws_lambda_function.this.function_name
  }
}

# Alarma de duración
resource "aws_cloudwatch_metric_alarm" "duration" {
  alarm_name          = "${var.function_name}-high-duration"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Average"
  threshold           = var.duration_threshold
  alarm_actions       = [var.sns_topic_arn]

  dimensions = {
    FunctionName = aws_lambda_function.this.function_name
  }
}
```

### Outputs

```hcl
output "function_name" {
  value = aws_lambda_function.this.function_name
}

output "function_arn" {
  value = aws_lambda_function.this.arn
}

output "invoke_arn" {
  value = aws_lambda_function.this.invoke_arn
}
```

---

## Módulo API Gateway

**Ubicación:** `modules/gateway/`

### Estructura

```
gateway/
├── http-v2/     # API Gateway HTTP v2 (producción)
├── rest-v1/     # API Gateway REST v1 (desarrollo)
└── wrapper/     # Abstracción unificada
```

### HTTP v2 (Producción)

API Gateway HTTP v2 es más eficiente y económico:

```hcl
# API principal
resource "aws_apigatewayv2_api" "this" {
  name          = var.api_name
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins     = var.cors_origins
    allow_methods     = ["GET", "POST", "OPTIONS"]
    allow_headers     = ["Content-Type", "Authorization"]
    expose_headers    = ["*"]
    max_age           = 300
  }
}

# Stage con auto-deploy
resource "aws_apigatewayv2_stage" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = var.stage_name
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = var.throttle_burst
    throttling_rate_limit  = var.throttle_rate
  }

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api.arn
    format          = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      responseLength = "$context.responseLength"
      integrationError = "$context.integrationErrorMessage"
    })
  }
}

# Integración Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  for_each = var.lambda_routes

  api_id                 = aws_apigatewayv2_api.this.id
  integration_type       = "AWS_PROXY"
  integration_uri        = each.value.invoke_arn
  payload_format_version = "2.0"
}

# Rutas
resource "aws_apigatewayv2_route" "routes" {
  for_each = var.lambda_routes

  api_id    = aws_apigatewayv2_api.this.id
  route_key = each.key  # Ej: "POST /api/v1/schemas"
  target    = "integrations/${aws_apigatewayv2_integration.lambda[each.key].id}"
}
```

### Rutas Configuradas

| Route Key | Lambda | Descripción |
|-----------|--------|-------------|
| `POST /api/v1/schemas` | process-handler | Crear conversión |
| `GET /api/v1/schemas` | query-handler | Listar conversiones |
| `GET /api/v1/schemas/{id}` | query-handler | Obtener por ID |

---

## Módulo DynamoDB

**Ubicación:** `modules/dynamodb/`

### Diseño de Tabla

```hcl
resource "aws_dynamodb_table" "schemas" {
  name         = var.table_name
  billing_mode = "PAY_PER_REQUEST"  # On-demand, sin provisioning
  hash_key     = "conversionId"

  # Atributos para keys e índices
  attribute {
    name = "conversionId"
    type = "S"
  }

  attribute {
    name = "conversionDate"
    type = "S"
  }

  # TTL - registros expiran automáticamente
  ttl {
    attribute_name = "expiresAt"
    enabled        = true
  }

  # GSI para listar por fecha
  global_secondary_index {
    name            = "byDate"
    hash_key        = "conversionDate"
    projection_type = "ALL"
  }

  # Encriptación
  server_side_encryption {
    enabled = true
  }

  # Point-in-time recovery
  point_in_time_recovery {
    enabled = var.enable_pitr
  }

  tags = var.tags
}

# Alarma de throttling
resource "aws_cloudwatch_metric_alarm" "throttle" {
  alarm_name          = "${var.table_name}-throttled-requests"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ThrottledRequests"
  namespace           = "AWS/DynamoDB"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_actions       = [var.sns_topic_arn]

  dimensions = {
    TableName = aws_dynamodb_table.schemas.name
  }
}
```

### Esquema de Datos

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Tabla: schemas                               │
├──────────────────┬──────────────────────────────────────────────────┤
│ conversionId (PK)│ UUID único de la conversión                      │
│ conversionDate   │ Fecha YYYY-MM-DD (para GSI)                      │
│ createdAt        │ Timestamp Unix de creación                       │
│ expiresAt        │ Timestamp Unix de expiración (TTL)               │
│ status           │ PENDING|PROCESSING|DESIGN_COMPLETED|...          │
│ sqlContent       │ SQL original del usuario                         │
│ optimizationType │ balanced|read-heavy|write-heavy|cost-optimized   │
│ noSqlSchema      │ JSON string con el diseño DynamoDB               │
│ tablesExtracted  │ Número de tablas extraídas del SQL               │
└──────────────────┴──────────────────────────────────────────────────┘
```

---

## Módulo SQS

**Ubicación:** `modules/sqs/`

### Arquitectura de Colas

```
┌───────────────────┐     ┌───────────────────┐
│ conversion_queue  │────►│ conversion_worker │
│ (visibility: 180s)│     │ Lambda            │
└───────────────────┘     └───────────────────┘
         │ (3 reintentos fallidos)
         ▼
┌───────────────────┐     ┌───────────────────┐
│ conversion_dlq    │────►│ dlq_handler       │
│ (Dead Letter)     │     │ Lambda            │
└───────────────────┘     └───────────────────┘

┌───────────────────────┐     ┌─────────────────────┐
│ access_pattern_queue  │────►│ access_pattern      │
│ (visibility: 120s)    │     │ worker Lambda       │
└───────────────────────┘     └─────────────────────┘
         │ (3 reintentos fallidos)
         ▼
┌───────────────────────┐     ┌───────────────────┐
│ access_pattern_dlq    │────►│ dlq_handler       │
└───────────────────────┘     │ Lambda            │
                              └───────────────────┘
```

### Configuración

```hcl
# Cola principal de conversión
resource "aws_sqs_queue" "conversion" {
  name                       = "${var.project_name}-conversion-queue"
  visibility_timeout_seconds = 180  # 3 minutos para procesar
  message_retention_seconds  = 86400  # 1 día
  receive_wait_time_seconds  = 20  # Long polling

  # Encriptación
  sqs_managed_sse_enabled = true

  # Redrive policy a DLQ
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.conversion_dlq.arn
    maxReceiveCount     = 3  # 3 intentos antes de DLQ
  })
}

# Dead Letter Queue
resource "aws_sqs_queue" "conversion_dlq" {
  name                      = "${var.project_name}-conversion-dlq"
  message_retention_seconds = 1209600  # 14 días
  sqs_managed_sse_enabled   = true
}

# Cola de access patterns
resource "aws_sqs_queue" "access_pattern" {
  name                       = "${var.project_name}-access-pattern-queue"
  visibility_timeout_seconds = 120  # 2 minutos
  message_retention_seconds  = 86400

  sqs_managed_sse_enabled = true

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.access_pattern_dlq.arn
    maxReceiveCount     = 3
  })
}

# DLQ de access patterns
resource "aws_sqs_queue" "access_pattern_dlq" {
  name                      = "${var.project_name}-access-pattern-dlq"
  message_retention_seconds = 1209600
  sqs_managed_sse_enabled   = true
}
```

### Alarmas

```hcl
# Alerta si hay mensajes en DLQ
resource "aws_cloudwatch_metric_alarm" "dlq_messages" {
  alarm_name          = "${var.project_name}-dlq-has-messages"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_actions       = [var.sns_topic_arn]

  dimensions = {
    QueueName = aws_sqs_queue.conversion_dlq.name
  }
}

# Alerta si mensajes están envejeciendo
resource "aws_cloudwatch_metric_alarm" "message_age" {
  alarm_name          = "${var.project_name}-queue-message-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Maximum"
  threshold           = 3600  # 1 hora
  alarm_actions       = [var.sns_topic_arn]

  dimensions = {
    QueueName = aws_sqs_queue.conversion.name
  }
}
```

---

## Módulo S3

**Ubicación:** `modules/s3/`

### Buckets

| Bucket | Propósito |
|--------|-----------|
| `sql-to-nosql-frontend` | Hosting SPA Vue |
| `sql-to-nosql-logs` | Access logs de S3 |
| `sql-to-nosql-terraform-state` | State remoto Terraform |

### Configuración Frontend Bucket

```hcl
resource "aws_s3_bucket" "frontend" {
  bucket = var.bucket_name
}

# Bloquear acceso público directo
resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Encriptación
resource "aws_s3_bucket_server_side_encryption_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Versionado
resource "aws_s3_bucket_versioning" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Access logging
resource "aws_s3_bucket_logging" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  target_bucket = var.logs_bucket_id
  target_prefix = "s3-access-logs/${var.bucket_name}/"
}

# Política para CloudFront OAC
resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFrontServicePrincipal"
        Effect    = "Allow"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.frontend.arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = var.cloudfront_distribution_arn
          }
        }
      }
    ]
  })
}
```

---

## Módulo CloudFront

**Ubicación:** `modules/cloudfront/`

### Arquitectura

```
                     ┌─────────────────────┐
                     │     Cloudflare      │
                     │  (DNS + Proxy)      │
                     └──────────┬──────────┘
                                │
                     ┌──────────▼──────────┐
                     │     CloudFront      │
                     │   Distribution      │
                     └──────────┬──────────┘
                                │
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
┌────────▼────────┐  ┌──────────▼──────────┐  ┌────────▼────────┐
│  Origin: S3     │  │  Origin: API GW     │  │  CloudFront     │
│  (Frontend)     │  │  (Backend)          │  │  Function       │
│  Path: /*       │  │  Path: /prod/*      │  │  (Validate      │
│  OAC enabled    │  │  x-origin-secret    │  │   x-origin)     │
└─────────────────┘  └─────────────────────┘  └─────────────────┘
```

### Configuración Completa

```hcl
# Origin Access Control para S3
resource "aws_cloudfront_origin_access_control" "s3" {
  name                              = "${var.project_name}-s3-oac"
  description                       = "OAC for frontend S3 bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# CloudFront Distribution
resource "aws_cloudfront_distribution" "main" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  price_class         = "PriceClass_100"  # US, Canada, Europe
  aliases             = [var.custom_domain]

  # Origin 1: S3 (Frontend)
  origin {
    domain_name              = var.s3_bucket_domain
    origin_id                = "S3-Frontend"
    origin_access_control_id = aws_cloudfront_origin_access_control.s3.id
  }

  # Origin 2: API Gateway (Backend)
  origin {
    domain_name = var.api_gateway_domain
    origin_id   = "API-Gateway"
    origin_path = ""

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }

    # Header secreto para validar origen
    custom_header {
      name  = "x-origin-secret"
      value = var.origin_secret
    }
  }

  # Comportamiento default: S3 (Frontend)
  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-Frontend"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400

    # CloudFront Function para validación
    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.security.arn
    }
  }

  # Comportamiento API: /prod/*
  ordered_cache_behavior {
    path_pattern     = "/prod/*"
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "API-Gateway"

    forwarded_values {
      query_string = true
      headers      = ["Authorization", "Origin", "Accept"]
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "https-only"
    min_ttl                = 0
    default_ttl            = 0  # Sin cache para API
    max_ttl                = 0
  }

  # Comportamiento para assets estáticos
  ordered_cache_behavior {
    path_pattern     = "/assets/*"
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "S3-Frontend"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 31536000  # 1 año
    default_ttl            = 31536000
    max_ttl                = 31536000
    compress               = true
  }

  # Restricciones geográficas (opcional)
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # Certificado SSL
  viewer_certificate {
    acm_certificate_arn      = var.acm_certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  # Error pages para SPA
  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  custom_error_response {
    error_code         = 403
    response_code      = 200
    response_page_path = "/index.html"
  }
}
```

### CloudFront Function (Seguridad)

```javascript
// Valida que las requests vengan de Cloudflare
function handler(event) {
    var request = event.request;
    var headers = request.headers;

    // Verificar header secreto de origen
    var originSecret = headers['x-origin-verify'];

    if (!originSecret || originSecret.value !== 'EXPECTED_SECRET') {
        return {
            statusCode: 403,
            statusDescription: 'Forbidden',
            body: {
                encoding: 'text',
                data: 'Access denied'
            }
        };
    }

    return request;
}
```

---

## Módulo Bedrock

**Ubicación:** `modules/bedrock/`

### Modelos Utilizados

| Modelo | ID | Uso |
|--------|-----|-----|
| Claude 3.5 Sonnet v2 | `us.anthropic.claude-3-5-sonnet-20241022-v2:0` | Conversión SQL→DynamoDB |
| Claude 3.5 Haiku | `us.anthropic.claude-3-5-haiku-20241022-v1:0` | Generación access patterns |

### Políticas IAM

```hcl
# Política para invocar Bedrock
resource "aws_iam_policy" "bedrock_invoke" {
  name        = "${var.project_name}-bedrock-invoke"
  description = "Allow Lambda to invoke Bedrock models"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Resource = [
          "arn:aws:bedrock:${var.region}::foundation-model/anthropic.claude-3-5-sonnet-20241022-v2:0",
          "arn:aws:bedrock:${var.region}::foundation-model/anthropic.claude-3-5-haiku-20241022-v1:0",
          "arn:aws:bedrock:us.*:${data.aws_caller_identity.current.account_id}:inference-profile/us.anthropic.*"
        ]
      }
    ]
  })
}

# Attach a Lambda roles
resource "aws_iam_role_policy_attachment" "conversion_worker_bedrock" {
  role       = var.conversion_worker_role_name
  policy_arn = aws_iam_policy.bedrock_invoke.arn
}

resource "aws_iam_role_policy_attachment" "access_pattern_worker_bedrock" {
  role       = var.access_pattern_worker_role_name
  policy_arn = aws_iam_policy.bedrock_invoke.arn
}
```

---

## Gestión de IAM

### Decisión de Arquitectura

**No existe un módulo IAM centralizado.** Las políticas IAM se definen junto a los recursos que protegen, siguiendo el patrón de "colocación" recomendado por [AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/structure.html).

> Ver [ADR-001: Colocación de Políticas IAM](../architecture/decisions/001-iam-policy-colocation.md) para la justificación completa.

### Estructura Actual

```
environments/prod/
├── main.tf                    # Roles IAM base (5 roles, uno por Lambda)
└── shared/
    ├── dynamodb.tf            # Políticas DynamoDB por Lambda
    ├── sqs.tf                 # Políticas SQS por Lambda
    └── ...

modules/bedrock/
└── main.tf                    # Políticas Bedrock (caso especial)
```

### Principio de Menor Privilegio

Cada Lambda tiene su propio rol con permisos mínimos:

```hcl
# environments/prod/main.tf - Roles base
resource "aws_iam_role" "process_handler" {
  name               = "sql-to-nosql-prod-process-handler-role"
  assume_role_policy = local.lambda_assume_role_policy
}

# environments/prod/shared/dynamodb.tf - Políticas específicas
resource "aws_iam_role_policy" "process_handler_dynamodb" {
  name = "prod-process-handler-dynamodb"
  role = var.process_handler_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["dynamodb:PutItem"]  # Solo lo que necesita
      Resource = [module.schemas_table.table_arn]
    }]
  })
}
```

### Matriz de Permisos por Lambda

| Lambda | DynamoDB | SQS | Bedrock |
|--------|----------|-----|---------|
| process-handler | PutItem | SendMessage | - |
| conversion-worker | UpdateItem | ReceiveMessage, DeleteMessage, SendMessage | InvokeModel |
| access-pattern-worker | UpdateItem | ReceiveMessage, DeleteMessage | InvokeModel |
| query-handler | GetItem, Query, Scan | - | - |
| dlq-handler | UpdateItem | ReceiveMessage, DeleteMessage | - |

### Beneficios de este Enfoque

1. **Auditoría fácil**: Abrir `dynamodb.tf` muestra todos los permisos de DynamoDB
2. **Cambios atómicos**: Modificar tabla + políticas en el mismo archivo
3. **Least privilege visible**: Cada política está junto al recurso que protege
4. **Sin thin wrappers**: Evita módulos que solo envuelven `aws_iam_role`

---

## Backend Remoto

**Ubicación:** `backend/`

### Configuración S3

```hcl
# backend/main.tf
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# Bucket para state
resource "aws_s3_bucket" "terraform_state" {
  bucket = "sql-to-nosql-terraform-state"
}

# Bloquear acceso público
resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Encriptación
resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

# Versionado
resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}
```

### Uso en Producción

```hcl
# environments/prod/main.tf
terraform {
  backend "s3" {
    bucket       = "sql-to-nosql-terraform-state"
    key          = "prod/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true  # File-based locking
  }
}
```

---

## Monitoreo y Alarmas

### Alarmas Configuradas

| Servicio | Alarma | Umbral | Acción |
|----------|--------|--------|--------|
| Lambda | High Error Rate | > N errores/min | SNS |
| Lambda | High Duration | > N ms promedio | SNS |
| DynamoDB | Throttled Requests | > 0 | SNS |
| SQS | DLQ Has Messages | > 0 | SNS |
| SQS | Message Age | > 1 hora | SNS |

### SNS Topic

```hcl
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-alerts"
}

resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}
```

### CloudWatch Dashboard (opcional)

```hcl
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.project_name}-dashboard"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title   = "Lambda Invocations"
          metrics = [
            ["AWS/Lambda", "Invocations", "FunctionName", "process-handler"],
            ["...", "conversion-worker"],
            ["...", "access-pattern-worker"],
            ["...", "query-handler"]
          ]
          period = 60
          stat   = "Sum"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title   = "Lambda Errors"
          metrics = [
            ["AWS/Lambda", "Errors", "FunctionName", "process-handler"],
            ["...", "conversion-worker"],
            ["...", "access-pattern-worker"]
          ]
          period = 60
          stat   = "Sum"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          title   = "SQS Messages"
          metrics = [
            ["AWS/SQS", "ApproximateNumberOfMessagesVisible", "QueueName", "sql-to-nosql-conversion-queue"],
            ["...", "sql-to-nosql-access-pattern-queue"]
          ]
          period = 60
          stat   = "Sum"
        }
      }
    ]
  })
}
```

---

## Comandos Útiles

```bash
# Inicializar Terraform
cd infra/terraform/environments/prod
terraform init

# Planificar cambios
terraform plan -var-file="terraform.tfvars"

# Aplicar cambios
terraform apply -var-file="terraform.tfvars"

# Ver estado actual
terraform state list

# Importar recurso existente
terraform import aws_s3_bucket.example my-bucket-name

# Destruir (con cuidado!)
terraform destroy -var-file="terraform.tfvars"

# Formatear código
terraform fmt -recursive

# Validar sintaxis
terraform validate
```

---

## Decisiones de Arquitectura

Las decisiones de diseño importantes están documentadas como ADRs (Architecture Decision Records):

| ADR | Título | Estado |
|-----|--------|--------|
| [ADR-001](../architecture/decisions/001-iam-policy-colocation.md) | Colocación de Políticas IAM junto a Recursos | Aceptado |

---

## Próximos Pasos Sugeridos

1. **WAF**: Agregar AWS WAF a CloudFront para protección adicional
2. **X-Ray**: Habilitar tracing distribuido para debugging
3. **Secrets Manager**: Migrar secrets de variables de entorno
4. **Auto-scaling**: Configurar provisioned capacity para DynamoDB en alta demanda
5. **Multi-region**: Replicación DynamoDB global para DR
