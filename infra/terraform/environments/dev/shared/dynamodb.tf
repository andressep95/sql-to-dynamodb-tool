# ============================================
# Shared Module - DynamoDB (Schemas Table)
# ============================================

module "schemas_table" {
  source = "../../../modules/dynamodb"

  table_name   = var.schemas_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "conversionId"

  # GSI attributes (conversionDate + createdAt for date-based queries)
  gsi_attributes = [
    { name = "conversionDate", type = "S" },
    { name = "createdAt", type = "S" }
  ]

  global_secondary_indexes = [
    {
      name            = "conversionDate-createdAt-index"
      hash_key        = "conversionDate"
      range_key       = "createdAt"
      projection_type = "ALL"
    }
  ]

  # TTL: 24 hours automatic cleanup
  ttl_attribute = "expiresAt"

  tags = var.common_tags
}

# ============================================
# IAM Policy - Lambda DynamoDB Write Access
# ============================================

resource "aws_iam_role_policy" "lambda_dynamodb_write" {
  name = "${var.environment}-lambda-dynamodb-write"
  role = var.lambda_role_name

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
