# Mail Service

## Overview
This Mail Service is a lightweight, efficient solution for handling contact form submissions on websites. It provides a seamless email forwarding and auto-response system, with support for both REST API and gRPC interfaces.

## Features
- Processes contact form submissions
- Sends an automatic thank you email to the user using a customizable template
- Forwards the submitted information to a designated personal email address using a customizable template
- Supports both REST API and gRPC interfaces
- Configurable through environment variables for easy deployment and management

## Architecture
```
[Website Contact Form] -> [Mail Service (REST/gRPC)] -> [Email Sending]
                                     |
                                     v
                    [Thank You Email] [Forward to Personal Email]
```

## Prerequisites
- Email sending capability (e.g., AWS SES, SMTP server)
- Docker installed (for local development and building)
- Kubernetes cluster (optional, for deployment)

## Setup and Deployment

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/mail-service.git
cd mail-service
```

### 2. Create Your Templates

Each template must be HTML. Template variables follow the AWS SES variable notation. Examples can be found in the `./emailtemplates`s folder.

#### Forward Template
The template used to forward the senders email (form user) to your personal email.

##### Variables
* {{name}} => The name of the person who submitted the form.
* {{subject}} => The subject of the form submission.
* {{email}} => The email address of the person who submitted the form.
* {{message}} => The message of the person who submitted the form.

#### Thank You Template
The template used to send a thank you email to the senders email address. 

##### Variables
* {{name}} => The name of the person who submitted the form.

### 3. Configure Environment Variables
Set the following environment variables:

```
EMAIL_SERVICE_PORT=8080
EMAIL_SERVICE_LISTEN_ADDRESS=0.0.0.0
EMAIL_SERVICE_GRPC_HOST=127.0.0.1
EMAIL_SERVICE_GRPC_PORT=8081
EMAIL_SERVICE_EMAIL_FROM=noreply@yourdomain.com
EMAIL_SERVICE_EMAIL_FORWARD=your-personal-email@example.com
EMAIL_SERVICE_EMAIL_THANK_YOU_TEMPLATE=base64_encoded_html_template
EMAIL_SERVICE_EMAIL_FORWARD_TEMPLATE=base64_encoded_html_template
```

Note: The thank you and forward templates should be base64 standard encoded HTML templates.

### 4. Build the Docker Image
```bash
docker build -t mail-service:latest .
```

### 5. Run the Service
```bash
docker run -p 8080:8080 -p 8081:8081 --env-file .env mail-service:latest
```

## Usage

### REST API Endpoint
POST `/v1/mail/send`

Request Body:
```json
{
    "name": "John Doe",
    "email": "john@example.com",
    "subject": "Hey There",
    "message": "Hello, I'd like to get in touch!"
}
```

### gRPC Service
The service also exposes a gRPC interface on the configured host and port (default: 127.0.0.1:8081).

## Configuration Details

The service can be configured using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| EMAIL_SERVICE_PORT | The port for the REST API | 8080 |
| EMAIL_SERVICE_LISTEN_ADDRESS | The address to listen on for the REST API | 0.0.0.0 |
| EMAIL_SERVICE_GRPC_HOST | The host for the gRPC service | 127.0.0.1 |
| EMAIL_SERVICE_GRPC_PORT | The port for the gRPC service | 8081 |
| EMAIL_SERVICE_EMAIL_FROM | The email address to send from | (required) |
| EMAIL_SERVICE_EMAIL_FORWARD | The email address to forward submissions to | (required) |
| EMAIL_SERVICE_EMAIL_THANK_YOU_TEMPLATE | Base64 encoded HTML template for thank you emails | (required) |
| EMAIL_SERVICE_EMAIL_FORWARD_TEMPLATE | Base64 encoded HTML template for forwarded emails | (required) |

## Monitoring and Logs
You can monitor the service using Docker or Kubernetes tools:

```bash
docker logs mail-service
# or
kubectl logs deployment/mail-service
```

## Troubleshooting
- Ensure all required environment variables are set correctly
- Check logs for any configuration or runtime errors

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.