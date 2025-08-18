package main

import (
  "crypto/tls"
  "fmt"
  "net/smtp"
)

func SendTestMail(server ServerConfig, recipient, subject, body string) error {
  auth := smtp.PlainAuth("", server.SMTPUser, server.SMTPPassword, server.SMTPServer)

  // Verwende mail_address falls definiert, sonst fallback auf smtp_user
  fromAddr := server.MailAddress
  if fromAddr == "" {
    fromAddr = server.SMTPUser
  }

  msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", fromAddr, recipient, subject, body))
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
    return smtp.SendMail(addr, auth, fromAddr, []string{recipient}, msg)
  }
}
