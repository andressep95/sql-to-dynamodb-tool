# ============================================
# Shared Module - SQS IAM Policies (Production)
# ============================================

# IAM Policy - process_handler: SQS Send
resource "aws_iam_role_policy" "lambda_sqs_send" {
  name = "${var.environment}-lambda-sqs-send"
  role = var.process_handler_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage"
        ]
        Resource = [
          var.sqs_queue_arn
        ]
      }
    ]
  })
}

# IAM Policy - conversion_worker: SQS Consume
resource "aws_iam_role_policy" "lambda_sqs_consume" {
  name = "${var.environment}-lambda-sqs-consume"
  role = var.conversion_worker_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          var.sqs_queue_arn
        ]
      }
    ]
  })
}

# IAM Policy - dlq_handler: SQS Consume from DLQ
resource "aws_iam_role_policy" "lambda_sqs_dlq_consume" {
  name = "${var.environment}-lambda-sqs-dlq-consume"
  role = var.dlq_handler_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          var.sqs_dlq_arn
        ]
      }
    ]
  })
}
