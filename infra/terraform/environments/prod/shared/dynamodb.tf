# ============================================
# Shared Module - DynamoDB (Schemas Table) - Production
# ============================================

module "schemas_table" {
  source = "../../../modules/dynamodb"

  table_name   = var.schemas_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "conversionId"

  # GSI attributes
  gsi_attributes = [
    { name = "conversionDate", type = "S" },
    { name = "createdAt", type = "S" },
    { name = "tenantId", type = "S" },
  ]

  global_secondary_indexes = [
    {
      name            = "conversionDate-createdAt-index"
      hash_key        = "conversionDate"
      range_key       = "createdAt"
      projection_type = "ALL"
    },
    {
      name            = "tenantId-createdAt-index"
      hash_key        = "tenantId"
      range_key       = "createdAt"
      projection_type = "ALL"
    },
  ]

  # TTL deshabilitado - usuario elimina manualmente
  ttl_attribute = null

  # Production: enable point-in-time recovery
  point_in_time_recovery = true

  # CloudWatch throttle alarm
  create_throttle_alarm = true
  alarm_sns_topic_arns  = [aws_sns_topic.alarms.arn]

  tags = var.common_tags
}

# ============================================
# Invitations Table - Códigos de invitación temporales
# ============================================

module "invitations_table" {
  source = "../../../modules/dynamodb"

  table_name   = "${var.environment}-invitations"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "invitationCode"

  # TTL para expiración automática (24 horas)
  ttl_attribute = "expiresAt"

  point_in_time_recovery = false
  create_throttle_alarm  = false

  tags = var.common_tags
}

# ============================================
# IAM Policy - process_handler: DynamoDB PutItem
# ============================================

resource "aws_iam_role_policy" "process_handler_dynamodb" {
  name = "${var.environment}-process-handler-dynamodb"
  role = var.process_handler_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem"
        ]
        Resource = [
          module.schemas_table.table_arn
        ]
      }
    ]
  })
}

# ============================================
# IAM Policy - conversion_worker: DynamoDB UpdateItem
# ============================================

resource "aws_iam_role_policy" "conversion_worker_dynamodb" {
  name = "${var.environment}-conversion-worker-dynamodb"
  role = var.conversion_worker_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:UpdateItem"
        ]
        Resource = [
          module.schemas_table.table_arn
        ]
      }
    ]
  })
}

# ============================================
# IAM Policy - query_handler: DynamoDB Read Access
# ============================================

resource "aws_iam_role_policy" "query_handler_dynamodb" {
  name = "${var.environment}-query-handler-dynamodb"
  role = var.query_handler_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          module.schemas_table.table_arn,
          "${module.schemas_table.table_arn}/index/*"
        ]
      }
    ]
  })
}

# ============================================
# IAM Policy - dlq_handler: DynamoDB UpdateItem
# ============================================

resource "aws_iam_role_policy" "dlq_handler_dynamodb" {
  name = "${var.environment}-dlq-handler-dynamodb"
  role = var.dlq_handler_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:UpdateItem"
        ]
        Resource = [
          module.schemas_table.table_arn
        ]
      }
    ]
  })
}

# ============================================
# IAM Policy - access_pattern_worker: DynamoDB UpdateItem
# ============================================

resource "aws_iam_role_policy" "access_pattern_worker_dynamodb" {
  name = "${var.environment}-access-pattern-worker-dynamodb"
  role = var.access_pattern_worker_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:UpdateItem"
        ]
        Resource = [
          module.schemas_table.table_arn
        ]
      }
    ]
  })
}
