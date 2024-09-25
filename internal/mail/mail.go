package mail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	mailservice_v1 "github.com/brice-aldrich/mail-service/gen/go/mailservice.v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// sesClient is an interface that defines the methods from the AWS SES client that are used by the Orchestrator.
type sesClient interface {
	GetEmailTemplate(ctx context.Context, params *sesv2.GetEmailTemplateInput, optFns ...func(*sesv2.Options)) (*sesv2.GetEmailTemplateOutput, error)
	CreateEmailTemplate(ctx context.Context, params *sesv2.CreateEmailTemplateInput, optFns ...func(*sesv2.Options)) (*sesv2.CreateEmailTemplateOutput, error)
	UpdateEmailTemplate(ctx context.Context, params *sesv2.UpdateEmailTemplateInput, optFns ...func(*sesv2.Options)) (*sesv2.UpdateEmailTemplateOutput, error)
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

// Orchestrator defines the interface for sending emails.
// It includes a single method, SendMail, which is responsible for sending an email based on the provided request.
//
// Methods:
//   - SendMail: Sends an email based on the provided request. It forwards the email to a predefined address and sends a thank you email to the original sender.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - req: The SendMailRequest object containing the email message and recipient information.
//
// Returns:
//   - *mailservice_v1.SendMailResponse: The response object indicating the result of the send mail operation.
//   - error: An error if any occurred during the preparation of template data or sending of emails.
type Orchestrator interface {
	SendMail(ctx context.Context, req *mailservice_v1.SendMailRequest) (*mailservice_v1.SendMailResponse, error)
}

// Config holds the configuration required to initialize the Orchestrator.
// It includes the SES client for sending emails, the forward email address, and the from email address.
//
// Fields:
//   - SES: The sesv2.Client object used to interact with AWS SES for sending emails.
//   - ForwardEmail: The email address to which incoming emails will be forwarded.
//   - FromEmail: The email address from which emails will be sent.
//   - Logger: The zap.Logger object used for logging.
type Config struct {
	SES                     sesClient
	ForwardEmail            string
	FromEmail               string
	ForwardTemplateEncoded  string
	ThankYouTemplateEncoded string
	Logger                  *zap.Logger
}

type orchestrator struct {
	ses              sesClient
	forwardEmail     string
	fromEmail        string
	thankYouTemplate emailTemplate
	forwardTemplate  emailTemplate
	logger           *zap.Logger
}

// emailTemplate - a wrapper for CreateEmailTemplateInput and UpdateEmailTemplateInput from AWS SES sdk
type emailTemplate struct {
	Name    string
	Content *types.EmailTemplateContent
}

// New creates a new instance of the Orchestrator with the provided configuration.
// It initializes the orchestrator with the SES client, forward email address, and from email address from the configuration.
// It also initializes or updates the email templates in AWS SES.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - cfg: The Config object containing the SES client, forward email address, and from email address.
//
// Returns:
//   - Orchestrator: The newly created Orchestrator instance.
//   - error: An error if any occurred during the initialization of the email templates.
func New(ctx context.Context, cfg Config) (Orchestrator, error) {
	o := &orchestrator{
		ses:          cfg.SES,
		forwardEmail: cfg.ForwardEmail,
		fromEmail:    cfg.FromEmail,
		logger:       cfg.Logger,
	}

	var err error
	o.forwardTemplate, err = constructForwardTemplate(cfg.ForwardTemplateEncoded)
	if err != nil {
		return nil, err
	}

	o.thankYouTemplate, err = constructThankYouTemplate(cfg.ThankYouTemplateEncoded)
	if err != nil {
		return nil, err
	}

	if err := o.initTemplate(ctx, o.forwardTemplate); err != nil {
		return nil, fmt.Errorf("failed to initialize forward template: %w", err)
	}

	if err := o.initTemplate(ctx, o.thankYouTemplate); err != nil {
		return nil, fmt.Errorf("failed to initialize thank you template: %w", err)
	}

	return o, nil
}

// initTemplate initializes or updates a single email template in AWS SES based on the provided template.
// It performs the following actions:
// 1. Checks if the template already exists in AWS SES.
// 2. If the template does not exist, it creates the template in AWS SES.
// 3. If the template exists, it updates the template in AWS SES.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - t: The emailTemplate object containing the template name and content.
//
// Returns:
//   - error: An error if any occurred during the initialization or updating of the email template.
func (o orchestrator) initTemplate(ctx context.Context, t emailTemplate) error {
	_, err := o.ses.GetEmailTemplate(ctx, &sesv2.GetEmailTemplateInput{
		TemplateName: &t.Name,
	})
	if err != nil {
		var notFoundErr *types.NotFoundException
		if errors.As(err, &notFoundErr) {
			_, err := o.ses.CreateEmailTemplate(ctx, &sesv2.CreateEmailTemplateInput{
				TemplateName:    &t.Name,
				TemplateContent: t.Content,
			})
			if err != nil {
				return fmt.Errorf("failed to create email template with aws ses: %w", err)
			}

			return nil
		}

		return fmt.Errorf("failed to initialize email template with aws ses: %w", err)
	}

	_, err = o.ses.UpdateEmailTemplate(ctx, &sesv2.UpdateEmailTemplateInput{
		TemplateName:    &t.Name,
		TemplateContent: t.Content,
	})
	if err != nil {
		return fmt.Errorf("failed to update email template with aws ses: %w", err)
	}

	return nil
}

// SendMail sends an email based on the provided request. It performs two main actions:
// 1. Forwards the email to a predefined address using a forward template.
// 2. Sends a thank you email to the original sender using a thank you template.
//
// It first constructs the forward template data and sends the forward email.
// Then, constructs the thank you template data and sends the thank you email.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - req: The SendMailRequest object containing the email message and recipient information.
//
// Returns:
//   - *mailservice_v1.SendMailResponse: The response object indicating the result of the send mail operation.
//   - error: An error if any occurred during the preparation of template data or sending of emails.
func (o orchestrator) SendMail(ctx context.Context, req *mailservice_v1.SendMailRequest) (*mailservice_v1.SendMailResponse, error) {
	forwardData, err := constructForwardTemplateData(req.Message, req.Name, req.Email, req.Subject)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to prepare forward template data: %v", err)
	}

	_, err = o.ses.SendEmail(ctx, &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateName: &o.forwardTemplate.Name,
				TemplateData: forwardData,
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{o.forwardEmail},
		},
		FromEmailAddress: &o.fromEmail,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send forward email: %v", err)
	}

	o.logger.Info("Forward email sent", zap.String("to", o.forwardEmail))

	thankYouData, err := constructThankYouTemplateData(req.Message)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to prepare thank you template data: %v", err)
	}

	_, err = o.ses.SendEmail(ctx, &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Template: &types.Template{
				TemplateName: &o.thankYouTemplate.Name,
				TemplateData: thankYouData,
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{req.Email},
		},
		FromEmailAddress: &o.fromEmail,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send email thank you email: %v", err)
	}

	return &mailservice_v1.SendMailResponse{}, nil
}

