resource "aws_route53_zone" "samsverynice" {
  name          = "samsverynice.website"
  force_destroy = false
  comment       = "HostedZone created by Route53 Registrar"
}

resource "aws_route53_record" "samsverynice_alias_cloudfront" {
  zone_id = aws_route53_zone.samsverynice.zone_id
  name    = "samsverynice.website."
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.verynice.domain_name
    zone_id                = aws_cloudfront_distribution.verynice.hosted_zone_id
    evaluate_target_health = false
  }
}
