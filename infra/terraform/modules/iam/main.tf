# ============================================
# IAM Module - Generic Lambda Execution Role
# ============================================

resource "aws_iam_role" "lambda_execution" {
  name               = "${var.domain_name}-${var.role_purpose}-lambda-${var.environment}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  tags               = var.tags
}


# Trust Policy
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

# ============================================
# Managed Policy Attachments
# ============================================

resource "aws_iam_role_policy_attachment" "managed_policies" {
  for_each = toset(var.managed_policy_arns)

  role       = aws_iam_role.lambda_execution.name
  policy_arn = each.value
}

# ============================================
# Custom Inline Policies
# ============================================

resource "aws_iam_role_policy" "inline_policies" {
  for_each = var.inline_policies

  role   = aws_iam_role.lambda_execution.name
  name   = each.key
  policy = each.value
}

# ============================================
# Custom Managed Policies
# ============================================

resource "aws_iam_policy" "custom_policies" {
  for_each = var.custom_policies

  name        = "${var.domain_name}-${each.key}-${var.environment}"
  description = each.value.description
  policy      = each.value.policy_document

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "custom_policies" {
  for_each = aws_iam_policy.custom_policies

  role       = aws_iam_role.lambda_execution.name
  policy_arn = each.value.arn
}
