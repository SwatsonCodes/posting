resource "aws_route53_zone" "samsverynice" {
  name          = "samsverynice.website"
  force_destroy = false
  comment       = "HostedZone created by Route53 Registrar"
}
