resource "aws_route53_zone" "samsverynice" {
  name          = "samsverynice.website"
  force_destroy = false
  comment       = "HostedZone created by Route53 Registrar"
}

resource "aws_route53_record" "samsverynice_alias_apig" {
  zone_id = aws_route53_zone.samsverynice.zone_id
  name    = "samsverynice.website."
  type    = "A"

  alias {
    name                   = aws_api_gateway_domain_name.samsverynice_website.cloudfront_domain_name
    zone_id                = aws_api_gateway_domain_name.samsverynice_website.cloudfront_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "samsverynice_cert_validation" {
  name    = aws_acm_certificate.samsverynice_website.domain_validation_options.0.resource_record_name
  type    = aws_acm_certificate.samsverynice_website.domain_validation_options.0.resource_record_type
  zone_id = aws_route53_zone.samsverynice.id
  records = [aws_acm_certificate.samsverynice_website.domain_validation_options.0.resource_record_value]
  ttl     = 300
}
