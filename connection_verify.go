package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"github.com/emersion/go-imap/client"
)

// ConnectionTestResult speichert das Ergebnis eines Verbindungstests
type ConnectionTestResult struct {
	ServerName string
	Protocol   string // "SMTP" oder "IMAP"
	Success    bool
	Error      string
	Duration   time.Duration
}

// VerifySMTPConnection testet die SMTP-Verbindung für einen Server
func VerifySMTPConnection(server ServerConfig) ConnectionTestResult {
	start := time.Now()
	result := ConnectionTestResult{
		ServerName: server.Name,
		Protocol:   "SMTP",
		Success:    false,
	}

	auth := smtp.PlainAuth("", server.SMTPUser, server.SMTPPassword, server.SMTPServer)
	addr := fmt.Sprintf("%s:%d", server.SMTPServer, server.SMTPPort)

	if server.TLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: server.SkipCertVerify,
			ServerName:         server.SMTPServer,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			result.Error = fmt.Sprintf("TLS-Verbindung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, server.SMTPServer)
		if err != nil {
			result.Error = fmt.Sprintf("SMTP-Client-Erstellung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
		defer c.Quit()

		if err := c.Auth(auth); err != nil {
			result.Error = fmt.Sprintf("SMTP-Authentifizierung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
	} else {
		// Für unverschlüsselte Verbindungen verwenden wir eine einfache Verbindung
		c, err := smtp.Dial(addr)
		if err != nil {
			result.Error = fmt.Sprintf("SMTP-Verbindung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
		defer c.Quit()

		if err := c.Auth(auth); err != nil {
			result.Error = fmt.Sprintf("SMTP-Authentifizierung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	result.Success = true
	result.Duration = time.Since(start)
	return result
}

// VerifyIMAPConnection testet die IMAP-Verbindung für einen Server
func VerifyIMAPConnection(server ServerConfig) ConnectionTestResult {
	start := time.Now()
	result := ConnectionTestResult{
		ServerName: server.Name,
		Protocol:   "IMAP",
		Success:    false,
	}

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
		result.Error = fmt.Sprintf("IMAP-Verbindung fehlgeschlagen: %v", err)
		result.Duration = time.Since(start)
		return result
	}
	defer c.Logout()

	if err := c.Login(server.IMAPUser, server.IMAPPassword); err != nil {
		result.Error = fmt.Sprintf("IMAP-Authentifizierung fehlgeschlagen: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Versuche INBOX zu öffnen um sicherzustellen, dass die Verbindung vollständig funktioniert
	_, err = c.Select("INBOX", true) // readonly=true
	if err != nil {
		result.Error = fmt.Sprintf("INBOX-Auswahl fehlgeschlagen: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = true
	result.Duration = time.Since(start)
	return result
}

// VerifyAllConnections testet alle Server-Verbindungen in der Konfiguration
func VerifyAllConnections(cfg *Config) []ConnectionTestResult {
	var results []ConnectionTestResult

	fmt.Println("=== Überprüfung der Zugangsdaten ===")
	fmt.Println()

	// Test des Testservers
	fmt.Printf("Teste Testserver '%s':\n", cfg.TestServer.Name)
	
	// SMTP Test
	fmt.Printf("  SMTP (%s:%d)... ", cfg.TestServer.SMTPServer, cfg.TestServer.SMTPPort)
	smtpResult := VerifySMTPConnection(cfg.TestServer)
	results = append(results, smtpResult)
	if smtpResult.Success {
		fmt.Printf("✓ OK (%.2fs)\n", smtpResult.Duration.Seconds())
	} else {
		fmt.Printf("✗ FEHLER: %s (%.2fs)\n", smtpResult.Error, smtpResult.Duration.Seconds())
	}

	// IMAP Test
	fmt.Printf("  IMAP (%s:%d)... ", cfg.TestServer.IMAPServer, cfg.TestServer.IMAPPort)
	imapResult := VerifyIMAPConnection(cfg.TestServer)
	results = append(results, imapResult)
	if imapResult.Success {
		fmt.Printf("✓ OK (%.2fs)\n", imapResult.Duration.Seconds())
	} else {
		fmt.Printf("✗ FEHLER: %s (%.2fs)\n", imapResult.Error, imapResult.Duration.Seconds())
	}

	fmt.Println()

	// Test der externen Server
	for _, server := range cfg.ExternalServers {
		fmt.Printf("Teste externen Server '%s':\n", server.Name)
		
		// SMTP Test
		fmt.Printf("  SMTP (%s:%d)... ", server.SMTPServer, server.SMTPPort)
		smtpResult := VerifySMTPConnection(server)
		results = append(results, smtpResult)
		if smtpResult.Success {
			fmt.Printf("✓ OK (%.2fs)\n", smtpResult.Duration.Seconds())
		} else {
			fmt.Printf("✗ FEHLER: %s (%.2fs)\n", smtpResult.Error, smtpResult.Duration.Seconds())
		}

		// IMAP Test
		fmt.Printf("  IMAP (%s:%d)... ", server.IMAPServer, server.IMAPPort)
		imapResult := VerifyIMAPConnection(server)
		results = append(results, imapResult)
		if imapResult.Success {
			fmt.Printf("✓ OK (%.2fs)\n", imapResult.Duration.Seconds())
		} else {
			fmt.Printf("✗ FEHLER: %s (%.2fs)\n", imapResult.Error, imapResult.Duration.Seconds())
		}

		fmt.Println()
	}

	// Zusammenfassung
	successCount := 0
	totalCount := len(results)
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	fmt.Println("=== Zusammenfassung ===")
	fmt.Printf("Erfolgreiche Verbindungen: %d/%d\n", successCount, totalCount)
	if successCount == totalCount {
		fmt.Println("✓ Alle Verbindungen erfolgreich!")
	} else {
		fmt.Printf("✗ %d Verbindung(en) fehlgeschlagen!\n", totalCount-successCount)
	}
	fmt.Println()

	return results
}
