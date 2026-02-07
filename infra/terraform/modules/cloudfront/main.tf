locals {
  s3_origin_id          = "${var.project_name}-s3-origin"
  api_gateway_origin_id = "${var.project_name}-api-origin"
  api_gateway_domain    = replace(replace(var.api_gateway_endpoint, "https://", ""), "/", "")
  use_custom_domain     = var.custom_domain != ""
  use_secret_header     = var.cloudflare_secret_header_value != ""
  enable_logging        = var.logging_bucket != ""
}

# ============================================
# ACM Certificate (conditional on custom_domain)
# ============================================

resource "aws_acm_certificate" "this" {
  count             = local.use_custom_domain ? 1 : 0
  domain_name       = var.custom_domain
  validation_method = "DNS"

  tags = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_validation" "this" {
  count           = local.use_custom_domain ? 1 : 0
  certificate_arn = aws_acm_certificate.this[0].arn
}

# ============================================
# CloudFront Function — validate X-Cf-Secret
# ============================================

resource "aws_cloudfront_function" "verify_cf_secret" {
  count   = local.use_secret_header ? 1 : 0
  name    = "${var.environment}-${var.project_name}-verify-cf-secret"
  runtime = "cloudfront-js-2.0"
  comment = "Validates X-Cf-Secret header from Cloudflare"
  publish = true

  code = <<-EOF
    function handler(event) {
      var request = event.request;
      
      // Allow OPTIONS requests for CORS preflight
      if (request.method === 'OPTIONS') {
        return request;
      }
      
      var secret = request.headers['x-origin-secret'];
      var expected = '${var.cloudflare_secret_header_value}';
      if (!secret || secret.value !== expected) {
        return {
          statusCode: 403,
          statusDescription: 'Forbidden',
          body: { encoding: 'text', data: 'Forbidden' }
        };
      }
      return request;
    }
  EOF
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
  aliases             = local.use_custom_domain ? [var.custom_domain] : []

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

  # Default cache behavior: S3 origin (using cache policy instead of deprecated forwarded_values)
  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = local.s3_origin_id
    viewer_protocol_policy = "redirect-to-https"

    cache_policy_id = data.aws_cloudfront_cache_policy.caching_optimized.id

    compress = true

    dynamic "function_association" {
      for_each = local.use_secret_header ? [1] : []
      content {
        event_type   = "viewer-request"
        function_arn = aws_cloudfront_function.verify_cf_secret[0].arn
      }
    }
  }

  # Ordered cache behavior: Static assets with long TTL (1 year)
  ordered_cache_behavior {
    path_pattern           = "/assets/*"
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = local.s3_origin_id
    viewer_protocol_policy = "redirect-to-https"

    cache_policy_id = data.aws_cloudfront_cache_policy.caching_optimized.id

    compress = true

    dynamic "function_association" {
      for_each = local.use_secret_header ? [1] : []
      content {
        event_type   = "viewer-request"
        function_arn = aws_cloudfront_function.verify_cf_secret[0].arn
      }
    }
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

    dynamic "function_association" {
      for_each = local.use_secret_header ? [1] : []
      content {
        event_type   = "viewer-request"
        function_arn = aws_cloudfront_function.verify_cf_secret[0].arn
      }
    }
  }

  # Access logging (conditional)
  dynamic "logging_config" {
    for_each = local.enable_logging ? [1] : []
    content {
      include_cookies = false
      bucket          = var.logging_bucket
      prefix          = "cloudfront/${var.environment}/"
    }
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

  dynamic "viewer_certificate" {
    for_each = local.use_custom_domain ? [1] : []
    content {
      acm_certificate_arn      = aws_acm_certificate_validation.this[0].certificate_arn
      ssl_support_method       = "sni-only"
      minimum_protocol_version = "TLSv1.2_2021"
    }
  }

  dynamic "viewer_certificate" {
    for_each = local.use_custom_domain ? [] : [1]
    content {
      cloudfront_default_certificate = true
    }
  }

  tags = var.tags
}

# ============================================
# Managed cache/origin request policies
# ============================================

data "aws_cloudfront_cache_policy" "caching_disabled" {
  name = "Managed-CachingDisabled"
}

data "aws_cloudfront_cache_policy" "caching_optimized" {
  name = "Managed-CachingOptimized"
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
