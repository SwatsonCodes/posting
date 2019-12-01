resource "aws_acm_certificate" "samsverynice_website" {
  provider          = aws.us-east-1
  domain_name       = "samsverynice.website"
  validation_method = "DNS"
}
