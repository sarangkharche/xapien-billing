provider "aws" {
  region = "us-east-1"
}

resource "aws_dynamodb_table" "organisation_usage" {
  name         = "organisation_usage"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "org_id"

  attribute {
    name = "org_id"
    type = "S"
  }

  tags = {
    Name = "organisation-usage"
  }
}
