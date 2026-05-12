package main

import (
	"fmt"
	"sync"
	"time"
)

type MailTestResult struct {
	From     string
	To       string
	Success  bool
	Duration float64
	Error    string
}

var (
	MailTestResults []MailTestResult
	resultsMu       sync.Mutex
)

func appendResult(r MailTestResult) {
	resultsMu.Lock()
	defer resultsMu.Unlock()
	MailTestResults = append(MailTestResults, r)
}

func runSingleTest(fromCfg, toCfg ServerConfig, fromName, toName, recipientAddr string, timeout, poll time.Duration) {
	subject := fmt.Sprintf("Mail-Server-Test %s->%s %d", fromName, toName, time.Now().Unix())
	body := fmt.Sprintf("Testmail von %s an %s.", fromName, toName)
	start := time.Now()
	fmt.Printf("Sende Testmail von %s an %s...\n", fromName, toName)

	result := MailTestResult{From: fromName, To: toName}
	if err := SendTestMail(fromCfg, recipientAddr, subject, body); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Fehler beim Senden: %v", err)
		result.Duration = time.Since(start).Seconds()
		appendResult(result)
		fmt.Printf("Mail von %s an %s konnte nicht versendet werden: %v\n", fromName, toName, err)
		return
	}
	fmt.Printf("Mail %s->%s versendet, warte auf Zustellung (max %s, Poll alle %s)...\n", fromName, toName, timeout, poll)

	_, err := WaitAndCleanTestMail(toCfg, subject, timeout, poll)
	elapsed := time.Since(start)
	result.Duration = elapsed.Seconds()
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Fehler beim Abrufen: %v", err)
		fmt.Printf("Mail von %s an %s NICHT angekommen nach %s: %v\n", fromName, toName, elapsed, err)
	} else {
		result.Success = true
		fmt.Printf("Mail von %s an %s angekommen nach %s\n", fromName, toName, elapsed)
	}
	appendResult(result)
}

func RunMailTests(cfg *Config) error {
	timeoutMin := cfg.DeliveryTimeoutMinutes
	if timeoutMin <= 0 {
		timeoutMin = 30
	}
	pollSec := cfg.DeliveryPollSeconds
	if pollSec <= 0 {
		pollSec = 5
	}
	timeout := time.Duration(timeoutMin) * time.Minute
	poll := time.Duration(pollSec) * time.Second

	// Altlasten aus vorherigen (ggf. fehlgeschlagenen) Läufen entfernen, bevor
	// die parallelen Tests starten — danach räumt jeder Test nur noch seine
	// eigene Mail per exaktem Subject ab.
	CleanupOldTestMails(cfg.TestServer)
	for _, ext := range cfg.ExternalServers {
		CleanupOldTestMails(ext)
	}

	var wg sync.WaitGroup
	for _, ext := range cfg.ExternalServers {
		ext := ext

		recipientExt := ext.MailAddress
		if recipientExt == "" {
			recipientExt = ext.IMAPUser
		}
		recipientTestserver := cfg.TestServer.MailAddress
		if recipientTestserver == "" {
			recipientTestserver = cfg.TestServer.IMAPUser
		}

		wg.Add(2)
		go func() {
			defer wg.Done()
			runSingleTest(cfg.TestServer, ext, cfg.TestServer.Name, ext.Name, recipientExt, timeout, poll)
		}()
		go func() {
			defer wg.Done()
			runSingleTest(ext, cfg.TestServer, ext.Name, cfg.TestServer.Name, recipientTestserver, timeout, poll)
		}()
	}
	wg.Wait()
	return nil
}
