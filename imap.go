package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/textproto"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

const testMailSubjectPrefix = "Mail-Server-Test "

// dialNetTimeout begrenzt den TCP-Connect, damit unreachable Server keinen
// Test- oder Cleanup-Goroutine über Minuten blockieren.
const dialNetTimeout = 30 * time.Second

func dialIMAP(server ServerConfig) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", server.IMAPServer, server.IMAPPort)
	dialer := &net.Dialer{Timeout: dialNetTimeout}
	if server.TLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: server.SkipCertVerify,
			ServerName:         server.IMAPServer,
		}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		return client.New(conn)
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return client.New(conn)
}

// WaitAndCleanTestMail pollt das Postfach bis eine Mail mit exaktem Subject
// auftaucht oder das Timeout erreicht ist. Nach dem Fund wird genau diese Mail
// (nicht der gesamte Prefix-Match) gelöscht, damit parallele Tests gegen
// dasselbe Postfach sich nicht gegenseitig Mails wegputzen.
func WaitAndCleanTestMail(server ServerConfig, exactSubject string, timeout, pollInterval time.Duration) (*imap.Message, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		msg, found, err := tryFetchAndDelete(server, exactSubject)
		if err != nil {
			lastErr = err
		} else if found {
			return msg, nil
		}
		if time.Now().After(deadline) {
			if lastErr != nil {
				return nil, fmt.Errorf("Testmail mit Subject %q nicht innerhalb von %s gefunden, letzter Fehler: %v", exactSubject, timeout, lastErr)
			}
			return nil, fmt.Errorf("Testmail mit Subject %q nicht innerhalb von %s im Postfach gefunden", exactSubject, timeout)
		}
		time.Sleep(pollInterval)
	}
}

func tryFetchAndDelete(server ServerConfig, exactSubject string) (*imap.Message, bool, error) {
	c, err := dialIMAP(server)
	if err != nil {
		return nil, false, err
	}
	defer c.Logout()
	if err := c.Login(server.IMAPUser, server.IMAPPassword); err != nil {
		return nil, false, err
	}
	if _, err := c.Select("INBOX", false); err != nil {
		return nil, false, err
	}

	searchCriteria := &imap.SearchCriteria{
		Header: textproto.MIMEHeader{"Subject": {exactSubject}},
	}
	ids, err := c.Search(searchCriteria)
	if err != nil {
		return nil, false, err
	}
	if len(ids) == 0 {
		return nil, false, nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)
	messages := make(chan *imap.Message, len(ids))
	section := &imap.BodySectionName{}
	if err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages); err != nil {
		return nil, false, err
	}
	var msg *imap.Message
	for m := range messages {
		if msg == nil {
			msg = m
		}
	}

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.Store(seqset, item, flags, nil); err != nil {
		fmt.Printf("Setzen des \\Deleted-Flags fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
	} else if err := c.Expunge(nil); err != nil {
		fmt.Printf("Expunge fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
	}
	return msg, true, nil
}

// CleanupOldTestMails entfernt alle Testmails (Match per Subject-Prefix) aus
// dem Postfach. Wird einmal pro Lauf vor den eigentlichen Tests aufgerufen,
// um Altlasten aus abgebrochenen Läufen zu beseitigen. Fehler werden geloggt
// aber nicht propagiert, damit Cleanup-Probleme den Testlauf nicht blockieren.
func CleanupOldTestMails(server ServerConfig) {
	c, err := dialIMAP(server)
	if err != nil {
		fmt.Printf("Cleanup-Verbindung fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
		return
	}
	defer c.Logout()
	if err := c.Login(server.IMAPUser, server.IMAPPassword); err != nil {
		fmt.Printf("Cleanup-Login fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
		return
	}
	if _, err := c.Select("INBOX", false); err != nil {
		fmt.Printf("Cleanup-Select fehlgeschlagen für %s: %v\n", server.IMAPUser, err)
		return
	}
	cleanupTestMails(c, server)
}

// cleanupTestMails markiert alle Mails mit dem Test-Subject-Prefix als
// \Deleted und führt ein Expunge aus. Helper für CleanupOldTestMails.
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
	fmt.Printf("Cleanup: %d Altlast-Testmail(s) aus Postfach %s gelöscht\n", len(ids), server.IMAPUser)
}
