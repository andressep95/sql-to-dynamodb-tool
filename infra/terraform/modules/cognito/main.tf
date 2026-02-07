# ============================================
# Cognito User Pool
# ============================================

resource "aws_cognito_user_pool" "this" {
  name                     = var.user_pool_name
  deletion_protection      = var.deletion_protection ? "ACTIVE" : "INACTIVE"
  auto_verified_attributes = ["email"]
  username_attributes      = ["email"]
  mfa_configuration        = var.mfa_configuration

  password_policy {
    minimum_length    = var.password_minimum_length
    require_lowercase = true
    require_numbers   = true
    require_symbols   = false
    require_uppercase = true
  }

  schema {
    name                = "tenant_id"
    attribute_data_type = "String"
    mutable             = true

    string_attribute_constraints {
      min_length = 1
      max_length = 256
    }
  }

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }

  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
    email_subject        = "Tu código de verificación"
    email_message        = "Tu código de verificación es {####}"
  }

  dynamic "lambda_config" {
    for_each = var.post_confirmation_lambda_arn != null ? [1] : []
    content {
      post_confirmation = var.post_confirmation_lambda_arn
    }
  }

  tags = var.tags
}

# ============================================
# Cognito User Pool Client (SPA - no secret)
# ============================================

resource "aws_cognito_user_pool_client" "this" {
  name         = var.client_name
  user_pool_id = aws_cognito_user_pool.this.id

  generate_secret = false

  explicit_auth_flows = [
    "ALLOW_USER_SRP_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]

  supported_identity_providers = ["COGNITO"]

  access_token_validity  = var.access_token_validity_hours
  id_token_validity      = var.access_token_validity_hours
  refresh_token_validity = var.refresh_token_validity_days

  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }

  callback_urls = var.callback_urls
  logout_urls   = var.logout_urls

  allowed_oauth_flows_user_pool_client = length(var.callback_urls) > 0
  allowed_oauth_flows                  = length(var.callback_urls) > 0 ? ["code"] : []
  allowed_oauth_scopes                 = length(var.callback_urls) > 0 ? ["openid", "email", "profile"] : []

  prevent_user_existence_errors = "ENABLED"
}
