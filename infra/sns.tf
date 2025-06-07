resource "aws_sns_topic" "billing_notifications" {
  name = "xapien-billing-notifications"

  tags = {
    Name = "xapien-billing-notifications"
    Purpose = "Customer credit limit notifications"
  }
}

resource "aws_sns_topic_subscription" "email_notification" {
  topic_arn = aws_sns_topic.billing_notifications.arn
  protocol  = "email"
  endpoint  = var.notification_email
}

output "sns_topic_arn" {
  value = aws_sns_topic.billing_notifications.arn
  description = "ARN of the SNS topic for billing notifications"
}