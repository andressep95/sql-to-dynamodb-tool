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

  schema {
    name                = "role"
    attribute_data_type = "String"
    mutable             = true

    string_attribute_constraints {
      min_length = 1
      max_length = 50
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
    email_subject        = "Bienvenido a SQL to DynamoDB Converter"
    email_message        = <<-EOT
      <html>
        <body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
          <div style="background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
            <h1 style="color: white; margin: 0;">SQL to DynamoDB Converter</h1>
          </div>
          <div style="background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px;">
            <h2 style="color: #111827; margin-top: 0;">¡Bienvenido!</h2>
            <p style="color: #374151; font-size: 16px; line-height: 1.6;">
              Gracias por unirte a nuestra plataforma. Para completar tu registro, por favor usa el siguiente código de verificación:
            </p>
            <div style="background: white; border: 2px solid #3b82f6; border-radius: 8px; padding: 20px; text-align: center; margin: 20px 0;">
              <span style="font-size: 32px; font-weight: bold; color: #3b82f6; letter-spacing: 5px;">{####}</span>
            </div>
            <p style="color: #6b7280; font-size: 14px; margin-top: 20px;">
              Este código expirará en 24 horas. Si no solicitaste este registro, puedes ignorar este mensaje.
            </p>
          </div>
          <div style="text-align: center; padding: 20px; color: #9ca3af; font-size: 12px;">
            <p>© 2026 SQL to DynamoDB Converter. Todos los derechos reservados.</p>
          </div>
        </body>
      </html>
    EOT
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
