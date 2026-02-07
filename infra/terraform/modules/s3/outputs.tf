output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.this.bucket
}

output "bucket_id" {
  description = "ID of the S3 bucket"
  value       = aws_s3_bucket.this.id
}

output "bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.this.arn
}

output "bucket_region" {
  description = "Region where the bucket is created"
  value       = aws_s3_bucket.this.region
}

output "bucket_regional_domain_name" {
  description = "Regional domain name of the S3 bucket"
  value       = aws_s3_bucket.this.bucket_regional_domain_name
}

output "static_files_etags" {
  description = "Map of static file keys to their etags, used to trigger CloudFront invalidation"
  value       = { for k, v in aws_s3_object.static_files : k => v.etag }
}

output "html_object_keys" {
  description = "Uploaded HTML object keys"
  value       = keys(aws_s3_object.html_objects)
}
