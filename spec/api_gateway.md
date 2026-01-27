# API Gateway - Guía de Construcción IaC

## Introducción

Esta guía documenta el proceso de construcción de la infraestructura de API Gateway utilizando Terraform. La arquitectura está diseñada con un patrón modular que permite:

- **LocalStack (desarrollo)**: Usa REST API v1 (única versión soportada)
- **AWS (producción)**: Usará HTTP API v2 (más moderno y económico)

## Estructura del Proyecto

```
infra/terraform/
├── modules/
│   └── gateway/
│       ├── rest-v1/        # Módulo REST API v1 (LocalStack)
│       ├── http-v2/        # Módulo HTTP API v2 (AWS futuro)
│       └── wrapper/        # Wrapper unificador
└── environments/
    └── dev/
        └── shared/
            └── api_gateway.tf  # Especificación del environment
```

---

## 1. Módulos Base

### 1.1 REST API v1 (LocalStack)

Este módulo implementa REST API Gateway v1, compatible con LocalStack.

#### `modules/gateway/rest-v1/variables.tf`

```hcl
variable "name" {
  type        = string
  description = "Nombre del API Gateway"
}

variable "region" {
  type        = string
  description = "AWS region"
}

variable "lambda_arn" {
  type        = string
  description = "ARN de la Lambda a integrar"
}

variable "lambda_name" {
  type        = string
  description = "Nombre de la Lambda (para permisos)"
}

variable "stage_name" {
  type    = string
  default = "dev"
}

variable "routes" {
  description = "Map of routes (route_key => {}). Format: 'METHOD /path'"
  type        = map(any)
  default = {
    "GET /" = {}
  }
}
```

| Variable | Tipo | Requerido | Descripción |
|----------|------|-----------|-------------|
| `name` | string | Sí | Nombre del API Gateway |
| `region` | string | Sí | Región AWS |
| `lambda_arn` | string | Sí | ARN de la función Lambda |
| `lambda_name` | string | Sí | Nombre de la Lambda para permisos |
| `stage_name` | string | No | Stage name (default: `dev`) |
| `routes` | map(any) | No | Rutas en formato `"METHOD /path"` |

#### `modules/gateway/rest-v1/main.tf`

**Recursos creados:**

1. **Locals** - Parseo de rutas `METHOD /path` a objetos
2. **`aws_api_gateway_rest_api`** - API REST principal
3. **`aws_api_gateway_resource`** - Recursos para paths no-root
4. **`aws_api_gateway_method`** - Métodos HTTP por ruta
5. **`aws_api_gateway_integration`** - Integración Lambda Proxy
6. **`aws_api_gateway_deployment`** - Deployment con triggers
7. **`aws_api_gateway_stage`** - Stage del API
8. **`aws_lambda_permission`** - Permiso para invocar Lambda

**Lógica de parseo de rutas:**

```hcl
locals {
  # "GET /api/convert" => {method: "GET", path: "/api/convert"}
  parsed_routes = {
    for route_key, route_config in var.routes : route_key => {
      method = split(" ", route_key)[0]
      path   = split(" ", route_key)[1]
    }
  }
}
```

#### `modules/gateway/rest-v1/outputs.tf`

```hcl
output "api_id" {
  value = aws_api_gateway_rest_api.this.id
}

output "stage_name" {
  value = aws_api_gateway_stage.this.stage_name
}

output "invoke_url" {
  value = "https://${aws_api_gateway_rest_api.this.id}.execute-api.${var.region}.amazonaws.com/${var.stage_name}/"
}
```

| Output | Descripción |
|--------|-------------|
| `api_id` | ID del API Gateway |
| `stage_name` | Nombre del stage |
| `invoke_url` | URL de invocación completa |

---

### 1.2 HTTP API v2 (AWS Producción)

Este módulo implementa HTTP API v2, más moderno y eficiente para AWS.

#### `modules/gateway/http-v2/variables.tf`

```hcl
variable "name" {
  description = "API Gateway name"
  type        = string
}

variable "stage_name" {
  description = "Stage name"
  type        = string
  default     = "$default"
}

variable "lambda_name" {
  description = "Lambda function name"
  type        = string
}

variable "lambda_invoke_arn" {
  description = "Lambda invoke ARN"
  type        = string
}

variable "routes" {
  description = "Map of routes (route_key => {})"
  type        = map(any)
  default = {
    "GET /"     = {}
    "POST /api" = {}
  }
}

variable "cors_allow_origins" {
  type    = list(string)
  default = ["*"]
}

variable "cors_allow_methods" {
  type    = list(string)
  default = ["GET", "POST", "OPTIONS"]
}

variable "cors_allow_headers" {
  type    = list(string)
  default = ["*"]
}

variable "tags" {
  type    = map(string)
  default = {}
}
```

