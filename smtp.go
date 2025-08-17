package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)


func SendTestMail(smtpCfg SMTPConfig, recipient, subject, body string) error {
	auth := smtp.PlainAuth("", smtpCfg.User, smtpCfg.Password, smtpCfg.Server)
	msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))
	addr := fmt.Sprintf("%s:%d", smtpCfg.Server, smtpCfg.Port)
	return smtp.SendMail(addr, auth, smtpCfg.User, []string{recipient}, msg)
}