func constructForwardTemplate(encodedTemplate string) (emailTemplate, error) {
	v, err := base64.StdEncoding.DecodeString(encodedTemplate)
	if err != nil {
		return emailTemplate{}, fmt.Errorf("failed to decode forward template: %w", err)
	}

	subject := "Portfolio Contact Form Submission"
	template := string(v)
	return emailTemplate{
		Name: "ForwardTemplate",
		Content: &types.EmailTemplateContent{
			Subject: &subject,
			Html:    &template,
		},
	}, nil
}

func constructThankYouTemplate(encodedTemplate string) (emailTemplate, error) {
	v, err := base64.StdEncoding.DecodeString(encodedTemplate)
	if err != nil {
		return emailTemplate{}, fmt.Errorf("failed to decode thank you template: %w", err)
	}

	subject := "Thank you for your interest"
	template := string(v)
	return emailTemplate{
		Name: "ThankYouTemplate",
		Content: &types.EmailTemplateContent{
			Subject: &subject,
			Html:    &template,
		},
	}, nil
}

func constructForwardTemplateData(message string, name string, email string, subject *string) (*string, error) {
	defaultSubject := "Portfolio Contact Form Inquiry"

	if subject == nil {
		subject = &defaultSubject
	}

	templateData := map[string]string{
		"message": message,
		"name":    name,
		"email":   email,
		"subject": *subject,
	}

	v, err := json.Marshal(&templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	templateDataString := string(v)
	return &templateDataString, nil
}

func constructThankYouTemplateData(name string) (*string, error) {
	templateData := map[string]string{
		"name": name,
	}

	v, err := json.Marshal(&templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	templateDataString := string(v)
	return &templateDataString, nil
}
