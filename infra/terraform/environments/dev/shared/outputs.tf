# ============================================
# Shared Module - Outputs
# ============================================

output "api_gateway_id" {
  description = "ID of the API Gateway"
  value       = module.api_gateway.api_id
}

output "api_gateway_endpoint" {
  description = "Invoke URL of the API Gateway"
  value       = module.api_gateway.api_endpoint
}

# DynamoDB outputs
output "schemas_table_name" {
  description = "Name of the schemas DynamoDB table"
  value       = module.schemas_table.table_name
}

output "schemas_table_arn" {
  description = "ARN of the schemas DynamoDB table"
  value       = module.schemas_table.table_arn
}
