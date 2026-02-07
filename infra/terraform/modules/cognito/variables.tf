variable "user_pool_name" {
  description = "Name of the Cognito User Pool"
  type        = string
}

variable "client_name" {
  description = "Name of the Cognito User Pool Client"
  type        = string
}

variable "password_minimum_length" {
  description = "Minimum password length"
  type        = number
  default     = 8
}

variable "mfa_configuration" {
  description = "MFA configuration: OFF, ON, or OPTIONAL"
  type        = string
  default     = "OFF"

  validation {
    condition     = contains(["OFF", "ON", "OPTIONAL"], var.mfa_configuration)
    error_message = "mfa_configuration must be OFF, ON, or OPTIONAL"
  }
}

variable "access_token_validity_hours" {
  description = "Access token validity in hours"
  type        = number
  default     = 1
}

variable "refresh_token_validity_days" {
  description = "Refresh token validity in days"
  type        = number
  default     = 30
}

variable "callback_urls" {
  description = "Callback URLs for OAuth flows"
  type        = list(string)
  default     = []
}

variable "logout_urls" {
  description = "Logout URLs for OAuth flows"
  type        = list(string)
  default     = []
}

variable "post_confirmation_lambda_arn" {
  description = "ARN of the Lambda function to invoke after user confirmation"
  type        = string
  default     = null
}

variable "deletion_protection" {
  description = "Enable deletion protection for the User Pool"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