| Variable | Tipo | Requerido | Descripción |
|----------|------|-----------|-------------|
| `name` | string | Sí | Nombre del API |
| `lambda_name` | string | Sí | Nombre de la Lambda |
| `lambda_invoke_arn` | string | Sí | ARN de invocación Lambda |
| `stage_name` | string | No | Stage (default: `$default`) |
| `routes` | map(any) | No | Rutas del API |
| `cors_allow_origins` | list(string) | No | Origins CORS permitidos |
| `cors_allow_methods` | list(string) | No | Métodos CORS permitidos |
| `cors_allow_headers` | list(string) | No | Headers CORS permitidos |
| `tags` | map(string) | No | Tags del recurso |

#### `modules/gateway/http-v2/main.tf`

**Recursos creados:**

1. **`aws_apigatewayv2_api`** - HTTP API con CORS
2. **`aws_apigatewayv2_integration`** - Integración Lambda Proxy v2
3. **`aws_apigatewayv2_route`** - Rutas dinámicas
4. **`aws_apigatewayv2_stage`** - Stage con auto_deploy
5. **`aws_lambda_permission`** - Permiso de invocación

**Diferencias clave con REST v1:**

- Usa `payload_format_version = "2.0"`
- Configuración CORS nativa
- `auto_deploy = true` en stage
- Rutas más simples (sin recursos intermedios)

#### `modules/gateway/http-v2/outputs.tf`

```hcl
output "api_id" {
  value = aws_apigatewayv2_api.this.id
}

output "api_endpoint" {
  value = aws_apigatewayv2_api.this.api_endpoint
}

output "execution_arn" {
  value = aws_apigatewayv2_api.this.execution_arn
}
```

| Output | Descripción |
|--------|-------------|
| `api_id` | ID del HTTP API |
| `api_endpoint` | Endpoint del API |
| `execution_arn` | ARN de ejecución |

---

## 2. Wrapper (Módulo Unificador)

El wrapper actúa como interfaz única para ambos tipos de gateway.

### `modules/gateway/wrapper/variables.tf`

```hcl
variable "gateway_type" {
  description = "API Gateway type: rest-v1 or http-v2"
  type        = string

  validation {
    condition     = contains(["rest-v1", "http-v2"], var.gateway_type)
    error_message = "gateway_type must be 'rest-v1' or 'http-v2'"
  }
}

variable "region" {
  description = "AWS region"
  type        = string
}

variable "name" {
  description = "API Gateway name"
  type        = string
}

variable "stage_name" {
  description = "Stage name"
  type        = string
  default     = "dev"
}

variable "lambda_name" {
  description = "Lambda function name"
  type        = string
}

# REST v1 usa ARN normal
variable "lambda_arn" {
  description = "Lambda ARN (REST v1)"
  type        = string
  default     = null
}

# HTTP v2 usa invoke ARN
variable "lambda_invoke_arn" {
  description = "Lambda invoke ARN (HTTP v2)"
  type        = string
  default     = null
}

variable "routes" {
  description = "Routes for HTTP API v2"
  type        = map(any)
  default     = {}
}

variable "tags" {
  type    = map(string)
  default = {}
}
```

| Variable | Tipo | Requerido | Descripción |
|----------|------|-----------|-------------|
| `gateway_type` | string | Sí | `rest-v1` o `http-v2` |
| `region` | string | Sí | Región AWS |
| `name` | string | Sí | Nombre del gateway |
| `lambda_name` | string | Sí | Nombre de la Lambda |
| `lambda_arn` | string | Condicional | ARN Lambda (solo REST v1) |
| `lambda_invoke_arn` | string | Condicional | Invoke ARN (solo HTTP v2) |
| `stage_name` | string | No | Stage name |
| `routes` | map(any) | No | Configuración de rutas |
| `tags` | map(string) | No | Tags |

### `modules/gateway/wrapper/main.tf`

```hcl
locals {
  is_rest_v1 = var.gateway_type == "rest-v1"
  is_http_v2 = var.gateway_type == "http-v2"
}

# REST API Gateway v1
module "rest_v1" {
  count  = local.is_rest_v1 ? 1 : 0
  source = "../rest-v1"

  name       = var.name
  region     = var.region
  stage_name = var.stage_name

  lambda_name = var.lambda_name
  lambda_arn  = var.lambda_arn

  routes = var.routes
}

# HTTP API Gateway v2
module "http_v2" {
  count  = local.is_http_v2 ? 1 : 0
  source = "../http-v2"

  name       = var.name
  stage_name = var.stage_name

  lambda_name       = var.lambda_name
  lambda_invoke_arn = var.lambda_invoke_arn

  routes = var.routes
  tags   = var.tags
}
```

