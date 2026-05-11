package main

import (
	"crypto/tls"
	"fmt"
	"net/textproto"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

const testMailSubjectPrefix = "Mail-Server-Test "

func dialIMAP(server ServerConfig) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", server.IMAPServer, server.IMAPPort)
	if server.TLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: server.SkipCertVerify,
			ServerName:         server.IMAPServer,
		}
		return client.DialTLS(addr, tlsConfig)
	}
	return client.Dial(addr)
}

// FetchAndCleanTestMails sucht die Mail mit exakt diesem Subject (Verifikation
// des aktuellen Tests) und löscht anschließend ALLE Mails mit dem Test-Prefix
// (inklusive Altlasten aus vorherigen Läufen).
func FetchAndCleanTestMails(server ServerConfig, exactSubject string) (*imap.Message, error) {
	c, err := dialIMAP(server)
	if err != nil {
		return nil, err
	}
	defer c.Logout()
	if err := c.Login(server.IMAPUser, server.IMAPPassword); err != nil {
		return nil, err
	}
	if _, err := c.Select("INBOX", false); err != nil {
		return nil, err
	}

	searchCriteria := &imap.SearchCriteria{
		Header: textproto.MIMEHeader{"Subject": {exactSubject}},
	}
	ids, err := c.Search(searchCriteria)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		cleanupTestMails(c, server)
		return nil, fmt.Errorf("Testmail mit Subject %q nicht im Postfach gefunden", exactSubject)
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids[len(ids)-1])
	messages := make(chan *imap.Message, 1)
	section := &imap.BodySectionName{}
	if err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages); err != nil {
		return nil, err
	}
	msg := <-messages

	cleanupTestMails(c, server)
	return msg, nil
}

// cleanupTestMails markiert alle Mails mit dem Test-Subject-Prefix als \Deleted
// und führt ein Expunge aus. Fehler werden geloggt aber nicht propagiert, damit
// ein Cleanup-Problem den Test-Erfolg nicht überschreibt.
func cleanupTestMails(c *client.Client, server ServerConfig) {
	prefixCriteria := &imap.SearchCriteria{
		Header: textproto.MIMEHeader{"Subject": {testMailSubjectPrefix}},
	}
	ids, err := c.Search(prefixCriteria)
	if err != nil {
		fmt.Printf("Cleanup-Suche fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
		return
	}
	if len(ids) == 0 {
		return
	}
	delSet := new(imap.SeqSet)
	delSet.AddNum(ids...)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.Store(delSet, item, flags, nil); err != nil {
		fmt.Printf("Setzen des \\Deleted-Flags fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
		return
	}
	if err := c.Expunge(nil); err != nil {
		fmt.Printf("Expunge fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
		return
	}
	fmt.Printf("Cleanup: %d Testmail(s) aus Postfach %s gelöscht\n", len(ids), server.IMAPUser)
}
