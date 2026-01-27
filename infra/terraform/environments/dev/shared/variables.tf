# ============================================
# Shared Module - Variables
# ============================================

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
