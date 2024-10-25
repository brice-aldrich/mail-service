# Lambda function role
resource "aws_iam_role" "lambda_role" {
  name = "email_sender_lambda_role_qa"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# SES permissions policy
resource "aws_iam_role_policy" "ses_policy" {
  name = "ses_permissions"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ses:CreateEmailTemplate",
          "ses:UpdateEmailTemplate",
          "ses:GetEmailTemplate",
          "ses:SendEmail",
          "ses:SendTemplatedEmail"
        ]
        Resource = "*"
      }
    ]
  })
}

# CloudWatch Logs policy
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Lambda function
resource "aws_lambda_function" "email_sender" {
  filename         = data.archive_file.empty_lambda.output_path  # Empty placeholder
  function_name    = "email_sender_qa"
  role            = aws_iam_role.lambda_role.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"
  publish         = true  # Enable versioning

  environment {
    variables = {
      ALLOWED_ORIGIN = "https://${data.terraform_remote_state.website.outputs.website_url}"  
    }
  }

  lifecycle {
    ignore_changes = [
      filename,
      source_code_hash,
    ]
  }
}

data "archive_file" "empty_lambda" {
  type        = "zip"
  output_path = "${path.module}/empty.zip"

  source {
    content  = "exports.handler = async () => { return { statusCode: 500, body: 'Lambda not deployed' } }"
    filename = "bootstrap"
  }
}


# API Gateway
resource "aws_api_gateway_rest_api" "email_api" {
  name = "email-sender-api"
}

# API Gateway resource
resource "aws_api_gateway_resource" "email" {
  rest_api_id = aws_api_gateway_rest_api.email_api.id
  parent_id   = aws_api_gateway_rest_api.email_api.root_resource_id
  path_part   = "email"
}

# API Gateway method
resource "aws_api_gateway_method" "email_post" {
  rest_api_id   = aws_api_gateway_rest_api.email_api.id
  resource_id   = aws_api_gateway_resource.email.id
  http_method   = "POST"
  authorization = "NONE"
}

# Usage plan for rate limiting
resource "aws_api_gateway_usage_plan" "email_usage_plan" {
  name = "email-usage-plan"

  api_stages {
    api_id = aws_api_gateway_rest_api.email_api.id
    stage  = aws_api_gateway_stage.qa.stage_name
  }

  quota_settings {
    limit  = 10    
    period = "DAY"  
  }

  throttle_settings {
    burst_limit = 2 
    rate_limit  = 0.05 
  }
}

# API Gateway integration with Lambda
resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id = aws_api_gateway_rest_api.email_api.id
  resource_id = aws_api_gateway_resource.email.id
  http_method = aws_api_gateway_method.email_post.http_method
  
  integration_http_method = "POST"
  type                   = "AWS_PROXY"
  uri                    = aws_lambda_function.email_sender.invoke_arn
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.email_sender.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.email_api.execution_arn}/*/${aws_api_gateway_method.email_post.http_method}${aws_api_gateway_resource.email.path}"
}

# CORS configuration
resource "aws_api_gateway_method" "email_options" {
  rest_api_id   = aws_api_gateway_rest_api.email_api.id
  resource_id   = aws_api_gateway_resource.email.id
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "options_integration" {
  rest_api_id = aws_api_gateway_rest_api.email_api.id
  resource_id = aws_api_gateway_resource.email.id
  http_method = aws_api_gateway_method.email_options.http_method
  type        = "MOCK"
  
  request_templates = {
    "application/json" = "{\"statusCode\": 200}"
  }
}

resource "aws_api_gateway_method_response" "options_200" {
  rest_api_id = aws_api_gateway_rest_api.email_api.id
  resource_id = aws_api_gateway_resource.email.id
  http_method = aws_api_gateway_method.email_options.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }
}

resource "aws_api_gateway_integration_response" "options_integration_response" {
  rest_api_id = aws_api_gateway_rest_api.email_api.id
  resource_id = aws_api_gateway_resource.email.id
  http_method = aws_api_gateway_method.email_options.http_method
  status_code = aws_api_gateway_method_response.options_200.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key'"
    "method.response.header.Access-Control-Allow-Methods" = "'POST,OPTIONS'"
    "method.response.header.Access-Control-Allow-Origin"  = "'https://${data.terraform_remote_state.website.outputs.website_url}'"
  }
}

# API Gateway deployment
resource "aws_api_gateway_deployment" "qa" {
  rest_api_id = aws_api_gateway_rest_api.email_api.id
  depends_on  = [
    aws_api_gateway_integration.lambda_integration,
    aws_api_gateway_integration_response.options_integration_response
  ]
}

# API Gateway stage
resource "aws_api_gateway_stage" "qa" {
  deployment_id = aws_api_gateway_deployment.qa.id
  rest_api_id  = aws_api_gateway_rest_api.email_api.id
  stage_name   = "qa"
}