# ============================================
# Global Variables - Development Environment
# ============================================

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "idp"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

# LocalStack configuration
variable "use_localstack" {
  description = "Use LocalStack for local development"
  type        = bool
  default     = true
}

variable "localstack_endpoint" {
  description = "LocalStack endpoint URL"
  type        = string
  default     = "http://localhost:4566"
}

variable "localstack_lambda_endpoint" {
  description = "LocalStack endpoint URL reachable from inside Lambda containers"
  type        = string
  default     = "http://host.docker.internal:4566"
}

# Dev-specific configuration
variable "enable_deletion_protection" {
  description = "Enable deletion protection for resources"
  type        = bool
  default     = false
}

variable "log_level" {
  description = "Log level for Lambda functions"
  type        = string
  default     = "INFO"

  validation {
    condition     = contains(["DEBUG", "INFO", "WARN", "ERROR"], var.log_level)
    error_message = "Log level must be DEBUG, INFO, WARN, or ERROR"
  }
}

# DynamoDB table names
variable "schemas_table_name" {
  description = "Name for the schemas DynamoDB table"
  type        = string
  default     = "schemas"
}

# Common tags
variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}