**Patrón utilizado:** Conditional module instantiation con `count`

### `modules/gateway/wrapper/outputs.tf`

```hcl
output "api_id" {
  value = local.is_rest_v1 ? module.rest_v1[0].api_id : module.http_v2[0].api_id
}

output "api_endpoint" {
  value = local.is_rest_v1 ? module.rest_v1[0].invoke_url : module.http_v2[0].api_endpoint
}

output "execution_arn" {
  value = local.is_http_v2 ? module.http_v2[0].execution_arn : null
}
```

| Output | Descripción |
|--------|-------------|
| `api_id` | ID del API (unificado) |
| `api_endpoint` | URL del endpoint (unificado) |
| `execution_arn` | ARN de ejecución (solo HTTP v2) |

---

## 3. Especificación en Environment

### `environments/dev/shared/variables.tf`

```hcl
variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

# Lambda integration
variable "lambda_function_name" {
  description = "Name of the Lambda function to integrate with API Gateway"
  type        = string
}

variable "lambda_function_arn" {
  description = "ARN of the Lambda function"
  type        = string
}

# API Gateway configuration
variable "api_gateway_name" {
  description = "Name for the API Gateway"
  type        = string
  default     = "process_handler-api"
}

variable "stage_name" {
  description = "Stage name for API Gateway"
  type        = string
  default     = "dev"
}
```

### `environments/dev/shared/api_gateway.tf`

```hcl
module "api_gateway" {
  source = "../../../modules/gateway/wrapper"

  gateway_type = "rest-v1"  # LocalStack compatible

  name       = var.api_gateway_name
  region     = var.aws_region
  stage_name = var.stage_name

  # Lambda integration
  lambda_name = var.lambda_function_name
  lambda_arn  = var.lambda_function_arn

  # Routes configuration
  routes = {
    "GET /" = {}
  }

  tags = var.common_tags
}
```

### `environments/dev/shared/outputs.tf`

```hcl
output "api_gateway_id" {
  description = "ID of the API Gateway"
  value       = module.api_gateway.api_id
}

output "api_gateway_endpoint" {
  description = "Invoke URL of the API Gateway"
  value       = module.api_gateway.api_endpoint
}
```

---

## 4. Orden de Construcción

### Paso 1: Crear módulo REST v1
```
modules/gateway/rest-v1/
├── variables.tf  # Definir inputs
├── main.tf       # Implementar recursos
└── outputs.tf    # Exponer valores
```

### Paso 2: Crear módulo HTTP v2
```
modules/gateway/http-v2/
├── variables.tf  # Definir inputs (incluye CORS)
├── main.tf       # Implementar recursos v2
└── outputs.tf    # Exponer valores
```

### Paso 3: Crear wrapper
```
modules/gateway/wrapper/
├── variables.tf  # Unificar variables de ambos
├── main.tf       # Lógica condicional
└── outputs.tf    # Outputs unificados
```

### Paso 4: Especificación en environment
```
environments/dev/shared/
├── variables.tf      # Variables del environment
├── api_gateway.tf    # Instanciar módulo wrapper
└── outputs.tf        # Outputs del environment
```

---

## 5. Migración LocalStack → AWS

Para migrar de LocalStack a AWS, solo se necesita cambiar:

```hcl
# Antes (LocalStack)
module "api_gateway" {
  source = "../../../modules/gateway/wrapper"

  gateway_type = "rest-v1"
  lambda_arn   = var.lambda_function_arn
  # ...
}

# Después (AWS)
module "api_gateway" {
  source = "../../../modules/gateway/wrapper"

  gateway_type      = "http-v2"
  lambda_invoke_arn = var.lambda_function_invoke_arn
  # ...
}
```

---

## 6. Notas Técnicas

### REST v1 vs HTTP v2

| Característica | REST v1 | HTTP v2 |
|----------------|---------|---------|
| Soporte LocalStack | ✅ | ❌ |
| CORS nativo | ❌ | ✅ |
| Auto deploy | ❌ | ✅ |
| Payload format | 1.0 | 2.0 |
| Costo | Mayor | Menor |
| Recursos intermedios | Requeridos | No requeridos |

### Formato de Rutas

Ambos módulos aceptan rutas en formato:
```hcl
routes = {
  "GET /"           = {}
  "POST /api"       = {}
  "GET /api/status" = {}
}
```

### Lambda ARN vs Invoke ARN

- **REST v1**: Usa `lambda_arn` (formato: `arn:aws:lambda:region:account:function:name`)
- **HTTP v2**: Usa `lambda_invoke_arn` (formato: `arn:aws:apigateway:region:lambda:path/...`)
