package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"time"
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
		if err != nil {
			result.Error = fmt.Sprintf("TLS-Verbindung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
	} else {
		conn, err = dialer.Dial("tcp", addr)
		if err != nil {
			result.Error = fmt.Sprintf("SMTP-Verbindung fehlgeschlagen: %v", err)
			result.Duration = time.Since(start)
			return result
		}
	}

	c, err := smtp.NewClient(conn, server.SMTPServer)
	if err != nil {
		conn.Close()
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

	c, err := dialIMAP(server)
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
