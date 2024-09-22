# Mail Service

## Overview
This Mail Service is a lightweight, efficient solution for handling contact form submissions on websites. Powered by AWS Simple Email Service (SES) and deployed on an Amazon Elastic Kubernetes Service (EKS) cluster, it provides a seamless email forwarding and auto-response system.

## Features
- Processes contact form submissions
- Sends an automatic thank you email to the user
- Forwards the submitted information to a designated personal email address
- Utilizes AWS SES for reliable email delivery
- Deployed on EKS for scalability and easy management

## Architecture
```
[Website Contact Form] -> [EKS Cluster] -> [Mail Service] -> [AWS SES]
                                                   |
                                                   v
                               [Thank You Email]   [Forward to Personal Email]
```

## Prerequisites
- AWS Account with SES configured
- EKS cluster set up and running
- `kubectl` configured to interact with your EKS cluster
- Docker installed (for local development and building)

## Setup and Deployment

### 1. Clone the Repository
```bash
git clone https://github.com/brice-aldrich/mail-service.git
cd mail-service
```

### 2. Configure AWS SES
Ensure that your AWS SES is set up and verified for both sending and receiving emails.

### 3. Set Environment Variables
Create a `.env` file with the following variables:
```
AWS_REGION=your-aws-region
FROM_EMAIL=noreply@yourdomain.com
FORWARD_EMAIL=your-personal-email@example.com
```

### 4. Build the Docker Image
```bash
docker build -t mail-service:latest .
```

### 5. Push to Your Container Registry
```bash
docker tag mail-service:latest your-registry/mail-service:latest
docker push your-registry/mail-service:latest
```

### 6. Deploy to EKS
```bash
kubectl apply -f kubernetes/deployment.yaml
kubectl apply -f kubernetes/service.yaml
```

## Usage
Once deployed, the mail service will listen for incoming requests from your website's contact form. It processes these requests as follows:

1. Receives form submission data
2. Sends a thank you email to the submitter
3. Forwards the submission details to your personal email

## API Endpoint
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

## Monitoring and Logs
You can monitor the service using Kubernetes tools:

```bash
kubectl logs deployment/mail-service
```

## Troubleshooting
- Ensure AWS credentials are correctly set up in your EKS cluster
- Verify that SES is properly configured and out of sandbox mode if necessary
- Check EKS cluster logs for any deployment issues

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.