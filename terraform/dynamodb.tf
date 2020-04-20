resource "aws_dynamodb_table" "posts" {
  name         = "posts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "post_id"

  attribute {
    name = "post_id"
    type = "S"
  }
}

data "aws_iam_policy_document" "dynamodb_posts_read_write" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:BatchGetItem",
      "dynamodb:BatchWriteItem",
      "dynamodb:ConditionCheckItem",
      "dynamodb:PutItem",
      "dynamodb:GetItem",
      "dynamodb:Scan",
      "dynamodb:Query",
      "dynamodb:UpdateItem",
    ]
    resources = [aws_dynamodb_table.posts.arn]
  }
}

resource "aws_iam_policy" "dynamodb_posts_read_write" {
  name   = "dynamodb_posts_read_write"
  policy = data.aws_iam_policy_document.dynamodb_posts_read_write.json
}
