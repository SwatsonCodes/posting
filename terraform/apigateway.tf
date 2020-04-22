resource "aws_api_gateway_rest_api" "samsverynice" {
  name        = "samsverynice"
  description = "its a nice website"
  endpoint_configuration {
    types = [
      "REGIONAL",
    ]
  }
}

resource "aws_api_gateway_method" "base_get" {
  rest_api_id   = aws_api_gateway_rest_api.samsverynice.id
  resource_id   = aws_api_gateway_rest_api.samsverynice.root_resource_id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "base_get_very_nice_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.samsverynice.id
  resource_id             = aws_api_gateway_rest_api.samsverynice.root_resource_id
  http_method             = aws_api_gateway_method.base_get.http_method
  type                    = "AWS_PROXY"
  integration_http_method = "POST"
  uri                     = aws_lambda_function.very_nice.invoke_arn
}

resource "aws_api_gateway_method_response" "base_get_200" {
  rest_api_id = aws_api_gateway_rest_api.samsverynice.id
  resource_id = aws_api_gateway_rest_api.samsverynice.root_resource_id
  http_method = aws_api_gateway_method.base_get.http_method
  status_code = "200"
  response_models = {
    "text/plain" = "Empty"
  }
}

resource "aws_api_gateway_resource" "proxy" {
  rest_api_id = aws_api_gateway_rest_api.samsverynice.id
  parent_id   = aws_api_gateway_rest_api.samsverynice.root_resource_id
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "proxy_any" {
  rest_api_id   = aws_api_gateway_rest_api.samsverynice.id
  resource_id   = aws_api_gateway_resource.proxy.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "proxy_any_very_nice_lambda" {
  rest_api_id             = aws_api_gateway_rest_api.samsverynice.id
  resource_id             = aws_api_gateway_resource.proxy.id
  http_method             = aws_api_gateway_method.proxy_any.http_method
  type                    = "AWS_PROXY"
  integration_http_method = "POST"
  uri                     = aws_lambda_function.very_nice.invoke_arn
}

resource "aws_api_gateway_method_response" "proxy_any_200" {
  rest_api_id = aws_api_gateway_rest_api.samsverynice.id
  resource_id = aws_api_gateway_resource.proxy.id
  http_method = aws_api_gateway_method.proxy_any.http_method
  status_code = "200"
  response_models = {
    "text/plain" = "Empty"
  }
}

resource "aws_api_gateway_deployment" "production" {
  depends_on  = [aws_api_gateway_integration.base_get_very_nice_lambda]
  rest_api_id = aws_api_gateway_rest_api.samsverynice.id
  stage_name  = "production"
}

resource "aws_api_gateway_domain_name" "samsverynice_website" {
  domain_name     = "samsverynice.website"
  certificate_arn = aws_acm_certificate.samsverynice_website.arn
}

resource "aws_api_gateway_base_path_mapping" "prod" {
  api_id      = aws_api_gateway_rest_api.samsverynice.id
  stage_name  = aws_api_gateway_deployment.production.stage_name
  domain_name = aws_api_gateway_domain_name.samsverynice_website.domain_name
}
