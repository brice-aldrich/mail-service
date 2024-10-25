output "api_endpoint" {
  value = "${aws_api_gateway_stage.dev.invoke_url}/email"
}

output "lambda_function_name" {
  value = aws_lambda_function.email_sender.function_name
  description = "The name of the Lambda function for CI/CD deployment"
}

output "lambda_function_arn" {
  value = aws_lambda_function.email_sender.arn
  description = "The ARN of the Lambda function for CI/CD deployment"
}