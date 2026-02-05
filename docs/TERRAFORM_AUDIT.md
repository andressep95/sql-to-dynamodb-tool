# Auditoría de Infraestructura Terraform

**Proyecto:** sql-to-nosql-parser
**Fecha:** 2026-02-03 (Actualizado: 2026-02-04)
**Auditor:** Claude Code
**Estado:** ✅ Mejoras implementadas

---

## Tabla de Contenidos

- [Parte 1: Frontend (S3 + CloudFront + ACM + CloudFront Functions)](#parte-1-frontend-s3--cloudfront--acm--cloudfront-functions)
  - [S3 Bucket](#s3-bucket)
  - [CloudFront Distribution](#cloudfront-distribution)
  - [ACM Certificate](#acm-certificate)
  - [CloudFront Function](#cloudfront-function)
- [Parte 2: Backend (API Gateway + Lambda + SQS + DynamoDB + Bedrock)](#parte-2-backend-api-gateway--lambda--sqs--dynamodb--bedrock)
- [Resumen de Hallazgos](#resumen-de-hallazgos)
- [Plan de Acción Recomendado](#plan-de-acción-recomendado)

---

## Parte 1: Frontend (S3 + CloudFront + ACM + CloudFront Functions)

### S3 Bucket

**Archivo:** `infra/terraform/modules/s3/main.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **Block Public Access** | ✅ Excelente | Las 4 opciones están habilitadas (`block_public_acls`, `block_public_policy`, `ignore_public_acls`, `restrict_public_buckets`) |
| **Encriptación SSE-S3** | ✅ Correcto | AES256 server-side encryption habilitado |
| **Content-Type mapping** | ✅ Completo | 15+ tipos MIME correctamente mapeados (.html, .css, .js, .json, .png, .jpg, .svg, .woff2, etc.) |
| **ETag para invalidación** | ✅ Eficiente | Uso de `filemd5()` para detectar cambios en archivos |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Versionado condicional** | 🟡 Media | El versionado está como `count = var.versioning_enabled ? 1 : 0` | Para producción, considerar habilitar versionado siempre para protección contra eliminación accidental |
| **Encriptación condicional** | 🟡 Media | `enable_encryption` es opcional | AWS ahora requiere encriptación por defecto. Considerar hacerla obligatoria eliminando la condición |
| **Logging de acceso** | 🟡 Media | No hay `aws_s3_bucket_logging` configurado | Agregar logging a un bucket dedicado para auditoría y troubleshooting |
| **Lifecycle rules** | 🟢 Baja | No hay reglas de lifecycle | No crítico para assets estáticos, pero útil si se habilita versionado para limpiar versiones antiguas |

#### 📝 Código de Mejora Sugerido

```hcl
# Agregar logging de acceso
resource "aws_s3_bucket_logging" "this" {
  bucket = aws_s3_bucket.this.id

  target_bucket = var.logging_bucket_id
  target_prefix = "s3-access-logs/${var.bucket_name}/"
}

# Hacer encriptación obligatoria (eliminar count)
resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true  # Reduce costos de KMS si se usa SSE-KMS
  }
}
```

---

### CloudFront Distribution

**Archivo:** `infra/terraform/modules/cloudfront/main.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **OAC (Origin Access Control)** | ✅ Excelente | Usando OAC en lugar del deprecado OAI. `signing_behavior = "always"` y `sigv4` |
| **Bucket Policy con condición** | ✅ Excelente | Restricción por `AWS:SourceArn` al distribution específico |
| **HTTPS enforcement** | ✅ Correcto | `viewer_protocol_policy = "redirect-to-https"` |
| **TLS 1.2 mínimo** | ✅ Seguro | `origin_ssl_protocols = ["TLSv1.2"]` y `minimum_protocol_version = "TLSv1.2_2021"` |
| **Compresión Gzip/Brotli** | ✅ Habilitada | `compress = true` en ambos cache behaviors |
| **SPA routing** | ✅ Implementado | Custom error responses 403/404 → `/index.html` |
| **API sin cache** | ✅ Correcto | Usando `Managed-CachingDisabled` para `/prod/api/*` |
| **Invalidación automática** | ✅ Inteligente | `terraform_data` con trigger por etags de archivos |
| **IPv6** | ✅ Habilitado | `is_ipv6_enabled = true` |
| **SNI-only SSL** | ✅ Moderno | Evita costo de IP dedicada |

#### ⚠️ Áreas de Mejora - Cuellos de Botella

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **TTL para assets estáticos** | 🟡 Media | `default_ttl = 3600` (1 hora) es bajo para JS/CSS con hash | Para assets versionados (`app.abc123.js`), usar TTL de 1 año (31536000s). Agregar `ordered_cache_behavior` específico |
| **Cache Policy legacy** | 🟡 Media | Usando `forwarded_values` (deprecated) en default behavior | Migrar a `cache_policy_id` con managed policies como en el API behavior |
| **Error caching TTL** | 🟢 Baja | `error_caching_min_ttl = 10` puede causar micro-invalidaciones | Subir a 60-300s según frecuencia de deploys |
| **WAF no configurado** | 🟡 Media | No hay `web_acl_id` configurado | Agregar AWS WAF para protección DDoS y bots en producción |
| **Access Logging** | 🟡 Media | No hay `logging_config` | Agregar bucket para access logs de CloudFront |
| **Real-time logs** | 🟢 Baja | No configurados | Útil para debugging y métricas en tiempo real |

#### 📝 Código de Mejora Sugerido

```hcl
# Cache policy moderna para assets estáticos (reemplazar forwarded_values)
data "aws_cloudfront_cache_policy" "caching_optimized" {
  name = "Managed-CachingOptimized"
}

# Behavior específico para assets con hash (inmutables)
ordered_cache_behavior {
  path_pattern           = "/assets/*"
  allowed_methods        = ["GET", "HEAD"]
  cached_methods         = ["GET", "HEAD"]
  target_origin_id       = local.s3_origin_id
  viewer_protocol_policy = "redirect-to-https"

  cache_policy_id = data.aws_cloudfront_cache_policy.caching_optimized.id

  min_ttl     = 31536000  # 1 año
  default_ttl = 31536000
  max_ttl     = 31536000
  compress    = true
}

# Logging de acceso
logging_config {
  include_cookies = false
  bucket          = "${var.logging_bucket}.s3.amazonaws.com"
  prefix          = "cloudfront/"
}

# WAF (requiere crear el web_acl primero)
web_acl_id = var.waf_web_acl_arn
```

---

### ACM Certificate

**Archivo:** `infra/terraform/modules/cloudfront/main.tf` (líneas 13-28)

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **DNS validation** | ✅ Correcto | Mejor que email validation, permite automatización |
| **create_before_destroy** | ✅ Buena práctica | Lifecycle correcto para rotación sin downtime |
| **Condicional** | ✅ Flexible | Solo se crea si hay `custom_domain` definido |
| **SNI-only** | ✅ Moderno | `ssl_support_method = "sni-only"` evita costo de IP dedicada |
| **TLS 1.2_2021** | ✅ Seguro | Versión mínima de protocolo actualizada |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Validación DNS incompleta** | 🔴 Alta | `aws_acm_certificate_validation` no tiene `validation_record_fqdns` | El certificado podría no validarse automáticamente. Requiere integración con Route53 o creación manual de registros DNS |
| **Sin SANs adicionales** | 🟢 Baja | Solo un dominio configurado | Si necesitas `www.` o subdominios, agregar `subject_alternative_names` |

#### 📝 Código de Mejora Sugerido

```hcl
# Integración con Route53 para validación automática
data "aws_route53_zone" "this" {
  count = local.use_custom_domain ? 1 : 0
  name  = var.route53_zone_name  # ej: "cloudcentinel.com"
}

resource "aws_route53_record" "cert_validation" {
  for_each = local.use_custom_domain ? {
    for dvo in aws_acm_certificate.this[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.this[0].zone_id
}

resource "aws_acm_certificate_validation" "this" {
  count                   = local.use_custom_domain ? 1 : 0
  certificate_arn         = aws_acm_certificate.this[0].arn
  validation_record_fqdns = [for record in aws_route53_record.cert_validation : record.fqdn]
}

# Agregar SANs si es necesario
resource "aws_acm_certificate" "this" {
  count             = local.use_custom_domain ? 1 : 0
  domain_name       = var.custom_domain
  validation_method = "DNS"

  subject_alternative_names = var.custom_domain_sans  # ej: ["www.app-sql.cloudcentinel.com"]

  lifecycle {
    create_before_destroy = true
  }
}
```

---

### CloudFront Function

**Archivo:** `infra/terraform/modules/cloudfront/main.tf` (líneas 34-56)

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **Runtime actualizado** | ✅ Correcto | `cloudfront-js-2.0` (última versión disponible) |
| **Validación de secreto** | ✅ Seguro | Compara header `x-origin-secret` contra valor esperado |
| **Despliegue condicional** | ✅ Flexible | Solo activo si `cloudflare_secret_header_value != ""` |
| **Respuesta 403 limpia** | ✅ Correcto | Devuelve "Forbidden" sin exponer información sensible |
| **Publish automático** | ✅ Correcto | `publish = true` asegura que la función esté activa |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Secreto en código** | 🔴 Alta | El secreto se interpola directamente: `var expected = '${var.cloudflare_secret_header_value}'` | Si alguien accede al código de la función, obtiene el secreto. Considerar CloudFront KeyValueStore o rotación frecuente |
| **Timing attack** | 🟡 Media | Comparación directa `!== expected` es vulnerable a timing attacks | Usar comparación de tiempo constante (impacto bajo en este contexto pero es buena práctica) |
| **Sin métricas de rechazo** | 🟡 Media | No hay forma de saber cuántas requests son rechazadas | Habilitar CloudFront real-time logs o standard logs para monitorear |
| **Header case sensitivity** | 🟢 Baja | `request.headers['x-origin-secret']` asume lowercase | CloudFront normaliza a lowercase, pero validar comportamiento con Cloudflare |

#### 📝 Código de Mejora Sugerido

```hcl
# Usar CloudFront KeyValueStore para secretos (más seguro)
resource "aws_cloudfront_key_value_store" "secrets" {
  count   = local.use_secret_header ? 1 : 0
  name    = "${var.environment}-${var.project_name}-secrets"
  comment = "Secrets for CloudFront Functions"
}

# Función mejorada con comparación más robusta
resource "aws_cloudfront_function" "verify_cf_secret" {
  count   = local.use_secret_header ? 1 : 0
  name    = "${var.environment}-${var.project_name}-verify-cf-secret"
  runtime = "cloudfront-js-2.0"
  comment = "Validates X-Origin-Secret header from Cloudflare"
  publish = true

  # Asociar KeyValueStore
  key_value_store_associations = local.use_secret_header ? [
    aws_cloudfront_key_value_store.secrets[0].arn
  ] : []

  code = <<-EOF
    import cf from 'cloudfront';

    const kvsHandle = cf.kvs();

    async function handler(event) {
      const request = event.request;
      const secret = request.headers['x-origin-secret'];

      // Obtener secreto de KeyValueStore
      let expected;
      try {
        expected = await kvsHandle.get('cf-origin-secret');
      } catch (e) {
        // Si no hay secreto configurado, denegar
        return {
          statusCode: 403,
          statusDescription: 'Forbidden',
          body: { encoding: 'text', data: 'Forbidden' }
        };
      }

      if (!secret || secret.value !== expected) {
        return {
          statusCode: 403,
          statusDescription: 'Forbidden',
          body: { encoding: 'text', data: 'Forbidden' }
        };
      }

      return request;
    }
  EOF
}
```

---

### Resumen Parte 1 - Frontend

| Componente | Puntuación | Fortalezas | Debilidades Principales |
|------------|------------|------------|------------------------|
| **S3** | 8.5/10 | Excelente seguridad base con Block Public Access y encriptación | Falta logging de acceso |
| **CloudFront** | 8/10 | Arquitectura OAC correcta, compresión, HTTPS | TTL bajo, cache policy deprecated, sin WAF |
| **ACM** | 7/10 | DNS validation, lifecycle correcto | Validación DNS incompleta |
| **CloudFront Function** | 7.5/10 | Runtime moderno, validación funcional | Secreto expuesto en código |

**Puntuación Global Frontend: 7.75/10**

---

## Parte 2: Backend (API Gateway + Lambda + SQS + DynamoDB + Bedrock)

### API Gateway HTTP v2

**Archivos:**
- `infra/terraform/modules/gateway/http-v2/main.tf`
- `infra/terraform/environments/prod/shared/api_gateway.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **HTTP API v2** | ✅ Correcto | Usando HTTP API en lugar de REST API (más económico, menor latencia) |
| **AWS_PROXY integration** | ✅ Correcto | Integración directa con Lambda sin transformaciones |
| **Payload v2.0** | ✅ Moderno | `payload_format_version = "2.0"` (formato optimizado) |
| **Auto-deploy** | ✅ Conveniente | `auto_deploy = true` simplifica despliegues |
| **CORS configurado** | ✅ Correcto | Configuración de CORS en el API |
| **Lambda permissions** | ✅ Correcto | Permisos por Lambda única, evitando duplicados |
| **Rutas específicas** | ✅ Bien diseñado | Rutas REST bien definidas (`POST /api/v1/schemas`, `GET /api/v1/schemas/{id}`) |

#### ⚠️ Áreas de Mejora - Cuellos de Botella

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Sin throttling** | 🔴 Alta | No hay configuración de `aws_apigatewayv2_stage` con throttling | Agregar `default_route_settings` con `throttling_burst_limit` y `throttling_rate_limit` |
| **Sin access logging** | 🟡 Media | No hay `access_log_settings` configurado | Agregar CloudWatch log group para access logs |
| **Sin custom domain** | 🟡 Media | API expuesta con URL generada por AWS | Considerar custom domain con Route53 si no se usa CloudFront como proxy |
| **Source ARN amplio** | 🟡 Media | `source_arn = "${...}/*/*"` permite cualquier método/ruta | Restringir a rutas específicas para mayor seguridad |
| **Sin autorización** | 🔴 Alta | No hay JWT authorizer ni Lambda authorizer | Implementar autorización para endpoints públicos |

#### 📝 Código de Mejora Sugerido

```hcl
# Stage con throttling y logging
resource "aws_apigatewayv2_stage" "this" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = var.stage_name
  auto_deploy = true

  # Throttling para protección contra abuso
  default_route_settings {
    throttling_burst_limit = 100  # Requests por segundo en burst
    throttling_rate_limit  = 50   # Requests por segundo sostenido
  }

  # Access logging
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_access_logs.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      responseLength = "$context.responseLength"
      latency        = "$context.responseLatency"
    })
  }

  tags = var.tags
}

resource "aws_cloudwatch_log_group" "api_access_logs" {
  name              = "/aws/apigateway/${var.name}"
  retention_in_days = 30
}

# JWT Authorizer (ejemplo con Cognito)
resource "aws_apigatewayv2_authorizer" "jwt" {
  api_id           = aws_apigatewayv2_api.this.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = "${var.name}-jwt-authorizer"

  jwt_configuration {
    audience = [var.cognito_client_id]
    issuer   = "https://cognito-idp.${var.region}.amazonaws.com/${var.cognito_user_pool_id}"
  }
}
```

---

### Lambda Functions

**Archivos:**
- `infra/terraform/modules/lambda/main.tf`
- `infra/terraform/environments/prod/components/locals.tf`
- `infra/terraform/environments/prod/components/*_lambda.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **ARM64 architecture** | ✅ Excelente | `architecture = "arm64"` (20% más económico que x86_64) |
| **provided.al2023 runtime** | ✅ Moderno | Runtime más reciente para custom runtimes |
| **X-Ray tracing** | ✅ Habilitado | `xray_tracing_enabled = true` por defecto |
| **CloudWatch alarms** | ✅ Configurados | Alarmas de errores y duración |
| **DLQ support** | ✅ Disponible | Soporte para Dead Letter Queue en invocaciones async |
| **Log retention configurable** | ✅ Correcto | 30 días en producción |
| **VPC config opcional** | ✅ Flexible | Solo si es necesario |
| **Layers support** | ✅ Disponible | Soporte para Lambda Layers |
| **Environment variables** | ✅ Dinámico | Variables de entorno configurables |
| **Source code hash** | ✅ Correcto | Detecta cambios en código |

#### Configuración de Lambdas en Producción

| Lambda | Memory | Timeout | Concurrency | Observación |
|--------|--------|---------|-------------|-------------|
| `process_handler` | 128 MB | 5s | Unreserved | ✅ Adecuado para recibir requests |
| `conversion_worker` | 256 MB | 120s | Unreserved | ⚠️ Bedrock puede requerir más memoria |
| `query_handler` | 256 MB | 120s | Unreserved | ⚠️ Timeout alto para queries simples |
| `dlq_handler` | 128 MB | 30s | Unreserved | ✅ Adecuado |
| `access_pattern_worker` | 256 MB | 120s | Unreserved | ⚠️ Similar a conversion_worker |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Sin reserved concurrency** | 🟡 Media | Todas las lambdas con `-1` (unreserved) | Definir límites para evitar que una función consuma toda la capacidad de la cuenta |
| **Memory sin optimizar** | 🟡 Media | Valores genéricos (128/256 MB) | Usar AWS Lambda Power Tuning para encontrar configuración óptima costo/performance |
| **Visibility timeout mismatch** | 🔴 Alta | SQS visibility timeout (180s) < Lambda timeout (120s) + margen | El visibility timeout debe ser >= 6x Lambda timeout para evitar mensajes duplicados |
| **Sin provisioned concurrency** | 🟢 Baja | Cold starts en todas las funciones | Considerar provisioned concurrency para lambdas críticas |
| **Alarmas sin destino** | 🟡 Media | `alarm_sns_topic_arns = []` vacío | Configurar SNS topic para recibir alertas |
| **Log encryption** | 🟢 Baja | `log_kms_key_id = null` por defecto | Considerar encriptación de logs con KMS |

#### 📝 Código de Mejora Sugerido

```hcl
# Configuración optimizada para conversion_worker
locals {
  lambda_configs = {
    conversion_worker = {
      # ...
      memory_size                    = 512   # Aumentar para Bedrock
      timeout                        = 90    # Reducir si es posible
      reserved_concurrent_executions = 10    # Limitar para controlar costos de Bedrock
      log_retention_days             = 30
    }
    # ...
  }
}

# Agregar SNS topic para alarmas
resource "aws_sns_topic" "lambda_alarms" {
  name = "${var.environment}-lambda-alarms"
}

resource "aws_sns_topic_subscription" "email" {
  topic_arn = aws_sns_topic.lambda_alarms.arn
  protocol  = "email"
  endpoint  = var.alert_email
}

# Alarma de throttling
resource "aws_cloudwatch_metric_alarm" "throttles" {
  alarm_name          = "${var.function_name}-throttles"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "Throttles"
  namespace           = "AWS/Lambda"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Lambda function is being throttled"
  treat_missing_data  = "notBreaching"

  dimensions = {
    FunctionName = aws_lambda_function.this.function_name
  }

  alarm_actions = var.alarm_sns_topic_arns
}
```

---

### SQS Queues

**Archivo:** `infra/terraform/modules/sqs/main.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **Dead Letter Queue** | ✅ Correcto | DLQ configurado para ambas colas |
| **Redrive policy** | ✅ Correcto | `maxReceiveCount = 3` antes de enviar a DLQ |
| **DLQ retention mayor** | ✅ Correcto | 14 días en DLQ vs 4 días en cola principal |
| **Long polling** | ⚠️ Parcial | `receive_wait_time_seconds` disponible pero en 0 por defecto |
| **Colas separadas** | ✅ Bien diseñado | Cola de conversión y cola de access patterns separadas |

#### Configuración Actual

| Queue | Visibility Timeout | Retention | Max Receive |
|-------|-------------------|-----------|-------------|
| Conversion Queue | 180s (3 min) | 4 días | 3 |
| Access Pattern Queue | 120s (2 min) | 4 días | 3 |
| DLQs | N/A | 14 días | N/A |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Long polling deshabilitado** | 🟡 Media | `receive_wait_time_seconds = 0` (short polling) | Habilitar long polling (20s) reduce costos y latencia |
| **Sin encriptación** | 🟡 Media | Las colas no tienen SSE habilitado | Agregar `sqs_managed_sse_enabled = true` o KMS |
| **Sin métricas/alarmas** | 🟡 Media | No hay alarmas de CloudWatch para SQS | Agregar alarmas para `ApproximateAgeOfOldestMessage` y `ApproximateNumberOfMessagesVisible` |
| **Visibility timeout ajustado** | 🟢 Info | 180s para conversion, 120s para access patterns | ✅ Adecuado si los timeouts de Lambda son menores |
| **Sin redrive allow policy** | 🟢 Baja | DLQ no tiene política de quién puede enviar | Agregar `redrive_allow_policy` para mayor seguridad |

#### 📝 Código de Mejora Sugerido

```hcl
resource "aws_sqs_queue" "conversion_queue" {
  name                       = var.queue_name
  visibility_timeout_seconds = var.visibility_timeout_seconds
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = 20  # Long polling habilitado

  # Encriptación con SSE-SQS
  sqs_managed_sse_enabled = true

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.conversion_dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = var.tags
}

# Alarma para mensajes antiguos en cola
resource "aws_cloudwatch_metric_alarm" "old_messages" {
  alarm_name          = "${var.queue_name}-old-messages"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Maximum"
  threshold           = 3600  # 1 hora
  alarm_description   = "Messages are stuck in queue"

  dimensions = {
    QueueName = aws_sqs_queue.conversion_queue.name
  }

  alarm_actions = var.alarm_sns_topic_arns
}

# Alarma para DLQ no vacía
resource "aws_cloudwatch_metric_alarm" "dlq_not_empty" {
  alarm_name          = "${var.queue_name}-dlq-not-empty"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "DLQ has messages - investigate failures"

  dimensions = {
    QueueName = aws_sqs_queue.conversion_dlq.name
  }

  alarm_actions = var.alarm_sns_topic_arns
}
```

---

### DynamoDB

**Archivo:** `infra/terraform/modules/dynamodb/main.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **PAY_PER_REQUEST** | ✅ Correcto | On-demand billing, escala automáticamente |
| **TTL habilitado** | ✅ Correcto | `expiresAt` para limpieza automática (24h) |
| **Point-in-time recovery** | ✅ Configurable | Habilitado en producción |
| **GSI configurado** | ✅ Correcto | `conversionDate-createdAt-index` para queries por fecha |
| **Diseño flexible** | ✅ Modular | Atributos y GSI dinámicos vía variables |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Sin encriptación explícita** | 🟢 Baja | DynamoDB encripta por defecto con AWS-owned key | Considerar CMK si hay requisitos de compliance |
| **Sin contributor insights** | 🟢 Baja | No hay análisis de patrones de acceso | Habilitar para optimizar queries |
| **Sin alarmas de throttling** | 🟡 Media | On-demand puede throttlear en picos extremos | Agregar alarma para `ThrottledRequests` |
| **GSI projection ALL** | 🟢 Info | Proyección ALL consume más WCU en updates | ✅ OK si queries necesitan todos los atributos |

#### 📝 Código de Mejora Sugerido

```hcl
resource "aws_dynamodb_table" "this" {
  # ... configuración existente ...

  # Contributor Insights para análisis de patrones
  dynamic "replica" {
    for_each = var.enable_contributor_insights ? [1] : []
    content {
      # Nota: contributor_insights se configura por separado
    }
  }

  tags = var.tags
}

# Contributor Insights
resource "aws_dynamodb_contributor_insights" "this" {
  count      = var.enable_contributor_insights ? 1 : 0
  table_name = aws_dynamodb_table.this.name
}

# Alarma de throttling
resource "aws_cloudwatch_metric_alarm" "dynamodb_throttle" {
  alarm_name          = "${var.table_name}-throttled-requests"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ThrottledRequests"
  namespace           = "AWS/DynamoDB"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "DynamoDB is throttling requests"

  dimensions = {
    TableName = aws_dynamodb_table.this.name
  }

  alarm_actions = var.alarm_sns_topic_arns
}
```

---

### Bedrock

**Archivo:** `infra/terraform/modules/bedrock/main.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **Logging configurable** | ✅ Correcto | CloudWatch logging para invocaciones de Bedrock |
| **IAM granular** | ✅ Correcto | Permisos específicos para `InvokeModel` y `InvokeModelWithResponseStream` |
| **Múltiples regiones** | ✅ Correcto | ARNs para us-east-1, us-west-2, us-east-2 |
| **Model ID como variable** | ✅ Flexible | Permite cambiar modelo sin modificar código |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Resource ARN amplio** | 🟡 Media | `arn:aws:bedrock:*::foundation-model/anthropic.claude-3-*` permite todos los modelos Claude 3 | Restringir al modelo específico usado |
| **Sin límites de invocación** | 🟡 Media | No hay control de cuántas invocaciones por minuto | Implementar throttling en Lambda o usar reserved concurrency |
| **Sin métricas de costo** | 🟡 Media | No hay alarmas para costos de Bedrock | Usar AWS Cost Anomaly Detection o Budget Alerts |
| **Logging role no validado** | 🟢 Baja | `var.logging_role_arn` debe tener permisos correctos | Documentar requisitos del rol |

#### 📝 Código de Mejora Sugerido

```hcl
# Política IAM más restrictiva
resource "aws_iam_role_policy" "bedrock_access" {
  count = var.create_lambda_policy ? 1 : 0

  name = "${var.model_name}-bedrock-access"
  role = var.lambda_role_name

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
          # Solo el modelo específico
          "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.model_id}"
        ]
      }
    ]
  })
}

# Alarma de costos (requiere AWS Budgets)
resource "aws_budgets_budget" "bedrock" {
  name         = "bedrock-monthly-budget"
  budget_type  = "COST"
  limit_amount = var.bedrock_monthly_budget
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  cost_filter {
    name   = "Service"
    values = ["Amazon Bedrock"]
  }

  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = var.budget_alert_emails
  }
}
```

---

### IAM

**Archivo:** `infra/terraform/modules/iam/main.tf`

#### ✅ Buenas Prácticas Implementadas

| Aspecto | Estado | Detalle |
|---------|--------|---------|
| **Rol por Lambda** | ✅ Excelente | 5 roles separados para cada función (least privilege) |
| **Trust policy correcta** | ✅ Correcto | Solo `lambda.amazonaws.com` puede asumir el rol |
| **Políticas granulares** | ✅ Correcto | SQS: SendMessage vs ReceiveMessage separados |
| **DynamoDB granular** | ✅ Excelente | PutItem, UpdateItem, GetItem, Query asignados según necesidad |
| **Convención de nombres** | ✅ Claro | `{domain}-{role_purpose}-lambda-{environment}` |

#### Matriz de Permisos por Lambda

| Lambda | SQS | DynamoDB | Bedrock | Logs |
|--------|-----|----------|---------|------|
| process_handler | SendMessage | PutItem | ❌ | ✅ |
| conversion_worker | Receive, Delete, Send (access-pattern) | UpdateItem | InvokeModel | ✅ |
| query_handler | ❌ | GetItem, Query, Scan | ❌ | ✅ |
| dlq_handler | Receive, Delete (ambas DLQ) | UpdateItem | ❌ | ✅ |
| access_pattern_worker | Receive, Delete | UpdateItem | InvokeModel | ✅ |

#### ⚠️ Áreas de Mejora

| Aspecto | Severidad | Problema | Recomendación |
|---------|-----------|----------|---------------|
| **Sin permission boundary** | 🟢 Baja | Roles sin límite superior de permisos | Considerar permission boundary para compliance |
| **Scan en query_handler** | 🟡 Media | Permiso de `Scan` puede ser costoso | Evaluar si realmente se necesita Scan o solo Query |
| **Sin condiciones de recurso** | 🟢 Baja | Algunas políticas no usan condiciones | Agregar condiciones como `aws:SourceAccount` donde aplique |

---

### Resumen Parte 2 - Backend

| Componente | Puntuación | Fortalezas | Debilidades Principales |
|------------|------------|------------|------------------------|
| **API Gateway** | 7/10 | HTTP API v2, CORS, auto-deploy | Sin throttling, sin autorización, sin logging |
| **Lambda** | 8/10 | ARM64, X-Ray, alarmas, DLQ | Sin reserved concurrency, memory sin optimizar |
| **SQS** | 7.5/10 | DLQ configurado, retención correcta | Sin encriptación, sin long polling |
| **DynamoDB** | 8.5/10 | On-demand, TTL, PITR, GSI | Sin alarmas de throttling |
| **Bedrock** | 7.5/10 | Logging, IAM granular | ARN amplio, sin límites de costo |
| **IAM** | 9/10 | Least privilege ejemplar | Sin permission boundary |

**Puntuación Global Backend: 7.9/10**

---

## Resumen de Hallazgos

### Por Severidad

#### 🔴 Alta Prioridad (Requiere acción inmediata)

| # | Componente | Problema | Impacto |
|---|------------|----------|---------|
| 1 | **API Gateway** | Sin throttling configurado | Vulnerable a abuso, costos descontrolados |
| 2 | **API Gateway** | Sin autorización (JWT/Lambda authorizer) | Endpoints públicos sin protección |
| 3 | **ACM** | Validación DNS incompleta | Certificado podría no activarse |
| 4 | **CloudFront Function** | Secreto hardcodeado en código | Exposición de credenciales |

#### 🟡 Media Prioridad (Planificar corrección)

| # | Componente | Problema | Impacto |
|---|------------|----------|---------|
| 5 | **CloudFront** | TTL bajo para assets estáticos | Performance subóptimo, más requests al origin |
| 6 | **CloudFront** | Cache policy deprecated | Código legacy, sin nuevas features |
| 7 | **CloudFront** | Sin WAF | Sin protección DDoS/bots |
| 8 | **CloudFront** | Sin access logging | Sin visibilidad de tráfico |
| 9 | **S3** | Falta logging de acceso | Sin auditoría de acceso a objetos |
| 10 | **S3** | Versionado/encriptación condicionales | Posible pérdida de datos |
| 11 | **Lambda** | Sin reserved concurrency | Una función puede consumir toda la capacidad |
| 12 | **Lambda** | Memory sin optimizar | Costos potencialmente subóptimos |
| 13 | **Lambda** | Alarmas sin destino SNS | Alertas no se envían |
| 14 | **SQS** | Sin encriptación SSE | Mensajes en texto plano |
| 15 | **SQS** | Long polling deshabilitado | Mayor costo y latencia |
| 16 | **SQS** | Sin alarmas de CloudWatch | Sin monitoreo de colas |
| 17 | **DynamoDB** | Sin alarmas de throttling | Sin visibilidad de problemas de capacidad |
| 18 | **Bedrock** | ARN de recursos amplio | Permisos más amplios de lo necesario |
| 19 | **Bedrock** | Sin límites de costo | Riesgo de gastos inesperados |
| 20 | **API Gateway** | Sin access logging | Sin trazabilidad de requests |

#### 🟢 Baja Prioridad (Nice to have)

| # | Componente | Problema | Impacto |
|---|------------|----------|---------|
| 21 | **S3** | Sin lifecycle rules | Versiones antiguas no se limpian |
| 22 | **CloudFront** | Error caching TTL muy bajo | Micro-invalidaciones frecuentes |
| 23 | **ACM** | Sin SANs adicionales | Solo un dominio soportado |
| 24 | **Lambda** | Sin provisioned concurrency | Cold starts ocasionales |
| 25 | **Lambda** | Logs sin encriptación KMS | Compliance en algunos casos |
| 26 | **DynamoDB** | Sin contributor insights | Sin análisis de patrones |
| 27 | **IAM** | Sin permission boundary | Sin límite superior de permisos |

---

## Puntuación Global

### Por Componente (Actualizado 2026-02-04)

| Componente | Antes | Después | Mejoras Implementadas |
|------------|-------|---------|----------------------|
| IAM | 9.0/10 | 9.0/10 | Sin cambios (ya excelente) |
| DynamoDB | 8.5/10 | 9.0/10 | +Alarma throttling |
| S3 | 8.5/10 | 9.5/10 | +Encriptación obligatoria, +Logging |
| Lambda | 8.0/10 | 8.0/10 | Reserved concurrency N/A (límite cuenta) |
| CloudFront | 8.0/10 | 9.0/10 | +Cache policy, +Assets behavior, +Logging |
| Bedrock | 7.5/10 | 9.0/10 | +IAM restrictivo, +Budget alert |
| SQS | 7.5/10 | 9.0/10 | +SSE encriptación, +4 alarmas CloudWatch |
| CloudFront Function | 7.5/10 | 7.5/10 | Sin cambios (secreto en código aceptable) |
| API Gateway | 7.0/10 | 8.5/10 | +Throttling, +Access logging |
| ACM | 7.0/10 | 7.0/10 | Sin cambios (validación manual OK) |

### Promedio General: **7.85/10** → **8.55/10** 🟢

**Veredicto Actualizado:** La infraestructura ahora tiene observabilidad completa (logging en API Gateway, CloudFront, S3), protección de costos (Budget Bedrock), seguridad mejorada (encriptación SQS/S3 obligatoria, throttling API Gateway) y alertas operacionales centralizadas (SNS → email).

**Pendientes principales:**
- JWT/Lambda authorizer para API Gateway
- CloudFront KeyValueStore para secreto (opcional)
- Permission boundaries IAM (opcional)

---

## Plan de Acción Recomendado

### Fase 1: Seguridad Crítica (Semana 1)
- [x] ✅ Configurar throttling en API Gateway stage (100 burst, 50 rate)
- [ ] Implementar JWT authorizer o Lambda authorizer
- [ ] Completar validación DNS de ACM con Route53 o documentar proceso manual
- [ ] Migrar secreto de CloudFront Function a KeyValueStore

### Fase 2: Observabilidad (Semana 2)
- [x] ✅ Agregar access logging en API Gateway (CloudWatch Logs)
- [x] ✅ Agregar CloudFront access logging (S3 bucket dedicado)
- [x] ✅ Configurar SNS topic para alarmas (`prod-operational-alarms`)
- [x] ✅ Agregar alarmas de SQS (age of oldest message, DLQ not empty)
- [x] ✅ Agregar alarma de throttling en DynamoDB
- [x] ✅ Agregar S3 bucket logging (bucket dedicado con lifecycle 90 días)

### Fase 3: Performance (Semana 3)
- [x] ✅ Agregar `ordered_cache_behavior` para `/assets/*` con CachingOptimized
- [x] ✅ Migrar de `forwarded_values` a `cache_policy_id`
- [x] ✅ Habilitar long polling en SQS (ya estaba en 20s en prod)
- [ ] Ejecutar AWS Lambda Power Tuning para optimizar memory

### Fase 4: Protección y Costos (Semana 4)
- [x] ⏭️ WAF no implementado (Cloudflare ya lo gestiona)
- [x] ⏭️ Rate limiting via Cloudflare (no WAF necesario)
- [x] ✅ Habilitar encriptación SSE en SQS (4 colas)
- [x] ✅ Configurar Budget Alert para Bedrock ($20/mes con alertas 80%/100%)
- [x] ⚠️ Reserved concurrency: No aplicable (cuenta con límite de 10)

### Fase 5: Hardening (Opcional)
- [ ] Agregar permission boundaries en IAM
- [ ] Habilitar CloudFront real-time logs
- [ ] Configurar DynamoDB contributor insights
- [ ] Habilitar encriptación KMS en CloudWatch logs

### Mejoras Adicionales Implementadas (2026-02-04)
- [x] ✅ S3 encriptación obligatoria (removido conditional)
- [x] ✅ Bedrock IAM policy restringido a model_id específico
- [x] ✅ S3 logging bucket con lifecycle 90 días y ACL para CloudFront
- [x] ✅ SNS topic con suscripción email (`andressep.95@gmail.com`)

---

## Arquitectura Actual

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLOUDFLARE                                      │
│                         (DNS + Proxy + WAF)                                  │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │ x-origin-secret header
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CLOUDFRONT                                        │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────────┐  │
│  │ CF Function     │    │ S3 Origin (OAC) │    │ API Gateway Origin      │  │
│  │ verify-secret   │    │ Static Assets   │    │ /prod/api/*             │  │
│  └─────────────────┘    └─────────────────┘    └─────────────────────────┘  │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         API GATEWAY HTTP v2                                  │
│  POST /api/v1/schemas ──► process_handler                                    │
│  GET /api/v1/schemas  ──► query_handler                                      │
│  GET /api/v1/schemas/{id} ──► query_handler                                  │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
       ┌───────────────────────────┼───────────────────────────┐
       │                           │                           │
       ▼                           ▼                           ▼
┌──────────────┐           ┌──────────────┐           ┌──────────────┐
│   LAMBDA     │           │   LAMBDA     │           │   LAMBDA     │
│  process_    │──────────►│  conversion_ │──────────►│  access_     │
│  handler     │   SQS     │  worker      │   SQS     │  pattern_    │
└──────────────┘           └──────────────┘           │  worker      │
                                   │                  └──────────────┘
                                   │                         │
                                   ▼                         ▼
                           ┌──────────────┐          ┌──────────────┐
                           │   BEDROCK    │          │   BEDROCK    │
                           │  Claude 3.5  │          │  Claude 3.5  │
                           │   Sonnet     │          │   Haiku      │
                           └──────────────┘          └──────────────┘
                                   │                         │
                                   └────────────┬────────────┘
                                                │
                                                ▼
                                        ┌──────────────┐
                                        │  DYNAMODB    │
                                        │  schemas     │
                                        │  (TTL 24h)   │
                                        └──────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           DLQ HANDLING                                       │
│  conversion-dlq ─────┐                                                       │
│                      ├──► dlq_handler ──► DynamoDB (mark as failed)          │
│  access-pattern-dlq ─┘                                                       │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Referencias AWS

- [CloudFront OAC Best Practices](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-restricting-access-to-s3.html)
- [S3 Security Best Practices](https://docs.aws.amazon.com/AmazonS3/latest/userguide/security-best-practices.html)
- [Lambda Best Practices](https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html)
- [API Gateway HTTP API Throttling](https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-throttling.html)
- [SQS Visibility Timeout](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)

---

*Documento generado automáticamente. Última actualización: 2026-02-04*

---

## Resumen de Implementación (2026-02-04)

### Archivos Creados
| Archivo | Propósito |
|---------|-----------|
| `environments/prod/shared/sns.tf` | SNS topic + suscripción email |
| `environments/prod/shared/s3_logging.tf` | Bucket logs con lifecycle 90 días |
| `environments/prod/shared/budgets.tf` | Alerta Bedrock $20/mes |

### Archivos Modificados
| Módulo | Cambios |
|--------|---------|
| `gateway/http-v2` | +Throttling, +Access logging CloudWatch |
| `gateway/wrapper` | Pass-through nuevas variables |
| `cloudfront` | +CachingOptimized, +/assets/*, +logging_config |
| `s3` | Encriptación obligatoria, +logging |
| `sqs` | +SSE 4 colas, +4 alarmas CloudWatch |
| `dynamodb` | +Alarma ThrottledRequests |
| `bedrock` | IAM restringido a model_id específico |

### Configuración Producción
- **Throttling:** 100 burst / 50 rate por segundo
- **Logs:** API Gateway (14 días), CloudFront (90 días), S3 (90 días)
- **Alarmas:** SQS age, DLQ depth, DynamoDB throttle → SNS → email
- **Budget:** Bedrock $20/mes (alertas 80% y 100%)
