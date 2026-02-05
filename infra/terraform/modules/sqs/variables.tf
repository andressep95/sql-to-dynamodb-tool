# ============================================
# SQS Module - Variables
# ============================================

variable "queue_name" {
  description = "Base name for SQS queues"
  type        = string
}

variable "visibility_timeout_seconds" {
  description = "Visibility timeout for conversion queue"
  type        = number
  default     = 180 # 3 minutes for conversion
}

variable "access_pattern_visibility_timeout" {
  description = "Visibility timeout for access pattern queue"
  type        = number
  default     = 120 # 2 minutes for access patterns
}

variable "message_retention_seconds" {
  description = "Message retention period in seconds"
  type        = number
  default     = 345600 # 4 days
}

variable "receive_wait_time_seconds" {
  description = "Long polling wait time in seconds"
  type        = number
  default     = 0
}

variable "dlq_retention_seconds" {
  description = "DLQ message retention period in seconds"
  type        = number
  default     = 1209600 # 14 days
}

variable "max_receive_count" {
  description = "Max receive count before sending to DLQ"
  type        = number
  default     = 3
}

variable "tags" {
  description = "Tags for resources"
  type        = map(string)
  default     = {}
}

# ============================================
# CloudWatch Alarms Configuration
# ============================================

variable "create_alarms" {
  description = "Create CloudWatch alarms for SQS queues"
  type        = bool
  default     = false
}

variable "alarm_sns_topic_arns" {
  description = "List of SNS topic ARNs to notify when alarms trigger"
  type        = list(string)
  default     = []
}

variable "max_message_age_seconds" {
  description = "Maximum age of oldest message in queue before alarming (seconds)"
  type        = number
  default     = 3600 # 1 hour
}
