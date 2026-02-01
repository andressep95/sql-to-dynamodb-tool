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

# Conversion Worker outputs
output "worker_function_name" {
  description = "Name of the conversion worker Lambda"
  value       = module.lambda_components.worker_function_name
}

output "worker_function_arn" {
  description = "ARN of the conversion worker Lambda"
  value       = module.lambda_components.worker_function_arn
}
