# ============================================
# Components Module - Variables
# ============================================

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default     = {}
}

variable "role_arn" {
  description = "ARN of the IAM role for Lambda functions"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB schemas table"
  type        = string
  default     = ""
}

variable "dynamodb_endpoint" {
  description = "DynamoDB endpoint URL (for LocalStack)"
  type        = string
  default     = ""
}
