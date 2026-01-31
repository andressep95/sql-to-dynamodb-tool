# ============================================
# DLQ Handler Lambda
# ============================================

module "dlq_handler" {
  source        = "../../../modules/lambda"
  function_name = "dlq_handler"
  filename      = abspath("${path.root}/../../../../lambda/dlq-handler/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/dlq-handler/function.zip"))
  role_arn         = var.role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  # Resources
  memory_size                    = local.lambda_configs["dlq_handler"].memory_size
  timeout                        = local.lambda_configs["dlq_handler"].timeout
  reserved_concurrent_executions = local.lambda_configs["dlq_handler"].reserved_concurrent_executions

  # Environment variables
  environment_variables = {
    DYNAMODB_TABLE_NAME = var.dynamodb_table_name
    DYNAMODB_ENDPOINT   = var.dynamodb_endpoint
  }

  # Logging
  log_retention_days = local.lambda_configs["dlq_handler"].log_retention_days
  create_log_group   = false # Disabled for LocalStack

  # Tags
  environment = var.environment
  tags = merge(
    var.common_tags,
    {
      Component    = local.component_name
      UseCase      = local.lambda_configs["dlq_handler"].use_case
      ApiOperation = local.lambda_configs["dlq_handler"].api_operation
    }
  )
}

# ============================================
# SQS Event Source Mapping -> dlq_handler (from DLQ)
# ============================================

resource "aws_lambda_event_source_mapping" "dlq_to_dlq_handler" {
  event_source_arn = var.sqs_dlq_arn
  function_name    = module.dlq_handler.function_arn
  batch_size       = 1
  enabled          = true
}
