# ============================================
# Bedrock - Claude Sonnet 4 for SQL to DynamoDB conversion
# ============================================

module "bedrock" {
  source = "../../../modules/bedrock"

  model_name     = "claude-sonnet-4"
  model_id       = "anthropic.claude-sonnet-4-20250514-v1:0"
  model_provider = "anthropic"

  enable_logging        = false # LocalStack does not support Bedrock logging
  skip_model_validation = true  # LocalStack does not support Bedrock API
  create_lambda_policy = true
  lambda_role_name     = var.lambda_role_name

  tags = var.common_tags
}
