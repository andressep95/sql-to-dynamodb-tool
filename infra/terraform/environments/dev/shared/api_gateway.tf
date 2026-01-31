# ============================================
# Shared Module - API Gateway (REST v1)
# ============================================

module "api_gateway" {
  source = "../../../modules/gateway/wrapper"

  gateway_type = "rest-v1"

  name       = var.api_gateway_name
  region     = var.aws_region
  stage_name = var.stage_name

  # Lambda integration
  lambda_name = var.lambda_function_name
  lambda_arn  = var.lambda_function_arn

  # Routes configuration
  routes = {
    "POST /api/v1/schemas"     = {}
    "GET /api/v1/schemas"      = { lambda_arn = var.query_handler_function_arn, lambda_name = var.query_handler_function_name }
    "GET /api/v1/schemas/{id}" = { lambda_arn = var.query_handler_function_arn, lambda_name = var.query_handler_function_name }
  }

  tags = var.common_tags
}
