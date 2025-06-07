variable "region" {
  default = "us-east-1"
}

variable "notification_email" {
  description = "Email address to receive billing notifications"
  type        = string
  default     = "sarangkharche@gmail.com"
}
