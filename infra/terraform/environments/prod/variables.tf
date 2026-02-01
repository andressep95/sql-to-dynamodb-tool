# ============================================
# Global Variables - Production Environment
# ============================================

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "idp"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "aws_profile" {
  description = "AWS CLI profile to use"
  type        = string
  default     = "default"
}

# Production-specific configuration
variable "enable_deletion_protection" {
  description = "Enable deletion protection for resources"
  type        = bool
  default     = true
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
