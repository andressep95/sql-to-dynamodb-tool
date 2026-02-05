# ============================================
# Shared Module - Outputs (Production)
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

# CloudFront outputs
output "cloudfront_distribution_domain" {
  description = "Domain name of the CloudFront distribution"
  value       = module.cloudfront.distribution_domain_name
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = module.cloudfront.distribution_id
}

output "frontend_bucket_name" {
  description = "Name of the frontend S3 bucket"
  value       = module.frontend_bucket.bucket_name
}

output "acm_certificate_validation_records" {
  description = "DNS records required to validate the ACM certificate"
  value       = module.cloudfront.acm_certificate_validation_records
}

# SNS outputs
output "sns_topic_arn" {
  description = "ARN of the SNS topic for operational alarms"
  value       = aws_sns_topic.alarms.arn
}

# Logs bucket outputs
output "logs_bucket_name" {
  description = "Name of the logs S3 bucket"
  value       = aws_s3_bucket.logs.bucket
}

output "logs_bucket_arn" {
  description = "ARN of the logs S3 bucket"
  value       = aws_s3_bucket.logs.arn
}

output "logs_bucket_domain_name" {
  description = "Domain name of the logs S3 bucket"
  value       = aws_s3_bucket.logs.bucket_domain_name
}
