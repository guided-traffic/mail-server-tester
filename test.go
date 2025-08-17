package main

import (
	"fmt"
	"time"
)


func RunMailTests(cfg *Config) error {
       for _, ext := range cfg.ExternalServers {
	       subject := fmt.Sprintf("Mail-Server-Test %d", time.Now().Unix())
	       body := "Dies ist eine Testmail."
	       start := time.Now()
	       fmt.Printf("Sende Testmail an %s...\n", ext.Recipient)
	       if err := SendTestMail(cfg.SMTP, ext.Recipient, subject, body); err != nil {
		       return fmt.Errorf("Fehler beim Senden an %s: %w", ext.Recipient, err)
	       }
	       fmt.Println("Mail versendet, warte auf Zustellung...")
	       time.Sleep(10 * time.Second) // Warten auf Zustellung
	       msg, err := FetchLatestMail(ext)
	       if err != nil {
		       return fmt.Errorf("Fehler beim Abrufen von %s: %w", ext.IMAPUser, err)
	       }
	       elapsed := time.Since(start)
	       fmt.Printf("Mail an %s angekommen nach %s\n", ext.Recipient, elapsed)
	       // TODO: Inhalt prüfen
       }
       return nil
}
