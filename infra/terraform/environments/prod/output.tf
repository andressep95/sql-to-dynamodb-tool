output "function_name" {
  description = "Name of the Lambda function"
  value       = module.lambda_components.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = module.lambda_components.function_arn
}

output "invoke_arn" {
  description = "Invoke ARN of the Lambda function"
  value       = module.lambda_components.invoke_arn
}

output "runtime" {
  description = "Runtime of the Lambda function"
  value       = module.lambda_components.runtime
}

output "architecture" {
  description = "Architecture of the Lambda function"
  value       = module.lambda_components.architecture
}

output "memory_size" {
  description = "Memory size of the Lambda function"
  value       = module.lambda_components.memory_size
}

output "environment_variables" {
  description = "Environment variables of the Lambda function"
  value       = module.lambda_components.environment_variables
}

# API Gateway outputs
output "api_gateway_id" {
  description = "ID of the API Gateway"
  value       = module.shared.api_gateway_id
}

output "api_gateway_endpoint" {
  description = "Invoke URL of the API Gateway"
  value       = module.shared.api_gateway_endpoint
}

# DynamoDB outputs
output "schemas_table_name" {
  description = "Name of the schemas DynamoDB table"
  value       = module.shared.schemas_table_name
}

output "schemas_table_arn" {
  description = "ARN of the schemas DynamoDB table"
  value       = module.shared.schemas_table_arn
}

# SQS outputs
output "sqs_queue_url" {
  description = "URL of the conversion SQS queue"
  value       = module.conversion_queue.queue_url
}

output "sqs_queue_arn" {
  description = "ARN of the conversion SQS queue"
  value       = module.conversion_queue.queue_arn
}

# CloudFront outputs
output "cloudfront_distribution_domain" {
  description = "Domain name of the CloudFront distribution"
  value       = module.shared.cloudfront_distribution_domain
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = module.shared.cloudfront_distribution_id
}

output "frontend_bucket_name" {
  description = "Name of the frontend S3 bucket"
  value       = module.shared.frontend_bucket_name
}

# Conversion Worker outputs
output "worker_function_name" {
  description = "Name of the conversion worker Lambda"
  value       = module.lambda_components.worker_function_name
}

output "worker_function_arn" {
  description = "ARN of the conversion worker Lambda"
  value       = module.lambda_components.worker_function_arn
}

# ACM outputs
output "acm_certificate_validation_records" {
  description = "DNS records to add in Cloudflare for ACM certificate validation"
  value       = module.shared.acm_certificate_validation_records
}

# Cognito outputs
output "cognito_user_pool_id" {
  description = "ID of the Cognito User Pool"
  value       = module.cognito.user_pool_id
}

output "cognito_client_id" {
  description = "ID of the Cognito User Pool Client"
  value       = module.cognito.client_id
}

# Frontend configuration (copy to web/db-parser/.env.local)
output "frontend_env_config" {
  description = "Environment variables for frontend configuration"
  value = <<-EOT
    
    ╔════════════════════════════════════════════════════════════════╗
    ║  Frontend Configuration (.env.local)                           ║
    ╚════════════════════════════════════════════════════════════════╝
    
    Copy these values to: web/db-parser/.env.local
    
    VITE_BASE_PATH_URL=https://${module.shared.cloudfront_distribution_domain}
    VITE_ENDPOINT_URL=prod/api/v1/schemas
    VITE_COGNITO_USER_POOL_ID=${module.cognito.user_pool_id}
    VITE_COGNITO_CLIENT_ID=${module.cognito.client_id}
    
  EOT
}
