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
