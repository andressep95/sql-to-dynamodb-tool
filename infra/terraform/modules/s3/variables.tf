variable "bucket_name" {
  description = "Name of the S3 bucket"
  type        = string
}

variable "force_destroy" {
  description = "Allow bucket to be destroyed even if it contains objects"
  type        = bool
  default     = false
}

variable "versioning_enabled" {
  description = "Enable versioning for the bucket"
  type        = bool
  default     = true
}

variable "enable_encryption" {
  description = "Enable server-side encryption (SSE-S3)"
  type        = bool
  default     = true
}

variable "html_objects" {
  description = <<EOF
Map of HTML objects to upload to S3.
Key   = S3 object key (e.g. index.html)
Value = local file path
EOF
  type        = map(string)
  default     = {}
}

variable "static_files_path" {
  description = "Path to a directory of static files to upload to the bucket. All files are uploaded with auto-detected content types."
  type        = string
  default     = ""
}


variable "tags" {
  description = "Tags applied to the bucket"
  type        = map(string)
  default     = {}
}

variable "enable_logging" {
  description = "Enable access logging for this bucket"
  type        = bool
  default     = false
}

variable "logging_bucket" {
  description = "Target bucket ID for access logs (required if enable_logging is true)"
  type        = string
  default     = ""
}
