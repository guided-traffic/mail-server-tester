package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
)

func SendTestMail(server ServerConfig, recipient, subject, body string) error {
	auth := smtp.PlainAuth("", server.SMTPUser, server.SMTPPassword, server.SMTPServer)

	fromAddr := server.MailAddress
	if fromAddr == "" {
		fromAddr = server.SMTPUser
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", fromAddr, recipient, subject, body))
	addr := net.JoinHostPort(server.SMTPServer, strconv.Itoa(server.SMTPPort))
	dialer := &net.Dialer{Timeout: dialNetTimeout}

	var conn net.Conn
	var err error
	if server.TLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: server.SkipCertVerify,
			ServerName:         server.SMTPServer,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, server.SMTPServer)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Quit()

	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(fromAddr); err != nil {
		return err
	}
	if err := c.Rcpt(recipient); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(msg); err != nil {
		return err
	}
	return wc.Close()
}
