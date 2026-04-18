// infrastructure/email/email.go (версия с поддержкой UTF-8 в теме)
package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

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

func encodeSubject(subject string) string {
	return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
}

func (s *SMTPSender) Send(ctx context.Context, email domain.Email) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	encodedSubject := encodeSubject(email.Subject)

	headers := []string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", email.To),
		fmt.Sprintf("Subject: %s", encodedSubject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}

	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + email.Body

	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         s.host,
	})
	if err != nil {
		s.log.Error("failed to connect to SMTP server", slog.String("error", err.Error()))
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		s.log.Error("failed to create SMTP client", slog.String("error", err.Error()))
		return err
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		s.log.Error("authentication failed", slog.String("error", err.Error()))
		return err
	}

	if err := client.Mail(s.from); err != nil {
		s.log.Error("failed to set sender", slog.String("error", err.Error()))
		return err
	}

	if err := client.Rcpt(email.To); err != nil {
		s.log.Error("failed to set recipient", slog.String("error", err.Error()))
		return err
	}

	w, err := client.Data()
	if err != nil {
		s.log.Error("failed to get data writer", slog.String("error", err.Error()))
		return err
	}
	defer w.Close()

	if _, err := w.Write([]byte(message)); err != nil {
		s.log.Error("failed to write message", slog.String("error", err.Error()))
		return err
	}

	return client.Quit()
}
