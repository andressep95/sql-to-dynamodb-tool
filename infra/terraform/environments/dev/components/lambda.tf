module "healthcheck" {
  source        = "../../../modules/lambda"
  function_name = "healthcheck"
  filename      = abspath("${path.root}/../../../../lambda/diagrams/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/diagrams/function.zip"))
  role_arn         = var.role_arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  # Resources
  memory_size                    = local.lambda_configs["healthcheck"].memory_size
  timeout                        = local.lambda_configs["healthcheck"].timeout
  reserved_concurrent_executions = local.lambda_configs["healthcheck"].reserved_concurrent_executions

  # Logging
  log_retention_days = local.lambda_configs["healthcheck"].log_retention_days
  create_log_group   = false # Disabled for LocalStack

  # Tags
  environment = var.environment
  tags = merge(
    var.common_tags,
    {
      Component    = local.component_name
      UseCase      = local.lambda_configs["healthcheck"].use_case
      ApiOperation = local.lambda_configs["healthcheck"].api_operation
    }
  )
}
