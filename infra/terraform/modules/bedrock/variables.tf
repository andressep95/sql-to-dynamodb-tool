variable "model_name" {
  description = "Name identifier for the Bedrock model"
  type        = string
}

variable "model_id" {
  description = "Bedrock foundation model ID"
  type        = string
  
  validation {
    condition = can(regex("^[a-zA-Z0-9.-]+$", var.model_id))
    error_message = "Model ID must contain only alphanumeric characters, dots, and hyphens."
  }
}

variable "model_provider" {
  description = "Model provider (e.g., anthropic, amazon, cohere)"
  type        = string
  default     = "anthropic"
  
  validation {
    condition = contains(["anthropic", "amazon", "cohere", "ai21", "stability"], var.model_provider)
    error_message = "Model provider must be one of: anthropic, amazon, cohere, ai21, stability."
  }
}

variable "enable_logging" {
  description = "Enable CloudWatch logging for Bedrock invocations"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 7
  
  validation {
    condition = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.log_retention_days)
    error_message = "Log retention days must be a valid CloudWatch retention value."
  }
}

variable "logging_role_arn" {
  description = "IAM role ARN for Bedrock logging (required if enable_logging is true)"
  type        = string
  default     = null
}

variable "create_lambda_policy" {
  description = "Create IAM policy for Lambda to access Bedrock"
  type        = bool
  default     = true
}

variable "lambda_role_name" {
  description = "Lambda role name to attach Bedrock policy (required if create_lambda_policy is true)"
  type        = string
  default     = null
}

variable "skip_model_validation" {
  description = "Skip Bedrock API calls for model validation (required for LocalStack)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}