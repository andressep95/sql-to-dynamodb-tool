# Admin Handler Lambda
# Handles user and tenant management

module "admin_handler" {
  source        = "../../../modules/lambda"
  function_name = "${var.environment}-admin_handler"
  filename      = abspath("${path.root}/../../../../lambda/admin-handler/function.zip")

  source_code_hash = filebase64sha256(abspath("${path.root}/../../../../lambda/admin-handler/function.zip"))
  role_arn         = aws_iam_role.admin_handler.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architecture     = "arm64"

  memory_size = 256
  timeout     = 30

  environment_variables = {
    COGNITO_USER_POOL_ID  = var.cognito_user_pool_id
    INVITATIONS_TABLE     = "${var.environment}-invitations"
    RESEND_SECRET_NAME    = "${var.environment}-resend-api-key"
    APP_URL               = "https://${var.custom_domain}"
  }

  log_retention_days = 30
  create_log_group   = true

  environment = var.environment
  tags = merge(var.common_tags, {
    Component = "admin-handler"
  })
}

# IAM Role for Admin Handler
resource "aws_iam_role" "admin_handler" {
  name = "${var.environment}-admin-handler-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = var.common_tags
}

resource "aws_iam_role_policy_attachment" "admin_handler_basic_execution" {
  role       = aws_iam_role.admin_handler.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Cognito permissions
resource "aws_iam_role_policy" "admin_handler_cognito" {
  name = "${var.environment}-admin-handler-cognito-policy"
  role = aws_iam_role.admin_handler.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "cognito-idp:ListUsers",
          "cognito-idp:AdminCreateUser",
          "cognito-idp:AdminSetUserPassword",
          "cognito-idp:AdminUpdateUserAttributes",
          "cognito-idp:AdminDeleteUser",
          "cognito-idp:AdminGetUser"
        ]
        Resource = var.cognito_user_pool_arn
      }
    ]
  })
}

# DynamoDB permissions for invitations table
resource "aws_iam_role_policy" "admin_handler_invitations" {
  name = "${var.environment}-admin-handler-invitations-policy"
  role = aws_iam_role.admin_handler.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem"
        ]
        Resource = "arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/${var.environment}-invitations"
      }
    ]
  })
}

# Secrets Manager permissions for Resend API key
resource "aws_iam_role_policy" "admin_handler_secrets" {
  name = "${var.environment}-admin-handler-secrets-policy"
  role = aws_iam_role.admin_handler.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = var.resend_secret_arn
      }
    ]
  })
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
