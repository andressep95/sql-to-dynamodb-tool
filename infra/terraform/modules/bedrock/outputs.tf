output "model_arn" {
  description = "ARN of the Bedrock foundation model"
  value       = "arn:aws:bedrock:${data.aws_region.current.name}::foundation-model/${var.model_id}"
}

output "available_models" {
  description = "List of available foundation models from the provider"
  value       = var.skip_model_validation ? [] : data.aws_bedrock_foundation_models.available[0].model_summaries
}

output "log_group_name" {
  description = "CloudWatch log group name for Bedrock (if logging enabled)"
  value       = var.enable_logging ? aws_cloudwatch_log_group.bedrock[0].name : null
}

output "region" {
  description = "AWS region where Bedrock is deployed"
  value       = data.aws_region.current.name
}