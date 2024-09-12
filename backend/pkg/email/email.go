package email

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/devs-group/sloth/backend/config"
)

//go:embed invitation.html
var InvitationTemplate []byte

func SendMail(url, invitationToken, to string) error {
	from := config.SMTPFrom
	password := config.SMTPPassword

	SMTPHost := config.SMTPHost
	SMTPPort := config.SMTPPort

	template, err := template.New("invitation").Parse(string(InvitationTemplate))
	if err != nil {
		return fmt.Errorf("unable to parse email template: %w", err)
	}

	data := struct {
		Link string
	}{
		Link: url + "=" + invitationToken,
	}

	var body bytes.Buffer
	if err := template.Execute(&body, data); err != nil {
		return fmt.Errorf("unable to pass data to email template: %w", err)
	}

	subject := "Hey, you got an invitation 👀\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	msg := []byte(subject + mime + body.String())

	auth := smtp.PlainAuth("", from, password, SMTPHost)
	if err := smtp.SendMail(SMTPHost+":"+SMTPPort, auth, from, []string{to}, msg); err != nil {
		return fmt.Errorf("unable to send email: %w", err)
	}

	return nil
}
