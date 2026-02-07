# Módulos Terraform - Guía Detallada

Documentación exhaustiva de cada módulo Terraform del proyecto.

## Índice de Módulos

| Módulo | Ubicación | Propósito |
|--------|-----------|-----------|
| [lambda](#lambda) | `modules/lambda/` | Funciones Lambda |
| [gateway/http-v2](#gateway-http-v2) | `modules/gateway/http-v2/` | API Gateway HTTP v2 |
| [gateway/rest-v1](#gateway-rest-v1) | `modules/gateway/rest-v1/` | API Gateway REST v1 |
| [gateway/wrapper](#gateway-wrapper) | `modules/gateway/wrapper/` | Abstracción de API Gateway |
| [dynamodb](#dynamodb) | `modules/dynamodb/` | Tabla DynamoDB |
| [sqs](#sqs) | `modules/sqs/` | Colas de mensajes |
| [s3](#s3) | `modules/s3/` | Buckets S3 |
| [cloudfront](#cloudfront) | `modules/cloudfront/` | CDN y distribución |
| [bedrock](#bedrock) | `modules/bedrock/` | Acceso a modelos IA |

> **Nota:** No existe un módulo IAM centralizado. Las políticas IAM se definen junto a los recursos que protegen. Ver [ADR-001](../architecture/decisions/001-iam-policy-colocation.md).

---

## Lambda

**Path:** `modules/lambda/`

### Descripción

Módulo genérico para crear funciones Lambda con configuración estandarizada:
- Runtime Go en ARM64
- CloudWatch Logs con retención configurable
- Alarmas de monitoreo
- Event source mappings opcionales

### Archivos

```
modules/lambda/
├── main.tf           # Recursos principales
├── variables.tf      # Variables de entrada
├── outputs.tf        # Outputs del módulo
└── alarms.tf         # Alarmas CloudWatch
```

### Variables

```hcl
# variables.tf

variable "function_name" {
  description = "Nombre único de la función Lambda"
  type        = string
}

variable "description" {
  description = "Descripción de la función"
  type        = string
  default     = ""
}

variable "handler" {
  description = "Nombre del handler (default: bootstrap para Go)"
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
  default     = "arm64"  # Graviton2 - 20% más económico
  validation {
    condition     = contains(["arm64", "x86_64"], var.architecture)
    error_message = "Architecture must be arm64 or x86_64"
  }
}

variable "memory_size" {
  description = "Memoria asignada en MB (128-10240)"
  type        = number
  default     = 128
  validation {
    condition     = var.memory_size >= 128 && var.memory_size <= 10240
    error_message = "Memory must be between 128 and 10240 MB"
  }
}

variable "timeout" {
  description = "Timeout en segundos (1-900)"
  type        = number
  default     = 30
  validation {
    condition     = var.timeout >= 1 && var.timeout <= 900
    error_message = "Timeout must be between 1 and 900 seconds"
  }
}

variable "source_path" {
  description = "Path al archivo ZIP del código"
  type        = string
}

variable "role_arn" {
  description = "ARN del rol IAM para la función"
  type        = string
}

variable "environment_variables" {
  description = "Variables de entorno"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "reserved_concurrent_executions" {
  description = "Concurrencia reservada (-1 para unreserved)"
  type        = number
  default     = -1
}

variable "log_retention_days" {
  description = "Días de retención de logs"
  type        = number
  default     = 14
}

variable "enable_xray" {
  description = "Habilitar X-Ray tracing"
  type        = bool
  default     = false
}

variable "vpc_config" {
  description = "Configuración de VPC (opcional)"
  type = object({
    subnet_ids         = list(string)
    security_group_ids = list(string)
  })
  default = null
}

# Event source mapping (para SQS)
variable "sqs_event_source" {
  description = "Configuración de SQS event source"
  type = object({
    queue_arn        = string
    batch_size       = number
    max_batching_window = number
  })
  default = null
}

# Alarmas
variable "alarm_sns_topic_arn" {
  description = "ARN del SNS topic para alarmas"
  type        = string
  default     = ""
}

variable "error_threshold" {
  description = "Umbral de errores para alarma"
  type        = number
  default     = 1
}

variable "duration_threshold_ms" {
  description = "Umbral de duración en ms para alarma"
  type        = number
  default     = 25000  # 25 segundos
}

variable "tags" {
  description = "Tags para los recursos"
  type        = map(string)
  default     = {}
}
```

### Recursos

```hcl
# main.tf

# Función Lambda
resource "aws_lambda_function" "this" {
  function_name = var.function_name
  description   = var.description
  role          = var.role_arn
  handler       = var.handler
  runtime       = var.runtime
  architectures = [var.architecture]
  memory_size   = var.memory_size
  timeout       = var.timeout

  filename         = var.source_path
  source_code_hash = filebase64sha256(var.source_path)

  reserved_concurrent_executions = var.reserved_concurrent_executions

  dynamic "environment" {
    for_each = length(var.environment_variables) > 0 ? [1] : []
    content {
      variables = var.environment_variables
    }
  }

  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }

  dynamic "tracing_config" {
    for_each = var.enable_xray ? [1] : []
    content {
      mode = "Active"
    }
  }

  tags = var.tags
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "this" {
  name              = "/aws/lambda/${aws_lambda_function.this.function_name}"
  retention_in_days = var.log_retention_days
  tags              = var.tags
}

# SQS Event Source Mapping (opcional)
resource "aws_lambda_event_source_mapping" "sqs" {
  count = var.sqs_event_source != null ? 1 : 0

  event_source_arn                   = var.sqs_event_source.queue_arn
  function_name                      = aws_lambda_function.this.arn
  batch_size                         = var.sqs_event_source.batch_size
  maximum_batching_window_in_seconds = var.sqs_event_source.max_batching_window
  enabled                            = true
}
```

### Alarmas

```hcl
# alarms.tf

# Alarma de errores
resource "aws_cloudwatch_metric_alarm" "errors" {
  count = var.alarm_sns_topic_arn != "" ? 1 : 0

  alarm_name          = "${var.function_name}-high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = var.error_threshold
  alarm_description   = "Lambda ${var.function_name} has high error rate"
  alarm_actions       = [var.alarm_sns_topic_arn]
  ok_actions          = [var.alarm_sns_topic_arn]

  dimensions = {
    FunctionName = aws_lambda_function.this.function_name
  }

  tags = var.tags
}

# Alarma de duración
resource "aws_cloudwatch_metric_alarm" "duration" {
  count = var.alarm_sns_topic_arn != "" ? 1 : 0

  alarm_name          = "${var.function_name}-high-duration"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Average"
  threshold           = var.duration_threshold_ms
  alarm_description   = "Lambda ${var.function_name} has high average duration"
  alarm_actions       = [var.alarm_sns_topic_arn]

  dimensions = {
    FunctionName = aws_lambda_function.this.function_name
  }

  tags = var.tags
}

# Alarma de throttling
resource "aws_cloudwatch_metric_alarm" "throttles" {
  count = var.alarm_sns_topic_arn != "" ? 1 : 0

  alarm_name          = "${var.function_name}-throttled"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Throttles"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Lambda ${var.function_name} is being throttled"
  alarm_actions       = [var.alarm_sns_topic_arn]

  dimensions = {
    FunctionName = aws_lambda_function.this.function_name
  }

  tags = var.tags
}
```

### Outputs

```hcl
# outputs.tf

output "function_name" {
  description = "Nombre de la función Lambda"
  value       = aws_lambda_function.this.function_name
}

output "function_arn" {
  description = "ARN de la función Lambda"
  value       = aws_lambda_function.this.arn
}

output "invoke_arn" {
  description = "ARN para invocar la función (API Gateway)"
  value       = aws_lambda_function.this.invoke_arn
}

output "qualified_arn" {
  description = "ARN calificado (con versión)"
  value       = aws_lambda_function.this.qualified_arn
}

output "log_group_name" {
  description = "Nombre del log group"
  value       = aws_cloudwatch_log_group.this.name
}
```

### Uso

```hcl
module "process_handler" {
  source = "../../modules/lambda"

  function_name = "sql-to-nosql-process-handler"
  description   = "Validates SQL and creates conversion records"
  source_path   = "${path.root}/../../../dist/diagrams.zip"
  role_arn      = module.iam_process_handler.role_arn
  timeout       = 30
  memory_size   = 256

  environment_variables = {
    DYNAMODB_TABLE_NAME = module.dynamodb.table_name
    SQS_QUEUE_URL       = module.sqs.conversion_queue_url
  }

  alarm_sns_topic_arn   = module.sns.topic_arn
  error_threshold       = 3
  duration_threshold_ms = 20000

  tags = local.common_tags
}
```

---

## Gateway HTTP v2

**Path:** `modules/gateway/http-v2/`

### Descripción

API Gateway HTTP v2 para producción. Más eficiente y económico que REST v1.

### Archivos

```
modules/gateway/http-v2/
├── main.tf           # API, stage, access logs
├── routes.tf         # Rutas e integraciones
├── variables.tf      # Variables de entrada
└── outputs.tf        # Outputs
```

### Variables

```hcl
variable "api_name" {
  description = "Nombre de la API"
  type        = string
}

variable "stage_name" {
  description = "Nombre del stage"
  type        = string
  default     = "prod"
}

variable "cors_configuration" {
  description = "Configuración CORS"
  type = object({
    allow_origins     = list(string)
    allow_methods     = list(string)
    allow_headers     = list(string)
    expose_headers    = list(string)
    max_age           = number
    allow_credentials = bool
  })
  default = {
    allow_origins     = ["*"]
    allow_methods     = ["GET", "POST", "OPTIONS"]
    allow_headers     = ["Content-Type", "Authorization"]
    expose_headers    = ["*"]
    max_age           = 300
    allow_credentials = false
  }
}

variable "routes" {
  description = "Mapa de rutas a configurar"
  type = map(object({
    method      = string
    path        = string
    lambda_arn  = string
    invoke_arn  = string
  }))
}

variable "throttling" {
  description = "Configuración de throttling"
  type = object({
    burst_limit = number
    rate_limit  = number
  })
  default = {
    burst_limit = 1000
    rate_limit  = 500
  }
}

variable "access_log_settings" {
  description = "Configuración de access logs"
  type = object({
    enabled = bool
    format  = string
  })
  default = {
    enabled = true
    format  = ""  # Usa formato default
  }
}

variable "tags" {
  type    = map(string)
  default = {}
}
```

### Recursos Principales

```hcl
# main.tf

# API Gateway HTTP v2
resource "aws_apigatewayv2_api" "this" {
  name          = var.api_name
  protocol_type = "HTTP"
  description   = "HTTP API for ${var.api_name}"

  cors_configuration {
    allow_origins     = var.cors_configuration.allow_origins
    allow_methods     = var.cors_configuration.allow_methods
    allow_headers     = var.cors_configuration.allow_headers
    expose_headers    = var.cors_configuration.expose_headers
    max_age           = var.cors_configuration.max_age
    allow_credentials = var.cors_configuration.allow_credentials
  }

  tags = var.tags
}

# CloudWatch Log Group para access logs
resource "aws_cloudwatch_log_group" "access_logs" {
  count = var.access_log_settings.enabled ? 1 : 0

  name              = "/aws/apigateway/${var.api_name}"
  retention_in_days = 14
  tags              = var.tags
}

# Stage con auto-deploy
resource "aws_apigatewayv2_stage" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = var.stage_name
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = var.throttling.burst_limit
    throttling_rate_limit  = var.throttling.rate_limit
  }

  dynamic "access_log_settings" {
    for_each = var.access_log_settings.enabled ? [1] : []
    content {
      destination_arn = aws_cloudwatch_log_group.access_logs[0].arn
      format = var.access_log_settings.format != "" ? var.access_log_settings.format : jsonencode({
        requestId          = "$context.requestId"
        ip                 = "$context.identity.sourceIp"
        requestTime        = "$context.requestTime"
        httpMethod         = "$context.httpMethod"
        routeKey           = "$context.routeKey"
        status             = "$context.status"
        protocol           = "$context.protocol"
        responseLength     = "$context.responseLength"
        integrationLatency = "$context.integrationLatency"
        integrationError   = "$context.integrationErrorMessage"
      })
    }
  }

  tags = var.tags
}
```

### Rutas

```hcl
# routes.tf

# Integraciones Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  for_each = var.routes

  api_id                 = aws_apigatewayv2_api.this.id
  integration_type       = "AWS_PROXY"
  integration_uri        = each.value.invoke_arn
  payload_format_version = "2.0"
  description            = "Lambda integration for ${each.key}"
}

# Rutas
resource "aws_apigatewayv2_route" "routes" {
  for_each = var.routes

  api_id    = aws_apigatewayv2_api.this.id
  route_key = "${each.value.method} ${each.value.path}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda[each.key].id}"
}

# Permisos para API Gateway invocar Lambda
resource "aws_lambda_permission" "api_gateway" {
  for_each = var.routes

  statement_id  = "AllowAPIGateway-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = each.value.lambda_arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.this.execution_arn}/*/*"
}
```

### Outputs

```hcl
output "api_id" {
  value = aws_apigatewayv2_api.this.id
}

output "api_endpoint" {
  value = aws_apigatewayv2_api.this.api_endpoint
}

output "stage_invoke_url" {
  value = aws_apigatewayv2_stage.this.invoke_url
}

output "execution_arn" {
  value = aws_apigatewayv2_api.this.execution_arn
}
```

---

## DynamoDB

**Path:** `modules/dynamodb/`

### Descripción

Tabla DynamoDB con:
- Billing on-demand (pay-per-request)
- TTL para auto-expiración
- GSI para queries por fecha
- Alarmas de throttling

### Variables

```hcl
variable "table_name" {
  description = "Nombre de la tabla"
  type        = string
}

variable "hash_key" {
  description = "Atributo para partition key"
  type = object({
    name = string
    type = string  # S, N, B
  })
}

variable "range_key" {
  description = "Atributo para sort key (opcional)"
  type = object({
    name = string
    type = string
  })
  default = null
}

variable "billing_mode" {
  description = "Modo de billing: PAY_PER_REQUEST o PROVISIONED"
  type        = string
  default     = "PAY_PER_REQUEST"
}

variable "ttl_attribute" {
  description = "Nombre del atributo para TTL"
  type        = string
  default     = ""
}

variable "global_secondary_indexes" {
  description = "Lista de GSIs"
  type = list(object({
    name               = string
    hash_key           = string
    hash_key_type      = string
    range_key          = optional(string)
    range_key_type     = optional(string)
    projection_type    = string  # ALL, KEYS_ONLY, INCLUDE
    non_key_attributes = optional(list(string))
  }))
  default = []
}

variable "enable_point_in_time_recovery" {
  description = "Habilitar PITR para backups"
  type        = bool
  default     = false
}

variable "enable_deletion_protection" {
  description = "Protección contra eliminación accidental"
  type        = bool
  default     = false
}

variable "alarm_sns_topic_arn" {
  type    = string
  default = ""
}

variable "tags" {
  type    = map(string)
  default = {}
}
```

### Recursos

```hcl
# main.tf

# Construir lista de atributos únicos
locals {
  # Atributos de keys principales
  primary_attributes = concat(
    [var.hash_key],
    var.range_key != null ? [var.range_key] : []
  )

  # Atributos de GSIs
  gsi_attributes = flatten([
    for gsi in var.global_secondary_indexes : concat(
      [{ name = gsi.hash_key, type = gsi.hash_key_type }],
      gsi.range_key != null ? [{ name = gsi.range_key, type = gsi.range_key_type }] : []
    )
  ])

  # Todos los atributos únicos
  all_attributes = distinct(concat(local.primary_attributes, local.gsi_attributes))
}

resource "aws_dynamodb_table" "this" {
  name         = var.table_name
  billing_mode = var.billing_mode
  hash_key     = var.hash_key.name
  range_key    = var.range_key != null ? var.range_key.name : null

  # Definición de atributos
  dynamic "attribute" {
    for_each = local.all_attributes
    content {
      name = attribute.value.name
      type = attribute.value.type
    }
  }

  # TTL
  dynamic "ttl" {
    for_each = var.ttl_attribute != "" ? [1] : []
    content {
      attribute_name = var.ttl_attribute
      enabled        = true
    }
  }

  # GSIs
  dynamic "global_secondary_index" {
    for_each = var.global_secondary_indexes
    content {
      name            = global_secondary_index.value.name
      hash_key        = global_secondary_index.value.hash_key
      range_key       = global_secondary_index.value.range_key
      projection_type = global_secondary_index.value.projection_type

      dynamic "non_key_attributes" {
        for_each = global_secondary_index.value.projection_type == "INCLUDE" ? [1] : []
        content {
          non_key_attributes = global_secondary_index.value.non_key_attributes
        }
      }
    }
  }

  # Point-in-time recovery
  point_in_time_recovery {
    enabled = var.enable_point_in_time_recovery
  }

  # Deletion protection
  deletion_protection_enabled = var.enable_deletion_protection

  # Server-side encryption (default: AWS owned key)
  server_side_encryption {
    enabled = true
  }

  tags = var.tags
}

# Alarma de throttling
resource "aws_cloudwatch_metric_alarm" "throttled_requests" {
  count = var.alarm_sns_topic_arn != "" ? 1 : 0

  alarm_name          = "${var.table_name}-throttled-requests"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ThrottledRequests"
  namespace           = "AWS/DynamoDB"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "DynamoDB table ${var.table_name} has throttled requests"
  alarm_actions       = [var.alarm_sns_topic_arn]

  dimensions = {
    TableName = aws_dynamodb_table.this.name
  }

  tags = var.tags
}
```

### Outputs

```hcl
output "table_name" {
  value = aws_dynamodb_table.this.name
}

output "table_arn" {
  value = aws_dynamodb_table.this.arn
}

output "table_id" {
  value = aws_dynamodb_table.this.id
}

output "gsi_arns" {
  value = {
    for idx, gsi in var.global_secondary_indexes : gsi.name =>
      "${aws_dynamodb_table.this.arn}/index/${gsi.name}"
  }
}
```

### Uso

```hcl
module "dynamodb" {
  source = "../../modules/dynamodb"

  table_name = "sql-to-nosql-schemas"

  hash_key = {
    name = "conversionId"
    type = "S"
  }

  ttl_attribute = "expiresAt"

  global_secondary_indexes = [
    {
      name            = "byDate"
      hash_key        = "conversionDate"
      hash_key_type   = "S"
      projection_type = "ALL"
    }
  ]

  enable_point_in_time_recovery = true
  alarm_sns_topic_arn          = module.sns.topic_arn

  tags = local.common_tags
}
```

---

## SQS

**Path:** `modules/sqs/`

### Descripción

Colas SQS con Dead Letter Queues para manejo de errores.

### Variables

```hcl
variable "queue_name" {
  description = "Nombre base de la cola"
  type        = string
}

variable "visibility_timeout_seconds" {
  description = "Tiempo que el mensaje es invisible después de recibirse"
  type        = number
  default     = 30
}

variable "message_retention_seconds" {
  description = "Tiempo de retención de mensajes"
  type        = number
  default     = 86400  # 1 día
}

variable "receive_wait_time_seconds" {
  description = "Long polling wait time"
  type        = number
  default     = 20
}

variable "max_receive_count" {
  description = "Reintentos antes de enviar a DLQ"
  type        = number
  default     = 3
}

variable "dlq_message_retention_seconds" {
  description = "Retención en DLQ"
  type        = number
  default     = 1209600  # 14 días
}

variable "enable_encryption" {
  description = "Habilitar SSE-SQS"
  type        = bool
  default     = true
}

variable "alarm_sns_topic_arn" {
  type    = string
  default = ""
}

variable "message_age_threshold_seconds" {
  description = "Umbral para alarma de edad de mensajes"
  type        = number
  default     = 3600  # 1 hora
}

variable "tags" {
  type    = map(string)
  default = {}
}
```

### Recursos

```hcl
# main.tf

# Dead Letter Queue
resource "aws_sqs_queue" "dlq" {
  name                      = "${var.queue_name}-dlq"
  message_retention_seconds = var.dlq_message_retention_seconds
  sqs_managed_sse_enabled   = var.enable_encryption

  tags = merge(var.tags, {
    Type = "DLQ"
  })
}

# Cola principal
resource "aws_sqs_queue" "main" {
  name                       = var.queue_name
  visibility_timeout_seconds = var.visibility_timeout_seconds
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds
  sqs_managed_sse_enabled    = var.enable_encryption

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = var.tags
}

# Política para permitir redrive
resource "aws_sqs_queue_redrive_allow_policy" "dlq" {
  queue_url = aws_sqs_queue.dlq.id

  redrive_allow_policy = jsonencode({
    redrivePermission = "byQueue"
    sourceQueueArns   = [aws_sqs_queue.main.arn]
  })
}
```

### Alarmas

```hcl
# alarms.tf

# Alarma: mensajes en DLQ
resource "aws_cloudwatch_metric_alarm" "dlq_messages" {
  count = var.alarm_sns_topic_arn != "" ? 1 : 0

  alarm_name          = "${var.queue_name}-dlq-has-messages"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "DLQ ${var.queue_name}-dlq has messages"
  alarm_actions       = [var.alarm_sns_topic_arn]

  dimensions = {
    QueueName = aws_sqs_queue.dlq.name
  }

  tags = var.tags
}

# Alarma: mensajes envejeciendo
resource "aws_cloudwatch_metric_alarm" "message_age" {
  count = var.alarm_sns_topic_arn != "" ? 1 : 0

  alarm_name          = "${var.queue_name}-message-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Maximum"
  threshold           = var.message_age_threshold_seconds
  alarm_description   = "Queue ${var.queue_name} has old messages"
  alarm_actions       = [var.alarm_sns_topic_arn]

  dimensions = {
    QueueName = aws_sqs_queue.main.name
  }

  tags = var.tags
}
```

### Outputs

```hcl
output "queue_url" {
  value = aws_sqs_queue.main.url
}

output "queue_arn" {
  value = aws_sqs_queue.main.arn
}

output "queue_name" {
  value = aws_sqs_queue.main.name
}

output "dlq_url" {
  value = aws_sqs_queue.dlq.url
}

output "dlq_arn" {
  value = aws_sqs_queue.dlq.arn
}

output "dlq_name" {
  value = aws_sqs_queue.dlq.name
}
```

---

## CloudFront

**Path:** `modules/cloudfront/`

### Descripción

CloudFront distribution con:
- Múltiples orígenes (S3 + API Gateway)
- OAC para S3
- CloudFront Function para seguridad
- Certificado ACM

### Variables Principales

```hcl
variable "distribution_name" {
  type = string
}

variable "custom_domain" {
  type = string
}

variable "acm_certificate_arn" {
  type = string
}

variable "s3_origin" {
  type = object({
    domain_name = string
    origin_id   = string
  })
}

variable "api_origin" {
  type = object({
    domain_name   = string
    origin_id     = string
    origin_path   = string
    custom_header = object({
      name  = string
      value = string
    })
  })
}

variable "default_root_object" {
  type    = string
  default = "index.html"
}

variable "price_class" {
  type    = string
  default = "PriceClass_100"  # US, Canada, Europe
}
```

### Recursos Clave

```hcl
# Origin Access Control
resource "aws_cloudfront_origin_access_control" "s3" {
  name                              = "${var.distribution_name}-s3-oac"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# CloudFront Function
resource "aws_cloudfront_function" "security" {
  name    = "${var.distribution_name}-security"
  runtime = "cloudfront-js-2.0"
  code    = file("${path.module}/functions/validate-origin.js")
}

# Distribution
resource "aws_cloudfront_distribution" "this" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = var.default_root_object
  price_class         = var.price_class
  aliases             = [var.custom_domain]

  # Origin S3
  origin {
    domain_name              = var.s3_origin.domain_name
    origin_id                = var.s3_origin.origin_id
    origin_access_control_id = aws_cloudfront_origin_access_control.s3.id
  }

  # Origin API Gateway
  origin {
    domain_name = var.api_origin.domain_name
    origin_id   = var.api_origin.origin_id
    origin_path = var.api_origin.origin_path

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }

    custom_header {
      name  = var.api_origin.custom_header.name
      value = var.api_origin.custom_header.value
    }
  }

  # Default: S3
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = var.s3_origin.origin_id
    viewer_protocol_policy = "redirect-to-https"
    compress               = true

    forwarded_values {
      query_string = false
      cookies { forward = "none" }
    }

    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.security.arn
    }
  }

  # API: /prod/*
  ordered_cache_behavior {
    path_pattern           = "/prod/*"
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = var.api_origin.origin_id
    viewer_protocol_policy = "https-only"

    # Sin cache para API
    min_ttl     = 0
    default_ttl = 0
    max_ttl     = 0

    forwarded_values {
      query_string = true
      headers      = ["Authorization", "Origin"]
      cookies { forward = "none" }
    }
  }

  # SPA fallback
  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  viewer_certificate {
    acm_certificate_arn      = var.acm_certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }

  restrictions {
    geo_restriction { restriction_type = "none" }
  }
}
```

---

## Bedrock

**Path:** `modules/bedrock/`

### Descripción

Políticas IAM para acceso a modelos Bedrock (Claude).

### Variables

```hcl
variable "project_name" {
  type = string
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "models" {
  description = "Lista de modelos Bedrock a usar"
  type = list(object({
    model_id = string
    alias    = string
  }))
  default = [
    {
      model_id = "anthropic.claude-3-5-sonnet-20241022-v2:0"
      alias    = "claude-sonnet"
    },
    {
      model_id = "anthropic.claude-3-5-haiku-20241022-v1:0"
      alias    = "claude-haiku"
    }
  ]
}

variable "lambda_role_arns" {
  description = "ARNs de roles Lambda que necesitan acceso"
  type        = list(string)
}
```

### Recursos

```hcl
data "aws_caller_identity" "current" {}

# Política de acceso a Bedrock
resource "aws_iam_policy" "bedrock_invoke" {
  name        = "${var.project_name}-bedrock-invoke"
  description = "Allow invoking Bedrock models"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "InvokeModels"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Resource = concat(
          # Foundation models
          [for model in var.models :
            "arn:aws:bedrock:${var.aws_region}::foundation-model/${model.model_id}"
          ],
          # Inference profiles (cross-region)
          ["arn:aws:bedrock:us.*:${data.aws_caller_identity.current.account_id}:inference-profile/us.anthropic.*"]
        )
      }
    ]
  })
}

# Attach a cada rol Lambda
resource "aws_iam_role_policy_attachment" "bedrock" {
  count = length(var.lambda_role_arns)

  role       = element(split("/", var.lambda_role_arns[count.index]), 1)
  policy_arn = aws_iam_policy.bedrock_invoke.arn
}
```

---

## Uso de Módulos en Producción

### Ejemplo Completo

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

locals {
  project_name = "sql-to-nosql"
  environment  = "prod"

  common_tags = {
    Project     = local.project_name
    Environment = local.environment
    ManagedBy   = "terraform"
  }
}

# DynamoDB
module "dynamodb" {
  source = "../../modules/dynamodb"

  table_name = "${local.project_name}-schemas"
  hash_key   = { name = "conversionId", type = "S" }

  ttl_attribute = "expiresAt"

  global_secondary_indexes = [{
    name            = "byDate"
    hash_key        = "conversionDate"
    hash_key_type   = "S"
    projection_type = "ALL"
  }]

  alarm_sns_topic_arn = module.sns.topic_arn
  tags               = local.common_tags
}

# SQS - Conversion Queue
module "sqs_conversion" {
  source = "../../modules/sqs"

  queue_name                 = "${local.project_name}-conversion"
  visibility_timeout_seconds = 180
  max_receive_count          = 3

  alarm_sns_topic_arn = module.sns.topic_arn
  tags               = local.common_tags
}

# Lambda - Process Handler
module "process_handler" {
  source = "../../modules/lambda"

  function_name = "${local.project_name}-process-handler"
  source_path   = "${path.root}/../../../dist/diagrams.zip"
  role_arn      = module.iam_process.role_arn
  timeout       = 30

  environment_variables = {
    DYNAMODB_TABLE_NAME = module.dynamodb.table_name
    SQS_QUEUE_URL       = module.sqs_conversion.queue_url
  }

  alarm_sns_topic_arn = module.sns.topic_arn
  tags               = local.common_tags
}

# API Gateway
module "api_gateway" {
  source = "../../modules/gateway/http-v2"

  api_name   = "${local.project_name}-api"
  stage_name = "prod"

  routes = {
    "post-schemas" = {
      method     = "POST"
      path       = "/api/v1/schemas"
      lambda_arn = module.process_handler.function_arn
      invoke_arn = module.process_handler.invoke_arn
    }
    "get-schemas" = {
      method     = "GET"
      path       = "/api/v1/schemas"
      lambda_arn = module.query_handler.function_arn
      invoke_arn = module.query_handler.invoke_arn
    }
  }

  tags = local.common_tags
}
```

---

## Mejores Prácticas

### 1. Versionado de Módulos

```hcl
module "lambda" {
  source  = "../../modules/lambda"
  # Para módulos externos:
  # source  = "git::https://github.com/org/modules.git//lambda?ref=v1.2.0"
}
```

### 2. Validaciones

```hcl
variable "environment" {
  type = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}
```

### 3. Outputs Consistentes

```hcl
# Siempre exportar: name, arn, id
output "table_name" { value = aws_dynamodb_table.this.name }
output "table_arn" { value = aws_dynamodb_table.this.arn }
output "table_id" { value = aws_dynamodb_table.this.id }
```

### 4. Tags Propagados

```hcl
locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# Usar en todos los módulos
tags = local.common_tags
```
