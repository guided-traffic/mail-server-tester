package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)



import (
       "crypto/tls"
       "net"
)

func SendTestMail(server ServerConfig, recipient, subject, body string) error {
       auth := smtp.PlainAuth("", server.SMTPUser, server.SMTPPassword, server.SMTPServer)
       msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))
       addr := fmt.Sprintf("%s:%d", server.SMTPServer, server.SMTPPort)
       if server.TLS {
	       tlsConfig := &tls.Config{
		       InsecureSkipVerify: server.SkipCertVerify,
		       ServerName:         server.SMTPServer,
	       }
	       conn, err := tls.Dial("tcp", addr, tlsConfig)
	       if err != nil {
		       return err
	       }
	       c, err := smtp.NewClient(conn, server.SMTPServer)
	       if err != nil {
		       return err
	       }
	       defer c.Quit()
	       if err := c.Auth(auth); err != nil {
		       return err
	       }
	       if err := c.Mail(server.SMTPUser); err != nil {
		       return err
	       }
	       if err := c.Rcpt(recipient); err != nil {
		       return err
	       }
	       wc, err := c.Data()
	       if err != nil {
		       return err
	       }
	       _, err = wc.Write(msg)
	       if err != nil {
		       return err
	       }
	       err = wc.Close()
	       if err != nil {
		       return err
	       }
	       return nil
       } else {
	       return smtp.SendMail(addr, auth, server.SMTPUser, []string{recipient}, msg)
       }
}
