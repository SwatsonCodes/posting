resource "aws_lambda_function" "very_nice" {
  function_name = "very_nice"
  filename      = "../nice_lambda.zip"
  runtime       = "go1.x"
  handler       = "main"
  role          = aws_iam_role.lambda_very_nice.arn
  memory_size   = 128
  timeout       = 2
  publish       = "true"
}

data "aws_iam_policy_document" "lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy" "lambda_basic_role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role" "lambda_very_nice" {
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role_policy.json
  name               = "lambda_very_nice"
}

resource "aws_iam_role_policy_attachment" "lambda_very_nice" {
  policy_arn = data.aws_iam_policy.lambda_basic_role.arn
  role       = aws_iam_role.lambda_very_nice.name
}

resource "aws_lambda_permission" "api_gateway_invoke_very_nice_lambda" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.very_nice.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.samsverynice.execution_arn}/*/${aws_api_gateway_method.base_get.http_method}/"
}
