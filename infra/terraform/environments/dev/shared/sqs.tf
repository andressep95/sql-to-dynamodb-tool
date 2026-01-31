# ============================================
# Shared Module - SQS IAM Policies
# ============================================

# IAM Policy - Lambda SQS Send (process_handler)
resource "aws_iam_role_policy" "lambda_sqs_send" {
  name = "${var.environment}-lambda-sqs-send"
  role = var.lambda_role_name

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

# IAM Policy - Lambda SQS Consume (conversion_worker)
resource "aws_iam_role_policy" "lambda_sqs_consume" {
  name = "${var.environment}-lambda-sqs-consume"
  role = var.lambda_role_name

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
