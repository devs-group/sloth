package email

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"

	"github.com/devs-group/sloth/backend/config"
)

//go:embed invitation.html
var InvitationTemplate []byte

func SendMail(url, invitationToken, receiver string) error {
	cfg := config.GetConfig()

	subject := "Hey, you got an invitation 👀\r\n"
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

	from := mail.Address{Name: "sloth", Address: cfg.SMTPFrom}
	to := mail.Address{Name: "", Address: receiver}

	connection := cfg.SMTPHost + ":" + cfg.SMTPPort
	message := []byte(
		fmt.Sprintf("From: %s <%s>\r\n", from.Name, from.Address) +
			fmt.Sprintf("To: %s\r\n", to.Address) +
			fmt.Sprintf("Subject: %s\r\n", subject) +
			"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
			body.String(),
	)

	var cl *smtp.Client
	cl, err = smtp.Dial(connection)
	if err != nil {
		return err
	}

	// Important to prevent sending from "localhost" .
	if err := cl.Hello("127.0.0.1"); err != nil {
		return err
	}

	// From
	if err = cl.Mail(from.Address); err != nil {
		return err
	}

	// To
	if err = cl.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := cl.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(message)
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	if err := cl.Quit(); err != nil {
		return err
	}

	return nil
}

func parseTemplate(templateData string, data interface{}) (string, error) {
	t, err := template.New("template").Parse(templateData)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return "", err
	}
	body := buf.String()
	return body, nil
}
