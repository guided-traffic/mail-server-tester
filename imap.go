package main

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)


func FetchLatestMail(ext ExternalServer) (*imap.Message, error) {
       addr := fmt.Sprintf("%s:%d", ext.IMAPServer, ext.IMAPPort)
       c, err := client.DialTLS(addr, nil)
       if err != nil {
	       return nil, err
       }
       defer c.Logout()
       if err := c.Login(ext.IMAPUser, ext.IMAPPassword); err != nil {
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
