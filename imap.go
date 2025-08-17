package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)



import "crypto/tls"

func FetchLatestMail(server ServerConfig) (*imap.Message, error) {
       addr := fmt.Sprintf("%s:%d", server.IMAPServer, server.IMAPPort)
       var c *client.Client
       var err error
       if server.TLS {
              tlsConfig := &tls.Config{
                     InsecureSkipVerify: server.SkipCertVerify,
                     ServerName:         server.IMAPServer,
              }
              c, err = client.DialTLS(addr, tlsConfig)
       } else {
              c, err = client.Dial(addr)
       }
       if err != nil {
              return nil, err
       }
       defer c.Logout()
       if err := c.Login(server.IMAPUser, server.IMAPPassword); err != nil {
              return nil, err
       }
       mbox, err := c.Select("INBOX", false)
       if err != nil {
              return nil, err
       }
       seqset := new(imap.SeqSet)
       if mbox.Messages == 0 {
              return nil, fmt.Errorf("Keine Nachrichten im Postfach")
       }
       seqset.AddNum(mbox.Messages)
       messages := make(chan *imap.Message, 1)
       section := &imap.BodySectionName{}
       if err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages); err != nil {
              return nil, err
       }
       msg := <-messages
       return msg, nil
}
