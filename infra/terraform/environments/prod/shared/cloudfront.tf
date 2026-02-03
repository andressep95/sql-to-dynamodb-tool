module "cloudfront" {
  source = "../../../modules/cloudfront"

  s3_bucket_arn                  = module.frontend_bucket.bucket_arn
  s3_bucket_id                   = module.frontend_bucket.bucket_id
  s3_bucket_regional_domain_name = module.frontend_bucket.bucket_regional_domain_name

  api_gateway_endpoint = module.api_gateway.api_endpoint

  static_files_etags = module.frontend_bucket.static_files_etags

  environment  = var.environment
  project_name = "sql-to-nosql"
  tags         = var.common_tags
}
