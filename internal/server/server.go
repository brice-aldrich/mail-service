package server

import (
	"context"

	mailservice_v1 "github.com/brice-aldrich/mail-service/gen/go/mailservice.v1"
	"github.com/brice-aldrich/mail-service/internal/mail"
)

// server implements the mailservice_v1.MailServiceServer interface.
// It holds a reference to the mail orchestrator which is used to handle email sending operations.
type server struct {
	mailOrch mail.Orchestrator
	mailservice_v1.UnimplementedMailServiceServer
}

// New creates a new instance of the server with the provided mail orchestrator.
// It returns an implementation of the mailservice_v1.MailServiceServer interface.
//
// Parameters:
//   - mailOrch: The mail.Orchestrator object used to handle email sending operations.
//
// Returns:
//   - mailservice_v1.MailServiceServer: The newly created server instance.
func New(mailOrch mail.Orchestrator) mailservice_v1.MailServiceServer {
	return &server{
		mailOrch: mailOrch,
	}
}

// SendMail handles the SendMail request by delegating the operation to the mail orchestrator.
// It sends an email based on the provided request and returns the response.
//
// Parameters:
//   - ctx: The context.Context object for the request.
//   - req: The mailservice_v1.SendMailRequest object containing the email message and recipient information.
//
// Returns:
//   - *mailservice_v1.SendMailResponse: The response object indicating the result of the send mail operation.
//   - error: An error if any occurred during the sending of the email.
func (s server) SendMail(ctx context.Context, req *mailservice_v1.SendMailRequest) (*mailservice_v1.SendMailResponse, error) {
	return s.mailOrch.SendMail(ctx, req)
}
