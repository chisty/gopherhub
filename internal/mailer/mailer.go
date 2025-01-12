package mailer

const (
	FromName   = "GopherHub"
	MaxRetries = 3
)

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}
