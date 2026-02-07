variable "environment" {
  description = "Environment name (dev, prod)"
  type        = string
}

variable "secret_name" {
  description = "Name of the secret (will be prefixed with environment)"
  type        = string
}

variable "description" {
  description = "Description of the secret"
  type        = string
  default     = ""
}

variable "secret_value" {
  description = "The secret value to store"
  type        = string
  sensitive   = true
}

variable "recovery_window_in_days" {
  description = "Number of days to retain secret after deletion"
  type        = number
  default     = 7
}

variable "tags" {
  description = "Tags to apply to the secret"
  type        = map(string)
  default     = {}
}
