package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	var sb strings.Builder
	sb.WriteString("# HELP mail_test_success 1=Erfolg, 0=Fehler\n")
	sb.WriteString("# TYPE mail_test_success gauge\n")
	sb.WriteString("# HELP mail_test_duration_seconds Dauer des Tests in Sekunden\n")
	sb.WriteString("# TYPE mail_test_duration_seconds gauge\n")
	sb.WriteString("# HELP mail_test_error_total Fehleranzahl pro Test\n")
	sb.WriteString("# TYPE mail_test_error_total counter\n")
	for _, res := range MailTestResults {
		labels := fmt.Sprintf("from=\"%s\",to=\"%s\"", res.From, res.To)
		sb.WriteString(fmt.Sprintf("mail_test_success{%s} %d\n", labels, boolToInt(res.Success)))
		sb.WriteString(fmt.Sprintf("mail_test_duration_seconds{%s} %.2f\n", labels, res.Duration))
		if res.Error != "" {
			sb.WriteString(fmt.Sprintf("mail_test_error_total{%s} 1\n", labels))
		} else {
			sb.WriteString(fmt.Sprintf("mail_test_error_total{%s} 0\n", labels))
		}
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(sb.String()))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func printUsage() {
	fmt.Println("Usage: mail-server-tester [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --configpath PATH              Pfad zur Konfigurationsdatei (default: config.yaml)")
	fmt.Println("  --exit-on-connection-error     Beende das Programm bei Verbindungsfehlern")
	fmt.Println("  --help                         Zeige diese Hilfe an")
}

func main() {
	fmt.Println("Mail Server Tester gestartet.")
	configPath := "config.yaml"
	exitOnError := false

	// Erste Durchlauf: Prüfe auf --help vor allen anderen Operationen
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--help" || os.Args[i] == "-h" {
			printUsage()
			os.Exit(0)
		}
	}

	// Zweite Durchlauf: Parse alle anderen Argumente
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--configpath" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			i++ // Skip next argument since it's the config path
		} else if os.Args[i] == "--exit-on-connection-error" {
			exitOnError = true
		} else if os.Args[i] != "--help" && os.Args[i] != "-h" {
			fmt.Fprintf(os.Stderr, "Unbekanntes Argument: %s\n", os.Args[i])
			printUsage()
			os.Exit(1)
		}
	}
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler beim Laden der Konfiguration: %v\n", err)
		os.Exit(1)
	}

	// Überprüfe alle Verbindungen beim Start
	connectionResults := VerifyAllConnections(cfg)

	// Prüfe ob kritische Fehler aufgetreten sind
	hasErrors := false
	for _, result := range connectionResults {
		if !result.Success {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		if exitOnError {
			fmt.Println("❌ Kritische Verbindungsfehler festgestellt. Programm wird beendet.")
			os.Exit(1)
		} else {
			fmt.Println("⚠️  Warnung: Es wurden Verbindungsfehler festgestellt. Das Programm wird fortgesetzt, aber einige Tests könnten fehlschlagen.")
			fmt.Println()
		}
	}

	// Start HTTP server for Prometheus metrics
	http.HandleFunc("/metrics", metricsHandler)
	go func() {
		fmt.Println("Metrics endpoint running on :8080/metrics")
		http.ListenAndServe(":8080", nil)
	}()

	interval := time.Duration(cfg.IntervalMinutes)
	if interval == 0 {
		interval = 60
	}
	interval = interval * time.Minute

	for {
		MailTestResults = nil // Reset results for each run
		fmt.Printf("Starting mail tests (%s interval)...\n", interval)
		if err := RunMailTests(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Testfehler: %v\n", err)
		} else {
			fmt.Println("Alle Tests abgeschlossen.")
		}
		time.Sleep(interval)
	}
}
