package mail

import "github.com/aws/aws-sdk-go-v2/service/sesv2/types"

type emailTemplate struct {
	Name    string
	Content *types.EmailTemplateContent
}

var (
	// templates is a slice of emailTemplate structs that holds the predefined email templates
	// used by the mail service. Each template includes a name and the corresponding content.
	//
	// The templates include:
	//   - ThankYouTemplate: A template used to send a thank you email to the original sender.
	//   - ForwardTemplate: A template used to forward the email to a predefined address.
	templates = []emailTemplate{
		{
			Name:    thankYouTemplateName,
			Content: thankYouTemplateContent,
		},
		{
			Name:    forwardName,
			Content: forwardContent,
		},
	}

	thankYouTemplateName    = "ThankYouTemplate"
	thankYouTemplateSubject = "Thank you for your interest"
	thankYouTemplateContent = &types.EmailTemplateContent{
		Subject: &thankYouTemplateSubject,
		Html:    &thankYouTemplateHTML,
	}
	thankYouTemplateHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Thank You for Reaching Out!</title>
    <link href="https://fonts.googleapis.com/css?family=Roboto:400,700&display=swap" rel="stylesheet">
    <script src="https://kit.fontawesome.com/73784867c6.js" crossorigin="anonymous"></script>
    <style>
        body {
            font-family: 'Roboto', Arial, sans-serif;
            background-color: #f9f9f9;
            margin: 0;
            padding: 0;
        }
        .email-container {
            background-color: #ffffff;
            width: 90%;
            max-width: 600px;
            margin: 40px auto;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
            color: #333333;
        }
        .header {
            text-align: center;
            margin-bottom: 20px;
        }
        .avatar {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            border: 4px solid #4CAF50;
            object-fit: cover;
            box-shadow: 0 4px 8px rgba(76, 175, 80, 0.2);
        }
        .greeting {
            font-size: 24px;
            margin: 20px 0 10px 0;
        }
        .message {
            font-size: 16px;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .footer-icon {
            text-align: center;
            color: #4CAF50;
            font-size: 36px;
            margin-bottom: 20px;
        }
        .footer {
            text-align: center;
            font-size: 14px;
            color: #777777;
        }
        /* Responsive Design */
        @media (max-width: 600px) {
            .email-container {
                padding: 20px;
            }
            .avatar {
                width: 100px;
                height: 100px;
            }
            .greeting {
                font-size: 20px;
            }
            .message {
                font-size: 14px;
            }
            .footer-icon {
                font-size: 28px;
            }
        }
    </style>
</head>
<body>
    <div class="email-container">
        <div class="header">
            <img src="https://healthauraassets.s3.us-east-2.amazonaws.com/81f8d481-f85e-47cb-8450-15820e642f5c.png" alt="Your Avatar" class="avatar">
        </div>
        <div class="content">
            <h1 class="greeting">Hi {{name}},</h1>
            <p class="message">
                Thank you for reaching out! I appreciate your interest in connecting with me. As a passionate software developer, I am excited to learn more about your projects and how we can collaborate.
            </p>
            <p class="message">
                I will get back to you as soon as possible to discuss your ideas and explore potential opportunities together.
            </p>
        </div>
        <div class="footer-icon">
            <i class="fas fa-envelope"></i>
        </div>
        <div class="footer">
            <p>If you need immediate assistance, feel free to contact me directly at <a href="mailto:me@bricealrich.com">me@bricealdrich.com</a> or (260) 582-9842.</p>
            <p>© 2024 Brice Aldrich. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	forwardName    = "ForwardTemplate"
	forwardSubject = "You have an inquiry"
	forwardContent = &types.EmailTemplateContent{
		Subject: &forwardSubject,
		Text:    &thankYouTemplateText,
	}
	thankYouTemplateText = "{{text}}"
)
