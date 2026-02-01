# ============================================
# Shared Module - Variables (Production)
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

variable "query_handler_function_name" {
  description = "Name of the query handler Lambda function"
  type        = string
}

variable "query_handler_function_arn" {
  description = "ARN of the query handler Lambda function"
  type        = string
}

variable "lambda_invoke_arn" {
  description = "Invoke ARN of the process handler Lambda function"
  type        = string
}

variable "query_handler_invoke_arn" {
  description = "Invoke ARN of the query handler Lambda function"
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
  default     = "prod"
}

# Lambda IAM roles (per-Lambda, least privilege)
variable "process_handler_role_name" {
  description = "Name of the process_handler IAM role for policy attachments"
  type        = string
}

variable "conversion_worker_role_name" {
  description = "Name of the conversion_worker IAM role for policy attachments"
  type        = string
}

variable "query_handler_role_name" {
  description = "Name of the query_handler IAM role for policy attachments"
  type        = string
}

variable "dlq_handler_role_name" {
  description = "Name of the dlq_handler IAM role for policy attachments"
  type        = string
}

# SQS
variable "sqs_queue_arn" {
  description = "ARN of the SQS conversion queue (for IAM policies)"
  type        = string
}

variable "sqs_dlq_arn" {
  description = "ARN of the SQS dead-letter queue (for IAM policies)"
  type        = string
}

# DynamoDB
variable "schemas_table_name" {
  description = "Name for the schemas DynamoDB table"
  type        = string
}
