package main

import (
	"fmt"
	"time"
)




func RunMailTests(cfg *Config) error {
       // Testserver -> externe Server
       for _, ext := range cfg.ExternalServers {
	       subject := fmt.Sprintf("Mail-Server-Test %s->%s %d", cfg.TestServer.Name, ext.Name, time.Now().Unix())
	       body := fmt.Sprintf("Testmail von %s an %s.", cfg.TestServer.Name, ext.Name)
	       start := time.Now()
	       fmt.Printf("Sende Testmail von %s an %s...\n", cfg.TestServer.Name, ext.Name)
	       if err := SendTestMail(cfg.TestServer, ext.IMAPUser, subject, body); err != nil {
		       return fmt.Errorf("Fehler beim Senden von %s an %s: %w", cfg.TestServer.Name, ext.Name, err)
	       }
	       fmt.Println("Mail versendet, warte auf Zustellung...")
	       time.Sleep(10 * time.Second)
	       msg, err := FetchLatestMail(ext)
	       if err != nil {
		       return fmt.Errorf("Fehler beim Abrufen bei %s: %w", ext.Name, err)
	       }
	       elapsed := time.Since(start)
	       fmt.Printf("Mail von %s an %s angekommen nach %s\n", cfg.TestServer.Name, ext.Name, elapsed)
	       // TODO: Inhalt prüfen
       }

       // Externe Server -> Testserver
       for _, ext := range cfg.ExternalServers {
	       subject := fmt.Sprintf("Mail-Server-Test %s->%s %d", ext.Name, cfg.TestServer.Name, time.Now().Unix())
	       body := fmt.Sprintf("Testmail von %s an %s.", ext.Name, cfg.TestServer.Name)
	       start := time.Now()
	       fmt.Printf("Sende Testmail von %s an %s...\n", ext.Name, cfg.TestServer.Name)
	       if err := SendTestMail(ext, cfg.TestServer.IMAPUser, subject, body); err != nil {
		       return fmt.Errorf("Fehler beim Senden von %s an %s: %w", ext.Name, cfg.TestServer.Name, err)
	       }
	       fmt.Println("Mail versendet, warte auf Zustellung...")
	       time.Sleep(10 * time.Second)
	       msg, err := FetchLatestMail(cfg.TestServer)
	       if err != nil {
		       return fmt.Errorf("Fehler beim Abrufen bei %s: %w", cfg.TestServer.Name, err)
	       }
	       elapsed := time.Since(start)
	       fmt.Printf("Mail von %s an %s angekommen nach %s\n", ext.Name, cfg.TestServer.Name, elapsed)
	       // TODO: Inhalt prüfen
       }
       return nil
}
