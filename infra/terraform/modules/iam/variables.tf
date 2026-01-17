# ============================================
# IAM Module - Variables (Generic)
# ============================================

variable "domain_name" {
  description = "Domain name (e.g., s3, bedrock)"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.domain_name))
    error_message = "Domain name must be lowercase alphanumeric with hyphens"
  }
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "managed_policy_arns" {
  description = "List of AWS managed policy ARNs to attach to the role"
  type        = list(string)
  default     = []
}

variable "inline_policies" {
  description = "Map of inline policies (policy name → policy document JSON string)"
  type        = map(string)
  default     = {}
}

variable "custom_policies" {
  description = "Map of custom managed policies to create and attach"
  type = map(object({
    description     = string
    policy_document = string
  }))
  default = {}
}

variable "tags" {
  description = "Additional tags for IAM resources"
  type        = map(string)
  default     = {}
}
