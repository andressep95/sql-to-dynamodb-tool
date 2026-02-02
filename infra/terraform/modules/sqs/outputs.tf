# ============================================
# SQS Module - Outputs
# ============================================

# Conversion Queue Outputs
output "conversion_queue_url" {
  description = "URL of the conversion SQS queue"
  value       = aws_sqs_queue.conversion_queue.url
}

output "conversion_queue_arn" {
  description = "ARN of the conversion SQS queue"
  value       = aws_sqs_queue.conversion_queue.arn
}

output "conversion_dlq_url" {
  description = "URL of the conversion dead letter queue"
  value       = aws_sqs_queue.conversion_dlq.url
}

output "conversion_dlq_arn" {
  description = "ARN of the conversion dead letter queue"
  value       = aws_sqs_queue.conversion_dlq.arn
}

# Access Pattern Queue Outputs
output "access_pattern_queue_url" {
  description = "URL of the access pattern SQS queue"
  value       = aws_sqs_queue.access_pattern_queue.url
}

output "access_pattern_queue_arn" {
  description = "ARN of the access pattern SQS queue"
  value       = aws_sqs_queue.access_pattern_queue.arn
}

output "access_pattern_dlq_url" {
  description = "URL of the access pattern dead letter queue"
  value       = aws_sqs_queue.access_pattern_dlq.url
}

output "access_pattern_dlq_arn" {
  description = "ARN of the access pattern dead letter queue"
  value       = aws_sqs_queue.access_pattern_dlq.arn
}

# Legacy outputs for backward compatibility
output "queue_url" {
  description = "URL of the main SQS queue (legacy)"
  value       = aws_sqs_queue.conversion_queue.url
}

output "queue_arn" {
  description = "ARN of the main SQS queue (legacy)"
  value       = aws_sqs_queue.conversion_queue.arn
}
