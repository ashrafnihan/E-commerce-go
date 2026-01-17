package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

type Mailer interface {
	Send(to, subject, body string) error
}

type smtpMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) Mailer {
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) Send(to, subject, body string) error {
	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))

	// connect
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return err
	}
	defer c.Close()

	// STARTTLS
	tlsCfg := &tls.Config{ServerName: m.cfg.Host}
	if err := c.StartTLS(tlsCfg); err != nil {
		return err
	}

	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
	if err := c.Auth(auth); err != nil {
		return err
	}

	if err := c.Mail(m.cfg.User); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	msg := buildMessage(m.cfg.From, to, subject, body)
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	return c.Quit()
}

func buildMessage(from, to, subject, body string) string {
	// Very standard website email formatting
	headers := ""
	headers += fmt.Sprintf("From: %s\r\n", from)
	headers += fmt.Sprintf("To: %s\r\n", to)
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/plain; charset=UTF-8\r\n"
	headers += "\r\n"
	return headers + body + "\r\n"
}
