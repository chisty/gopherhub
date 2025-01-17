package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGridMailer(fromEmail, apiKey string) (*SendGridMailer, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("sendgrid api key is required")
	}
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    sendgrid.NewSendClient(apiKey),
	}, nil
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return err
	}

	log.Printf("subject: %s\n\n, body: %s\n\n", subject.String(), body.String())

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

		log.Printf("email sent to %s, status code: %d, resp: %v", email, resp.StatusCode, resp)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", MaxRetries)

}
