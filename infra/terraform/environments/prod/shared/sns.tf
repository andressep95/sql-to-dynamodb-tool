# ============================================
# SNS Topic for Operational Alarms
# ============================================

resource "aws_sns_topic" "alarms" {
  name = "${var.environment}-operational-alarms"

  tags = var.common_tags
}

resource "aws_sns_topic_subscription" "email_alerts" {
  topic_arn = aws_sns_topic.alarms.arn
  protocol  = "email"
  endpoint  = "andressep.95@gmail.com"
}
