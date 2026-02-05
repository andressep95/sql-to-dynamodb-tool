# ============================================
# SQS Module - Multiple Queues Support
# ============================================

# Conversion Queue (existing)
resource "aws_sqs_queue" "conversion_dlq" {
  name                      = "${var.queue_name}-dlq"
  message_retention_seconds = var.dlq_retention_seconds

  # Enable SSE with SQS-managed encryption
  sqs_managed_sse_enabled = true

  tags = var.tags
}

resource "aws_sqs_queue" "conversion_queue" {
  name                       = var.queue_name
  visibility_timeout_seconds = var.visibility_timeout_seconds
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds

  # Enable SSE with SQS-managed encryption
  sqs_managed_sse_enabled = true

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

  # Enable SSE with SQS-managed encryption
  sqs_managed_sse_enabled = true

  tags = var.tags
}

resource "aws_sqs_queue" "access_pattern_queue" {
  name                       = "${var.queue_name}-access-pattern"
  visibility_timeout_seconds = var.access_pattern_visibility_timeout
  message_retention_seconds  = var.message_retention_seconds
  receive_wait_time_seconds  = var.receive_wait_time_seconds

  # Enable SSE with SQS-managed encryption
  sqs_managed_sse_enabled = true

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.access_pattern_dlq.arn
    maxReceiveCount     = var.max_receive_count
  })

  tags = var.tags
}

# ============================================
# CloudWatch Alarms (conditional)
# ============================================

# Alarm: Conversion Queue - Message Age
resource "aws_cloudwatch_metric_alarm" "conversion_queue_age" {
  count = var.create_alarms ? 1 : 0

  alarm_name          = "${var.queue_name}-message-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Maximum"
  threshold           = var.max_message_age_seconds
  alarm_description   = "Conversion queue has messages older than ${var.max_message_age_seconds} seconds"
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.conversion_queue.name
  }

  alarm_actions = var.alarm_sns_topic_arns
  ok_actions    = var.alarm_sns_topic_arns

  tags = var.tags
}

# Alarm: Access Pattern Queue - Message Age
resource "aws_cloudwatch_metric_alarm" "access_pattern_queue_age" {
  count = var.create_alarms ? 1 : 0

  alarm_name          = "${var.queue_name}-access-pattern-message-age"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "ApproximateAgeOfOldestMessage"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Maximum"
  threshold           = var.max_message_age_seconds
  alarm_description   = "Access pattern queue has messages older than ${var.max_message_age_seconds} seconds"
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.access_pattern_queue.name
  }

  alarm_actions = var.alarm_sns_topic_arns
  ok_actions    = var.alarm_sns_topic_arns

  tags = var.tags
}

# Alarm: Conversion DLQ - Messages Visible
resource "aws_cloudwatch_metric_alarm" "conversion_dlq_messages" {
  count = var.create_alarms ? 1 : 0

  alarm_name          = "${var.queue_name}-dlq-messages"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Conversion DLQ has messages - failed conversions detected"
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.conversion_dlq.name
  }

  alarm_actions = var.alarm_sns_topic_arns
  ok_actions    = var.alarm_sns_topic_arns

  tags = var.tags
}

# Alarm: Access Pattern DLQ - Messages Visible
resource "aws_cloudwatch_metric_alarm" "access_pattern_dlq_messages" {
  count = var.create_alarms ? 1 : 0

  alarm_name          = "${var.queue_name}-access-pattern-dlq-messages"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Access pattern DLQ has messages - failed access pattern processing detected"
  treat_missing_data  = "notBreaching"

  dimensions = {
    QueueName = aws_sqs_queue.access_pattern_dlq.name
  }

  alarm_actions = var.alarm_sns_topic_arns
  ok_actions    = var.alarm_sns_topic_arns

  tags = var.tags
}
