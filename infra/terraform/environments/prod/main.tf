# ============================================
# Main - Production Environment
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

  backend "s3" {
    bucket       = "sql-to-nosql-terraform-state"
    key          = "prod/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile

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

  # Per-Lambda role names
  process_handler_role_name   = "${local.project_name}-${local.environment}-process-handler-role"
  conversion_worker_role_name = "${local.project_name}-${local.environment}-conversion-worker-role"
  query_handler_role_name     = "${local.project_name}-${local.environment}-query-handler-role"
  dlq_handler_role_name       = "${local.project_name}-${local.environment}-dlq-handler-role"
}

# ============================================
# IAM Roles for Lambda (one per function)
# ============================================

locals {
  lambda_assume_role_policy = jsonencode({
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
}

resource "aws_iam_role" "process_handler" {
  name               = local.process_handler_role_name
  assume_role_policy = local.lambda_assume_role_policy
  tags               = local.common_tags
}

resource "aws_iam_role" "conversion_worker" {
  name               = local.conversion_worker_role_name
  assume_role_policy = local.lambda_assume_role_policy
  tags               = local.common_tags
}

resource "aws_iam_role" "query_handler" {
  name               = local.query_handler_role_name
  assume_role_policy = local.lambda_assume_role_policy
  tags               = local.common_tags
}

resource "aws_iam_role" "dlq_handler" {
  name               = local.dlq_handler_role_name
  assume_role_policy = local.lambda_assume_role_policy
  tags               = local.common_tags
}

# Lambda basic execution (CloudWatch Logs) — one per role
resource "aws_iam_role_policy_attachment" "process_handler_basic_execution" {
  role       = aws_iam_role.process_handler.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "conversion_worker_basic_execution" {
  role       = aws_iam_role.conversion_worker.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "query_handler_basic_execution" {
  role       = aws_iam_role.query_handler.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "dlq_handler_basic_execution" {
  role       = aws_iam_role.dlq_handler.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# ============================================
# SQS (Conversion Queue)
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

  process_handler_role_arn   = aws_iam_role.process_handler.arn
  conversion_worker_role_arn = aws_iam_role.conversion_worker.arn
  query_handler_role_arn     = aws_iam_role.query_handler.arn
  dlq_handler_role_arn       = aws_iam_role.dlq_handler.arn

  dynamodb_table_name = local.schemas_table_name
  sqs_queue_url       = module.conversion_queue.queue_url
  sqs_queue_arn       = module.conversion_queue.queue_arn
  sqs_dlq_arn         = module.conversion_queue.dlq_arn
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
  lambda_invoke_arn    = module.lambda_components.invoke_arn

  # Query handler Lambda
  query_handler_function_name = module.lambda_components.query_function_name
  query_handler_function_arn  = module.lambda_components.query_function_arn
  query_handler_invoke_arn    = module.lambda_components.query_invoke_arn

  # API Gateway config
  api_gateway_name = "${local.project_name}-api"
  stage_name       = var.environment

  # Lambda IAM role names for policy attachments (per-Lambda)
  process_handler_role_name   = aws_iam_role.process_handler.name
  conversion_worker_role_name = aws_iam_role.conversion_worker.name
  query_handler_role_name     = aws_iam_role.query_handler.name
  dlq_handler_role_name       = aws_iam_role.dlq_handler.name

  # DynamoDB
  schemas_table_name = local.schemas_table_name

  # SQS ARN for IAM policies
  sqs_queue_arn = module.conversion_queue.queue_arn
  sqs_dlq_arn   = module.conversion_queue.dlq_arn
}

# ============================================
# Bedrock - Claude 3.5 Sonnet v2 for SQL to DynamoDB conversion (Production)
# ============================================

module "bedrock" {
  source = "../../modules/bedrock"

  model_name     = "claude-3-5-sonnet-v2"
  model_id       = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"
  model_provider = "anthropic"

  enable_logging        = false
  skip_model_validation = false
  create_lambda_policy  = true
  lambda_role_name      = aws_iam_role.conversion_worker.name

  tags = local.common_tags

  depends_on = [aws_iam_role.conversion_worker]
}
