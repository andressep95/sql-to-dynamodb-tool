# ============================================
# Access Pattern Worker Lambda (Production)
# ============================================

module "access_pattern_worker" {
  source        = "../../../modules/lambda"
  function_name = "${var.environment}-access_pattern_worker"
  filename      = abspath("${path.root}/../../../../lambda/access-pattern-worker/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/access-pattern-worker/function.zip"))
  role_arn         = var.access_pattern_worker_role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  # Resources
  memory_size                    = local.lambda_configs["access_pattern_worker"].memory_size
  timeout                        = local.lambda_configs["access_pattern_worker"].timeout
  reserved_concurrent_executions = local.lambda_configs["access_pattern_worker"].reserved_concurrent_executions

  # Environment variables - no mock bedrock, no explicit credentials (uses IAM role)
  environment_variables = {
    DYNAMODB_TABLE_NAME = var.dynamodb_table_name
    BEDROCK_MODEL_ID    = "us.anthropic.claude-3-5-haiku-20241022-v1:0"
    BEDROCK_AWS_REGION  = "us-east-1"
  }

  # Logging
  log_retention_days = local.lambda_configs["access_pattern_worker"].log_retention_days
  create_log_group   = true

  # Tags
  environment = var.environment
  tags = merge(
    var.common_tags,
    {
      Component    = local.component_name
      UseCase      = local.lambda_configs["access_pattern_worker"].use_case
      ApiOperation = local.lambda_configs["access_pattern_worker"].api_operation
    }
  )
}

# ============================================
# SQS Event Source Mapping -> access_pattern_worker
# ============================================

resource "aws_lambda_event_source_mapping" "sqs_to_access_pattern_worker" {
  event_source_arn = var.access_pattern_queue_arn
  function_name    = module.access_pattern_worker.function_arn
  batch_size       = 1
  enabled          = true
}
