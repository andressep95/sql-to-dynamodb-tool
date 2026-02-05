resource "aws_s3_bucket" "this" {
  bucket        = var.bucket_name
  force_destroy = var.force_destroy

  tags = var.tags
}

# Bloqueo de acceso público (recomendado)
resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Versionado (opcional)
resource "aws_s3_bucket_versioning" "this" {
  count  = var.versioning_enabled ? 1 : 0
  bucket = aws_s3_bucket.this.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Encriptación en reposo (SSE-S3 - siempre habilitada)
resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Access logging (conditional)
resource "aws_s3_bucket_logging" "this" {
  count = var.enable_logging ? 1 : 0

  bucket        = aws_s3_bucket.this.id
  target_bucket = var.logging_bucket
  target_prefix = "s3/${var.bucket_name}/"
}

# Tipo de contenido correcto para HTML (opcional, recomendado)
resource "aws_s3_object" "html_objects" {
  for_each = var.html_objects

  bucket       = aws_s3_bucket.this.id
  key          = each.key
  source       = each.value
  content_type = "text/html"

  etag = filemd5(each.value)
}

# Archivos estáticos desde un directorio (frontend build)
locals {
  content_type_map = {
    ".html"  = "text/html"
    ".css"   = "text/css"
    ".js"    = "application/javascript"
    ".json"  = "application/json"
    ".png"   = "image/png"
    ".jpg"   = "image/jpeg"
    ".jpeg"  = "image/jpeg"
    ".gif"   = "image/gif"
    ".svg"   = "image/svg+xml"
    ".ico"   = "image/x-icon"
    ".woff"  = "font/woff"
    ".woff2" = "font/woff2"
    ".ttf"   = "font/ttf"
    ".eot"   = "application/vnd.ms-fontobject"
    ".map"   = "application/json"
    ".txt"   = "text/plain"
    ".xml"   = "application/xml"
    ".webp"  = "image/webp"
  }

  static_files = var.static_files_path != "" ? fileset(var.static_files_path, "**") : toset([])
}

resource "aws_s3_object" "static_files" {
  for_each = local.static_files

  bucket       = aws_s3_bucket.this.id
  key          = each.value
  source       = "${var.static_files_path}/${each.value}"
  content_type = lookup(local.content_type_map, regex("\\.[^.]+$", each.value), "application/octet-stream")

  etag = filemd5("${var.static_files_path}/${each.value}")
}
