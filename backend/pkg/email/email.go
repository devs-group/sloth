package email

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/devs-group/sloth/backend/config"
	"github.com/devs-group/sloth/backend/utils"
	"html/template"
	"net/mail"
	"net/smtp"
)

//go:embed invitation.html
var InvitationTemplate []byte

func SendInvitationMail(url, invitationToken, receiver string) error {
	cfg := config.GetConfig()

	subject := "Hey, you got an invitation 👀\r\n"
	tpl, err := template.New("invitation").Parse(string(InvitationTemplate))
	if err != nil {
		return fmt.Errorf("unable to parse email template: %w", err)
	}
	data := struct {
		Link string
	}{
		Link: url + "=" + invitationToken,
	}
	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return fmt.Errorf("unable to pass data to email template: %w", err)
	}

	from := mail.Address{Name: "sloth", Address: cfg.SMTPFrom}
	to := mail.Address{Name: "", Address: receiver}

	headers := map[string]string{
		"From":         from.String(),
		"To":           to.Address,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n") // 🔥 Crucial blank line between headers & body
	message.WriteString(body.String())

	connection := cfg.SMTPHost + ":" + cfg.SMTPPort

	var auth smtp.Auth
	if utils.IsProduction() {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}
	err = smtp.SendMail(connection, auth, from.Address, []string{to.Address}, message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
