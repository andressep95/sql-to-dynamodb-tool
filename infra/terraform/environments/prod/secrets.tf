# ============================================
# Secrets Manager - Resend API Key
# ============================================

module "resend_secret" {
  source = "../../modules/secrets-manager"

  environment             = var.environment
  secret_name             = "resend-api-key"
  description             = "Resend API key for sending invitation emails"
  secret_value            = var.resend_api_key
  recovery_window_in_days = 7

  tags = local.common_tags
}
