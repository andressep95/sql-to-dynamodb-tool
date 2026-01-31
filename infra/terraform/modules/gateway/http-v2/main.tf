locals {
  # Build a map of unique lambda invoke ARNs to create one integration per lambda
  route_lambdas = {
    for key, route in var.routes : key => {
      invoke_arn = coalesce(lookup(route, "lambda_invoke_arn", null), var.lambda_invoke_arn)
      name       = coalesce(lookup(route, "lambda_name", null), var.lambda_name)
    }
  }

  # Deduplicate integrations by lambda name
  unique_lambdas = {
    for key, cfg in local.route_lambdas : cfg.name => cfg...
  }
  # Flatten: take first entry from each group
  unique_lambdas_flat = {
    for name, cfgs in local.unique_lambdas : name => cfgs[0]
  }
}

resource "aws_apigatewayv2_api" "this" {
  name          = var.name
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = var.cors_allow_origins
    allow_methods = var.cors_allow_methods
    allow_headers = var.cors_allow_headers
  }

  tags = var.tags
}

# One integration per unique Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  for_each = local.unique_lambdas_flat

  api_id = aws_apigatewayv2_api.this.id

  integration_type       = "AWS_PROXY"
  integration_uri        = each.value.invoke_arn
  payload_format_version = "2.0"
}

# Routes pointing to the correct integration
resource "aws_apigatewayv2_route" "this" {
  for_each = var.routes

  api_id    = aws_apigatewayv2_api.this.id
  route_key = each.key
  target    = "integrations/${aws_apigatewayv2_integration.lambda[local.route_lambdas[each.key].name].id}"
}

# Stage
resource "aws_apigatewayv2_stage" "this" {
  api_id = aws_apigatewayv2_api.this.id
  name   = var.stage_name

  auto_deploy = true

  tags = var.tags
}

# One permission per unique Lambda
resource "aws_lambda_permission" "apigw" {
  for_each = local.unique_lambdas_flat

  statement_id  = "AllowHttpApiInvoke-${each.key}"
  action        = "lambda:InvokeFunction"
  function_name = each.value.name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.this.execution_arn}/*/*"
}
