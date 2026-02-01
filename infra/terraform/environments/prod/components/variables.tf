# ============================================
# Components Module - Variables (Production)
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

variable "process_handler_role_arn" {
  description = "ARN of the IAM role for process_handler Lambda"
  type        = string
}

variable "conversion_worker_role_arn" {
  description = "ARN of the IAM role for conversion_worker Lambda"
  type        = string
}

variable "query_handler_role_arn" {
  description = "ARN of the IAM role for query_handler Lambda"
  type        = string
}

variable "dlq_handler_role_arn" {
  description = "ARN of the IAM role for dlq_handler Lambda"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB schemas table"
  type        = string
  default     = ""
}

variable "sqs_queue_url" {
  description = "URL of the SQS conversion queue"
  type        = string
  default     = ""
}

variable "sqs_queue_arn" {
  description = "ARN of the SQS conversion queue"
  type        = string
  default     = ""
}

variable "sqs_dlq_arn" {
  description = "ARN of the SQS dead-letter queue"
  type        = string
  default     = ""
}
