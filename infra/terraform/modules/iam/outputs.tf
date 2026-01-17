# ============================================
# IAM Module - Outputs
# ============================================

output "role_arn" {
  description = "ARN of the Lambda execution role"
  value       = aws_iam_role.lambda_execution.arn
}

output "role_name" {
  description = "Name of the Lambda execution role"
  value       = aws_iam_role.lambda_execution.name
}

output "role_id" {
  description = "ID of the Lambda execution role"
  value       = aws_iam_role.lambda_execution.id
}

output "custom_policy_arns" {
  description = "Map of custom policy ARNs created"
  value = {
    for k, v in aws_iam_policy.custom_policies : k => v.arn
  }
}

output "custom_policy_names" {
  description = "Map of custom policy names created"
  value = {
    for k, v in aws_iam_policy.custom_policies : k => v.name
  }
}

variable "role_purpose" {
  description = "Purpose of the role (read, write, admin, etc)"
  type        = string
}
