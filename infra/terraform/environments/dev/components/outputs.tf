# ============================================
# Components Module - Outputs
# ============================================

# Healthcheck Lambda outputs
output "function_name" {
  description = "Name of the Lambda function"
  value       = module.healthcheck.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = module.healthcheck.function_arn
}

output "invoke_arn" {
  description = "Invoke ARN of the Lambda function"
  value       = module.healthcheck.invoke_arn
}

output "runtime" {
  description = "Runtime of the Lambda function"
  value       = module.healthcheck.runtime
}

output "architecture" {
  description = "Architecture of the Lambda function"
  value       = module.healthcheck.architecture
}

output "memory_size" {
  description = "Memory size of the Lambda function"
  value       = module.healthcheck.memory_size
}

output "environment_variables" {
  description = "Environment variables of the Lambda function"
  value       = module.healthcheck.environment_variables
}
