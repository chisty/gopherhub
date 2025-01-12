package mailer

import (
	"github.com/sendgrid/sendgrid-go"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGridMailer(fromEmail, apiKey string) *SendGridMailer {
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    sendgrid.NewSendClient(apiKey),
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	// from := mail.NewEmail(fromEmail, m.fromEmail)
	return nil
}
