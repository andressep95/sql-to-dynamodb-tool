locals {
  s3_origin_id          = "${var.project_name}-s3-origin"
  api_gateway_origin_id = "${var.project_name}-api-origin"
  api_gateway_domain    = replace(replace(var.api_gateway_endpoint, "https://", ""), "/", "")
}

# ============================================
# Origin Access Control (OAC) for S3
# ============================================

resource "aws_cloudfront_origin_access_control" "this" {
  name                              = "${var.environment}-${var.project_name}-oac"
  description                       = "OAC for ${var.project_name} S3 frontend bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# ============================================
# CloudFront Distribution
# ============================================

resource "aws_cloudfront_distribution" "this" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = var.default_root_object
  price_class         = var.price_class
  comment             = "${var.environment} ${var.project_name} distribution"

  # Origin 1: S3 (default)
  origin {
    domain_name              = var.s3_bucket_regional_domain_name
    origin_id                = local.s3_origin_id
    origin_access_control_id = aws_cloudfront_origin_access_control.this.id
  }

  # Origin 2: API Gateway
  origin {
    domain_name = local.api_gateway_domain
    origin_id   = local.api_gateway_origin_id

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  # Default cache behavior: S3 origin
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = local.s3_origin_id
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    min_ttl     = 0
    default_ttl = 3600
    max_ttl     = 86400
    compress    = true
  }

  # Ordered cache behavior: API Gateway (no caching)
  ordered_cache_behavior {
    path_pattern           = "/prod/api/*"
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = local.api_gateway_origin_id
    viewer_protocol_policy = "redirect-to-https"

    cache_policy_id          = data.aws_cloudfront_cache_policy.caching_disabled.id
    origin_request_policy_id = data.aws_cloudfront_origin_request_policy.all_viewer_except_host_header.id

    compress = true
  }

  # SPA routing: redirect 403/404 to index.html
  custom_error_response {
    error_code            = 403
    response_code         = 200
    response_page_path    = "/index.html"
    error_caching_min_ttl = 10
  }

  custom_error_response {
    error_code            = 404
    response_code         = 200
    response_page_path    = "/index.html"
    error_caching_min_ttl = 10
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = var.tags
}

# ============================================
# Managed cache/origin request policies
# ============================================

data "aws_cloudfront_cache_policy" "caching_disabled" {
  name = "Managed-CachingDisabled"
}

data "aws_cloudfront_origin_request_policy" "all_viewer_except_host_header" {
  name = "Managed-AllViewerExceptHostHeader"
}

# ============================================
# S3 Bucket Policy — allow CloudFront OAC
# ============================================

# ============================================
# Invalidar cache cuando cambian archivos estáticos
# ============================================

resource "terraform_data" "cloudfront_invalidation" {
  count = length(var.static_files_etags) > 0 ? 1 : 0

  triggers_replace = var.static_files_etags

  provisioner "local-exec" {
    command = "aws cloudfront create-invalidation --distribution-id ${aws_cloudfront_distribution.this.id} --paths '/*'"
  }
}

# ============================================
# S3 Bucket Policy — allow CloudFront OAC
# ============================================

resource "aws_s3_bucket_policy" "frontend" {
  bucket = var.s3_bucket_id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFrontServicePrincipal"
        Effect    = "Allow"
        Principal = { Service = "cloudfront.amazonaws.com" }
        Action    = "s3:GetObject"
        Resource  = "${var.s3_bucket_arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.this.arn
          }
        }
      }
    ]
  })
}
