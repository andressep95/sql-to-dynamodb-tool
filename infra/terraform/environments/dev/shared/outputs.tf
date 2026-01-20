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
