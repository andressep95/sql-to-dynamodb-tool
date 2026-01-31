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
  description = "Map of routes (route_key => config)"
  type = map(object({
    lambda_arn  = optional(string)
    lambda_name = optional(string)
  }))

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
