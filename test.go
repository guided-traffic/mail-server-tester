package main

import (
  "fmt"
  "time"
)

type MailTestResult struct {
  From      string
  To        string
  Success   bool
  Duration  float64
  Error     string
}

var MailTestResults []MailTestResult

func RunMailTests(cfg *Config) error {
	// Testserver -> externe Server
	for _, ext := range cfg.ExternalServers {
		subject := fmt.Sprintf("Mail-Server-Test %s->%s %d", cfg.TestServer.Name, ext.Name, time.Now().Unix())
		body := fmt.Sprintf("Testmail von %s an %s.", cfg.TestServer.Name, ext.Name)
		start := time.Now()
		fmt.Printf("Sende Testmail von %s an %s...\n", cfg.TestServer.Name, ext.Name)
		result := MailTestResult{From: cfg.TestServer.Name, To: ext.Name}
		err := SendTestMail(cfg.TestServer, ext.IMAPUser, subject, body)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Fehler beim Senden: %v", err)
			MailTestResults = append(MailTestResults, result)
			continue
		}
		fmt.Println("Mail versendet, warte auf Zustellung...")
		time.Sleep(10 * time.Second)
		_, err = FetchLatestMail(ext)
		elapsed := time.Since(start)
		result.Duration = elapsed.Seconds()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Fehler beim Abrufen: %v", err)
		} else {
			result.Success = true
		}
		MailTestResults = append(MailTestResults, result)
		fmt.Printf("Mail von %s an %s angekommen nach %s\n", cfg.TestServer.Name, ext.Name, elapsed)
	}

	// Externe Server -> Testserver
	for _, ext := range cfg.ExternalServers {
		subject := fmt.Sprintf("Mail-Server-Test %s->%s %d", ext.Name, cfg.TestServer.Name, time.Now().Unix())
		body := fmt.Sprintf("Testmail von %s an %s.", ext.Name, cfg.TestServer.Name)
		start := time.Now()
		fmt.Printf("Sende Testmail von %s an %s...\n", ext.Name, cfg.TestServer.Name)
		result := MailTestResult{From: ext.Name, To: cfg.TestServer.Name}
		err := SendTestMail(ext, cfg.TestServer.IMAPUser, subject, body)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Fehler beim Senden: %v", err)
			MailTestResults = append(MailTestResults, result)
			continue
		}
		fmt.Println("Mail versendet, warte auf Zustellung...")
		time.Sleep(10 * time.Second)
		_, err = FetchLatestMail(cfg.TestServer)
		elapsed := time.Since(start)
		result.Duration = elapsed.Seconds()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Fehler beim Abrufen: %v", err)
		} else {
			result.Success = true
		}
		MailTestResults = append(MailTestResults, result)
		fmt.Printf("Mail von %s an %s angekommen nach %s\n", ext.Name, cfg.TestServer.Name, elapsed)
	}
	return nil
}
