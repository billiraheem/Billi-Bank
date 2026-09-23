package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddr = "smtp.gmail.com"
	smtpServerAddr = "smtp.gmail.com:587"
)

type EmailSender interface {
	Send(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachedFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddr     string
	fromEmailPassword string
}

func NewGmailSender(name, fromEmailAddr, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddr:     fromEmailAddr,
		fromEmailPassword: fromEmailPassword,
	}
}

func (sender *GmailSender) Send(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachedFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddr)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, f := range attachedFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddr, sender.fromEmailPassword, smtpAuthAddr)
	return e.Send(smtpServerAddr, smtpAuth)
}
