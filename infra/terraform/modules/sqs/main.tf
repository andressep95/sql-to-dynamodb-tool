# ============================================
# SQS Module - Multiple Queues Support
# ============================================

# Conversion Queue (existing)
resource "aws_sqs_queue" "conversion_dlq" {
  name                      = "${var.queue_name}-dlq"
  message_retention_seconds = var.dlq_retention_seconds
  tags = var.tags
}

resource "aws_sqs_queue" "conversion_queue" {
  name                       = var.queue_name
  visibility_timeout_seconds = var.visibility_timeout_seconds
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.conversion_dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = var.tags
}

# Access Pattern Queue (new)
resource "aws_sqs_queue" "access_pattern_dlq" {
  name                      = "${var.queue_name}-access-pattern-dlq"
  message_retention_seconds = var.dlq_retention_seconds
  tags = var.tags
}

resource "aws_sqs_queue" "access_pattern_queue" {
  name                       = "${var.queue_name}-access-pattern"
  visibility_timeout_seconds = var.access_pattern_visibility_timeout
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.access_pattern_dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = var.tags
}
