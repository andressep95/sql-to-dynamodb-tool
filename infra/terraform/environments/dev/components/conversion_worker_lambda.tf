# ============================================
# Conversion Worker Lambda
# ============================================

module "conversion_worker" {
  source        = "../../../modules/lambda"
  function_name = "conversion_worker"
  filename      = abspath("${path.root}/../../../../lambda/conversion-worker/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/conversion-worker/function.zip"))
  role_arn         = var.role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  # Resources
  memory_size                    = local.lambda_configs["conversion_worker"].memory_size
  timeout                        = local.lambda_configs["conversion_worker"].timeout
  reserved_concurrent_executions = local.lambda_configs["conversion_worker"].reserved_concurrent_executions

  # Environment variables
  environment_variables = {
    DYNAMODB_TABLE_NAME = var.dynamodb_table_name
    DYNAMODB_ENDPOINT   = var.dynamodb_endpoint
    SQS_QUEUE_URL       = var.sqs_queue_url
    SQS_ENDPOINT        = var.sqs_endpoint
  }

  # Logging
  log_retention_days = local.lambda_configs["conversion_worker"].log_retention_days
  create_log_group   = false # Disabled for LocalStack

  # Tags
  environment = var.environment
  tags = merge(
    var.common_tags,
    {
      Component    = local.component_name
      UseCase      = local.lambda_configs["conversion_worker"].use_case
      ApiOperation = local.lambda_configs["conversion_worker"].api_operation
    }
  )
}

# ============================================
# SQS Event Source Mapping -> conversion_worker
# ============================================

resource "aws_lambda_event_source_mapping" "sqs_to_conversion_worker" {
  event_source_arn = var.sqs_queue_arn
  function_name    = module.conversion_worker.function_arn
  batch_size       = 1
  enabled          = true
}
