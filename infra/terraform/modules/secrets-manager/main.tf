# ============================================
# Secrets Manager Module
# ============================================

resource "aws_secretsmanager_secret" "this" {
  name                    = "${var.environment}-${var.secret_name}"
  description             = var.description
  recovery_window_in_days = var.recovery_window_in_days

  tags = merge(var.tags, {
    Name        = "${var.environment}-${var.secret_name}"
    Environment = var.environment
  })
}

resource "aws_secretsmanager_secret_version" "this" {
  secret_id     = aws_secretsmanager_secret.this.id
  secret_string = var.secret_value
}
