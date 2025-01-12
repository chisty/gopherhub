package mailer

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	subject := new(bytes.Buffer)
	body := new(bytes.Buffer)

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	for i := 0; i < MaxRetries; i++ {
		resp, err := m.client.Send(message)
		if err != nil {
			log.Printf("failed to send email to %s, attempt %d: %v", email, i+1, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Printf("email sent to %s, status code: %d", email, resp.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", MaxRetries)

}
