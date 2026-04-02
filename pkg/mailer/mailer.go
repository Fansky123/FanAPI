package mailer

import (
	"crypto/tls"
	"fanapi/internal/config"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	cfg *config.SMTPConfig
}

func New(cfg *config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(to, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	d := gomail.NewDialer(m.cfg.Host, m.cfg.Port, m.cfg.User, m.cfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: false}
	return d.DialAndSend(msg)
}
