module "frontend_bucket" {
  source = "../../../modules/s3"

  bucket_name        = "${var.environment}-sql-to-nosql-frontend"
  force_destroy      = true
  versioning_enabled = false
  enable_encryption  = true
  static_files_path  = "${path.module}/../../../../../web/db-parser/dist"
  tags               = var.common_tags

  # Access logging to dedicated logs bucket
  enable_logging = true
  logging_bucket = aws_s3_bucket.logs.id
}
