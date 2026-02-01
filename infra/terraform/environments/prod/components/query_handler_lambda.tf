module "query_handler" {
  source        = "../../../modules/lambda"
  function_name = "${var.environment}-query_handler"
  filename      = abspath("${path.root}/../../../../lambda/query/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/query/function.zip"))
  role_arn         = var.query_handler_role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  # Resources
  memory_size                    = local.lambda_configs["query_handler"].memory_size
  timeout                        = local.lambda_configs["query_handler"].timeout
  reserved_concurrent_executions = local.lambda_configs["query_handler"].reserved_concurrent_executions

  # Environment variables
  environment_variables = {
    DYNAMODB_TABLE_NAME = var.dynamodb_table_name
    SQS_QUEUE_URL       = var.sqs_queue_url
  }

  # Logging
  log_retention_days = local.lambda_configs["query_handler"].log_retention_days
  create_log_group   = true

  # Tags
  environment = var.environment
  tags = merge(
    var.common_tags,
    {
      Component    = local.component_name
      UseCase      = local.lambda_configs["query_handler"].use_case
      ApiOperation = local.lambda_configs["query_handler"].api_operation
    }
  )
}
