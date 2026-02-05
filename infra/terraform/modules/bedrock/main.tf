# ============================================
# Bedrock Module - Foundation Model Access
# ============================================

# Data source para obtener modelos disponibles (skipped in LocalStack)
data "aws_bedrock_foundation_models" "available" {
  count       = var.skip_model_validation ? 0 : 1
  by_provider = var.model_provider
}

# Configuración de acceso al modelo
resource "aws_bedrock_model_invocation_logging_configuration" "this" {
  count = var.enable_logging ? 1 : 0

  logging_config {
    cloudwatch_config {
      log_group_name = aws_cloudwatch_log_group.bedrock[0].name
      role_arn       = var.logging_role_arn
    }
  }
}

# CloudWatch Log Group para Bedrock
resource "aws_cloudwatch_log_group" "bedrock" {
  count = var.enable_logging ? 1 : 0

  name              = "/aws/bedrock/${var.model_name}"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# IAM Role para Lambda acceder a Bedrock
locals {
  # Strip "us." prefix from model_id for foundation model ARN
  # e.g., "us.anthropic.claude-3-5-sonnet-20241022-v2:0" -> "anthropic.claude-3-5-sonnet-20241022-v2:0"
  foundation_model_id = replace(var.model_id, "/^us\\./", "")
}

resource "aws_iam_role_policy" "bedrock_access" {
  count = var.create_lambda_policy ? 1 : 0

  name = "${var.model_name}-bedrock-access"
  role = var.lambda_role_name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Resource = [
          # Inference profiles for specific model (cross-region)
          "arn:aws:bedrock:us-east-1:${data.aws_caller_identity.current.account_id}:inference-profile/${var.model_id}",
          "arn:aws:bedrock:us-west-2:${data.aws_caller_identity.current.account_id}:inference-profile/${var.model_id}",
          "arn:aws:bedrock:us-east-2:${data.aws_caller_identity.current.account_id}:inference-profile/${var.model_id}",
          # Foundation model (without us. prefix, any region)
          "arn:aws:bedrock:*::foundation-model/${local.foundation_model_id}"
        ]
      }
    ]
  })
}

# Data sources
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}