# ============================================
# Shared Module - API Gateway (REST v1)
# ============================================

module "api_gateway" {
  source = "../../../modules/gateway/wrapper"

  gateway_type = "http-v2"

  name       = var.api_gateway_name
  region     = var.aws_region
  stage_name = var.stage_name

  # Default Lambda integration
  lambda_name       = var.lambda_function_name
  lambda_invoke_arn = var.lambda_invoke_arn

  # Routes configuration
  routes = {
    "POST /api/v1/schemas"     = { lambda_invoke_arn = var.lambda_invoke_arn, lambda_name = var.lambda_function_name }
    "GET /api/v1/schemas"      = { lambda_invoke_arn = var.query_handler_invoke_arn, lambda_name = var.query_handler_function_name }
    "GET /api/v1/schemas/{id}" = { lambda_invoke_arn = var.query_handler_invoke_arn, lambda_name = var.query_handler_function_name }
  }

  tags = var.common_tags
}
