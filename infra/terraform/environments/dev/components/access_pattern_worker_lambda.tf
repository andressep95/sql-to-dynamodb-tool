# ============================================
# Access Pattern Worker Lambda
# ============================================

module "access_pattern_worker" {
  source        = "../../../modules/lambda"
  function_name = "access_pattern_worker"
  filename      = abspath("${path.root}/../../../../lambda/access-pattern-worker/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/access-pattern-worker/function.zip"))
  role_arn         = var.role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  # Resources
  memory_size                    = local.lambda_configs["access_pattern_worker"].memory_size
  timeout                        = local.lambda_configs["access_pattern_worker"].timeout
  reserved_concurrent_executions = local.lambda_configs["access_pattern_worker"].reserved_concurrent_executions

  # Environment variables
  environment_variables = {
    DYNAMODB_TABLE_NAME           = var.dynamodb_table_name
    DYNAMODB_ENDPOINT             = var.dynamodb_endpoint
    BEDROCK_MODEL_ID              = "us.anthropic.claude-sonnet-4-20250514-v1:0"
    USE_MOCK_BEDROCK              = tostring(var.use_mock_bedrock)
    BEDROCK_ENDPOINT              = "https://bedrock-runtime.${var.aws_region}.amazonaws.com"
    BEDROCK_AWS_ACCESS_KEY_ID     = var.aws_access_key_id
    BEDROCK_AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
    BEDROCK_AWS_REGION            = var.aws_region
  }

  # Logging
  log_retention_days = local.lambda_configs["access_pattern_worker"].log_retention_days
  create_log_group   = false # Disabled for LocalStack

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
