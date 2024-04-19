package email

import (
	"fmt"
	"net/smtp"

	"github.com/devs-group/sloth/backend/config"
)

func SendMail(url, invitationToken, to string) error {
	from := config.SMTPFrom
	password := config.SMTPPW

	SMTPHost := config.SMTPHost
	SMTPPort := config.SMTPPort

	msg := fmt.Sprintf("You have been invited follow this link to accept the invitation %s=%s", url, invitationToken)
	message := []byte(msg)

	auth := smtp.PlainAuth("", from, password, SMTPHost)
	if err := smtp.SendMail(SMTPHost+":"+SMTPPort, auth, from, []string{to}, message); err != nil {
		return err
	}

	return nil
}
