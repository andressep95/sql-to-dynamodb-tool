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
  default     = "arn:aws:iam::000000000000:role/lambda-role" # LocalStack dummy
}
