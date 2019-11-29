data "aws_region" "current" {}

resource "aws_cloudfront_distribution" "verynice" {
  enabled         = true
  is_ipv6_enabled = true
  price_class     = "PriceClass_100"
  aliases = [
    "samsverynice.website",
  ]

  origin {
    origin_id   = "apigatewaysamsverynice"
    domain_name = "${aws_api_gateway_rest_api.samsverynice.id}.execute-api.${data.aws_region.current.name}.amazonaws.com"
    origin_path = "/${aws_api_gateway_deployment.prod.stage_name}"

    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_keepalive_timeout = 5
      origin_protocol_policy   = "https-only"
      origin_read_timeout      = 30
      origin_ssl_protocols = [
        "TLSv1",
        "TLSv1.1",
        "TLSv1.2",
      ]
    }
  }

  default_cache_behavior {
    target_origin_id       = "apigatewaysamsverynice"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    viewer_protocol_policy = "redirect-to-https"
    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate.samsverynice_website.arn
    minimum_protocol_version = "TLSv1.1_2016"
    ssl_support_method       = "sni-only"
  }
}
