package mail

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	mailservice_v1 "github.com/brice-aldrich/mail-service/gen/go/mailservice.v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInitTemplatesUnit(t *testing.T) {
	type input struct {
		ses sesClient
	}

	type want struct {
		errAssertion func(t *testing.T, err error)
	}

	cases := []struct {
		name  string
		input input
		want  want
	}{
		{
			"handles failure to get email template",
			input{
				ses: &mockSESClient{
					getEmailTemplateErr: "failed to get email template",
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					require.NotEmpty(t, err)
					assert.Contains(t, err.Error(), "failed to get email template")
				},
			},
		},
		{
			"handles failure to create email template if not found",
			input{
				ses: &mockSESClient{
					getEmailTemplateErr:    "NotFoundException",
					createEmailTemplateErr: "failed to create email template",
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					require.NotEmpty(t, err)
					assert.Contains(t, err.Error(), "failed to create email template")
				},
			},
		},
		{
			"handles failure to update email template",
			input{
				ses: &mockSESClient{
					getEmailTemplateErr:    "",
					updateEmailTemplateErr: "failed to update email template",
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					require.NotEmpty(t, err)
					assert.Contains(t, err.Error(), "failed to update email template")
				},
			},
		},
		{
			"is successful",
			input{
				ses: &mockSESClient{
					getEmailTemplateErr:    "",
					updateEmailTemplateErr: "",
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					assert.Empty(t, err)
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			o := orchestrator{
				ses: tt.input.ses,
			}

			err := o.initTemplate(context.Background(), emailTemplate{})
			tt.want.errAssertion(t, err)
		})
	}
}

func TestSendMailUnit(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.Empty(t, err)

	type input struct {
		ses sesClient
	}

	type want struct {
		errAssertion func(t *testing.T, err error)
	}

	cases := []struct {
		name  string
		input input
		want  want
	}{
		{
			"handles failure to send forward email",
			input{
				ses: &mockSESClient{
					sendEmailErrors: []string{"error sending forward email"},
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					require.NotEmpty(t, err)
					assert.Contains(t, err.Error(), "error sending forward email")
				},
			},
		},
		{
			"handles failure to send thank you email",
			input{
				ses: &mockSESClient{
					sendEmailErrors: []string{"", "error sending thank you email"},
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					require.NotEmpty(t, err)
					assert.Contains(t, err.Error(), "error sending thank you email")
				},
			},
		},
		{
			"is successful",
			input{
				ses: &mockSESClient{
					sendEmailErrors: []string{"", ""},
				},
			},
			want{
				errAssertion: func(t *testing.T, err error) {
					assert.Empty(t, err)
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			o := orchestrator{
				ses:    tt.input.ses,
				logger: logger,
			}

			_, err := o.SendMail(context.Background(), &mailservice_v1.SendMailRequest{})
			tt.want.errAssertion(t, err)
		})
	}
}

var _ sesClient = &mockSESClient{}

type mockSESClient struct {
	getEmailTemplateErr    string
	createEmailTemplateErr string
	updateEmailTemplateErr string

	// sendEmailErrors is a slice of boolean values that indicate whether an error should be returned when sending an email.
	// In the SendEmail funciton two emails are sent with sesClient so this allows us to control the error for each email.
	sendEmailErrors []string
}

func (m mockSESClient) GetEmailTemplate(ctx context.Context, params *sesv2.GetEmailTemplateInput, optFns ...func(*sesv2.Options)) (*sesv2.GetEmailTemplateOutput, error) {
	switch m.getEmailTemplateErr {
	case "NotFoundException":
		return nil, &types.NotFoundException{
			Message: aws.String("Template not found"),
		}
	case "":
		return &sesv2.GetEmailTemplateOutput{}, nil
	default:
		return nil, errors.New(m.getEmailTemplateErr)
	}
}

func (m mockSESClient) CreateEmailTemplate(ctx context.Context, params *sesv2.CreateEmailTemplateInput, optFns ...func(*sesv2.Options)) (*sesv2.CreateEmailTemplateOutput, error) {
	if m.createEmailTemplateErr != "" {
		return nil, errors.New(m.createEmailTemplateErr)
	}

	return &sesv2.CreateEmailTemplateOutput{}, nil
}

func (m mockSESClient) UpdateEmailTemplate(ctx context.Context, params *sesv2.UpdateEmailTemplateInput, optFns ...func(*sesv2.Options)) (*sesv2.UpdateEmailTemplateOutput, error) {
	if m.updateEmailTemplateErr != "" {
		return nil, errors.New(m.updateEmailTemplateErr)
	}

	return &sesv2.UpdateEmailTemplateOutput{}, nil
}

func (m *mockSESClient) SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	if len(m.sendEmailErrors) > 0 {
		err := m.sendEmailErrors[0]
		m.sendEmailErrors = m.sendEmailErrors[1:]
		if err != "" {
			return nil, errors.New(err)
		}
	}

	return &sesv2.SendEmailOutput{}, nil
}
