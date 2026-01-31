# ============================================
# Main - Development Environment
# Orchestrates: Shared Resources + Domains
# ============================================

terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  # LocalStack configuration
  access_key = var.use_localstack ? "test" : null
  secret_key = var.use_localstack ? "test" : null

  skip_credentials_validation = var.use_localstack
  skip_metadata_api_check     = var.use_localstack
  skip_requesting_account_id  = var.use_localstack

  # Override endpoints for LocalStack
  dynamic "endpoints" {
    for_each = var.use_localstack ? [1] : []
    content {
      apigateway     = var.localstack_endpoint
      apigatewayv2   = var.localstack_endpoint
      cloudformation = var.localstack_endpoint
      cloudwatch     = var.localstack_endpoint
      dynamodb       = var.localstack_endpoint
      ec2            = var.localstack_endpoint
      iam            = var.localstack_endpoint
      lambda         = var.localstack_endpoint
      route53        = var.localstack_endpoint
      s3             = var.localstack_endpoint
      secretsmanager = var.localstack_endpoint
      ses            = var.localstack_endpoint
      sns            = var.localstack_endpoint
      sqs            = var.localstack_endpoint
      ssm            = var.localstack_endpoint
      stepfunctions  = var.localstack_endpoint
      sts            = var.localstack_endpoint
    }
  }

  default_tags {
    tags = merge(
      var.common_tags,
      {
        Project     = var.project_name
        Environment = var.environment
        ManagedBy   = "Terraform"
        Repository  = "sql-to-dynamodb"
      }
    )
  }
}

# ============================================
# Data Sources
# ============================================

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# ============================================
# Local Variables
# ============================================

locals {
  environment  = var.environment
  project_name = var.project_name
  aws_region   = var.aws_region

  common_tags = {
    Project     = local.project_name
    Environment = local.environment
    ManagedBy   = "Terraform"
    Repository  = "sql-to-dynamodb"
  }
}

# ============================================
# Lambda Components
# ============================================

locals {
  schemas_table_name = "${var.environment}-${var.schemas_table_name}"
  lambda_role_name   = "${local.project_name}-${local.environment}-lambda-role"
}

# ============================================
# IAM Role for Lambda
# ============================================

resource "aws_iam_role" "lambda" {
  name = local.lambda_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

# ============================================
# SQS (Conversion Queue) - instantiated at root
# to avoid circular dependency between components and shared
# ============================================

module "conversion_queue" {
  source = "../../modules/sqs"

  queue_name = "${var.environment}-conversion-queue"
  dlq_name   = "${var.environment}-conversion-dlq"

  visibility_timeout_seconds = 180
  message_retention_seconds  = 345600 # 4 days
  receive_wait_time_seconds  = 20     # Long polling
  max_receive_count          = 3

  tags = local.common_tags
}

module "lambda_components" {
  source      = "./components"
  environment = var.environment
  common_tags = local.common_tags

  role_arn            = aws_iam_role.lambda.arn
  dynamodb_table_name = local.schemas_table_name
  dynamodb_endpoint   = var.use_localstack ? var.localstack_lambda_endpoint : ""
  sqs_queue_url       = module.conversion_queue.queue_url
  sqs_queue_arn       = module.conversion_queue.queue_arn
  sqs_endpoint        = var.use_localstack ? var.localstack_lambda_endpoint : ""
  sqs_dlq_arn         = module.conversion_queue.dlq_arn
  use_mock_bedrock    = var.use_localstack
}

# ============================================
# Shared Resources (API Gateway, DynamoDB, IAM)
# ============================================

module "shared" {
  source = "./shared"

  environment = var.environment
  aws_region  = var.aws_region
  common_tags = local.common_tags

  # Lambda integration from components
  lambda_function_name = module.lambda_components.function_name
  lambda_function_arn  = module.lambda_components.function_arn

  # Query handler Lambda
  query_handler_function_name = module.lambda_components.query_function_name
  query_handler_function_arn  = module.lambda_components.query_function_arn

  # API Gateway config
  api_gateway_name = "${local.project_name}-api"
  stage_name       = var.environment

  # Lambda IAM role for policy attachments
  lambda_role_name = aws_iam_role.lambda.name

  # DynamoDB
  schemas_table_name = local.schemas_table_name

  # SQS ARN for IAM policies
  sqs_queue_arn = module.conversion_queue.queue_arn
}
