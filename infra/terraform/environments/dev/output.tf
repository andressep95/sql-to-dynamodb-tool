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
