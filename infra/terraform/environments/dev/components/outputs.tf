# ============================================
# Components Module - Outputs
# ============================================

# process_handler Lambda outputs
output "function_name" {
  description = "Name of the Lambda function"
  value       = module.process_handler.function_name
}

output "function_arn" {
  description = "ARN of the Lambda function"
  value       = module.process_handler.function_arn
}

output "invoke_arn" {
  description = "Invoke ARN of the Lambda function"
  value       = module.process_handler.invoke_arn
}

output "runtime" {
  description = "Runtime of the Lambda function"
  value       = module.process_handler.runtime
}

output "architecture" {
  description = "Architecture of the Lambda function"
  value       = module.process_handler.architecture
}

output "memory_size" {
  description = "Memory size of the Lambda function"
  value       = module.process_handler.memory_size
}

output "environment_variables" {
  description = "Environment variables of the Lambda function"
  value       = module.process_handler.environment_variables
}

# conversion_worker Lambda outputs
output "worker_function_name" {
  description = "Name of the conversion worker Lambda function"
  value       = module.conversion_worker.function_name
}

output "worker_function_arn" {
  description = "ARN of the conversion worker Lambda function"
  value       = module.conversion_worker.function_arn
}

