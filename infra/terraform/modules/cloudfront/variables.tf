variable "s3_bucket_arn" {
  description = "ARN of the S3 bucket for frontend assets"
  type        = string
}

variable "s3_bucket_id" {
  description = "ID of the S3 bucket for frontend assets"
  type        = string
}

variable "s3_bucket_regional_domain_name" {
  description = "Regional domain name of the S3 bucket"
  type        = string
}

variable "api_gateway_endpoint" {
  description = "Full URL of the API Gateway endpoint"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "tags" {
  description = "Tags applied to all resources"
  type        = map(string)
  default     = {}
}

variable "default_root_object" {
  description = "Default root object for the CloudFront distribution"
  type        = string
  default     = "index.html"
}

variable "price_class" {
  description = "CloudFront distribution price class"
  type        = string
  default     = "PriceClass_100"
}

variable "static_files_etags" {
  description = "Map of static file etags from S3 module, used to trigger cache invalidation on changes"
  type        = map(string)
  default     = {}
}
