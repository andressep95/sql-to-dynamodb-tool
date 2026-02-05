# ============================================
# S3 Bucket for Access Logs (CloudFront, S3, etc.)
# ============================================

resource "aws_s3_bucket" "logs" {
  bucket        = "${var.environment}-sql-to-nosql-logs"
  force_destroy = false

  tags = var.common_tags
}

# Block public access
resource "aws_s3_bucket_public_access_block" "logs" {
  bucket = aws_s3_bucket.logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Enable SSE-S3 encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Lifecycle policy - delete logs after 90 days
resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    id     = "expire-logs-after-90-days"
    status = "Enabled"

    expiration {
      days = 90
    }

    filter {
      prefix = ""
    }
  }
}

# Bucket ownership controls for CloudFront logging
resource "aws_s3_bucket_ownership_controls" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

# ACL for CloudFront logging (requires ownership controls)
resource "aws_s3_bucket_acl" "logs" {
  depends_on = [aws_s3_bucket_ownership_controls.logs]

  bucket = aws_s3_bucket.logs.id
  acl    = "log-delivery-write"
}
