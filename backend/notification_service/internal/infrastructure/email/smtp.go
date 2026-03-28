package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"

	"github.com/Artem09076/dp/backend/notification_service/internal/domain"
)

type SMTPSender struct {
	log  *slog.Logger
	host string
	port int
	user string
	pass string
	from string
}

func NewSMTPSender(log *slog.Logger, host string, port int, user, pass, from string) *SMTPSender {
	return &SMTPSender{
		log:  log,
		host: host,
		port: port,
		user: user,
		pass: pass,
		from: from,
	}
}

func (s *SMTPSender) Send(ctx context.Context, email domain.Email) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-version: 1.0;\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\";\r\n\r\n"+
			"%s",
		s.from, email.To, email.Subject, email.Body,
	))

	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
	})
	if err != nil {
		s.log.Info(err.Error())
		return err
	}
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		s.log.Info(err.Error())
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		s.log.Info(err.Error())
		return err
	}

	if err := client.Mail(s.from); err != nil {
		s.log.Info(err.Error())
		return err
	}

	if err := client.Rcpt(email.To); err != nil {
		s.log.Info(err.Error())
		return err
	}
	w, err := client.Data()
	if err != nil {
		s.log.Info(err.Error())
		return err
	}
	if _, err := w.Write(msg); err != nil {
		s.log.Info(err.Error())
		return err
	}
	if err := w.Close(); err != nil {
		s.log.Info(err.Error())
		return err
	}
	s.log.Info("dfghjtrewd")

	return client.Quit()
}
