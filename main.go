package main

import (
       "fmt"
       "os"
       "net/http"
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


func main() {
       fmt.Println("Mail Server Tester gestartet.")
       configPath := "config.yaml"
       for i := 1; i < len(os.Args)-1; i++ {
              if os.Args[i] == "--configpath" {
                    configPath = os.Args[i+1]
              }
       }
       cfg, err := LoadConfig(configPath)
       if err != nil {
              fmt.Fprintf(os.Stderr, "Fehler beim Laden der Konfiguration: %v\n", err)
              os.Exit(1)
       }

       // Starte HTTP-Server für Prometheus-Metrics
       http.HandleFunc("/metrics", metricsHandler)
       go func() {
              fmt.Println("Metrics Endpoint läuft auf :8080/metrics")
              http.ListenAndServe(":8080", nil)
       }()

       interval := time.Duration(cfg.IntervalMinutes)
       if interval == 0 {
              interval = 60
       }
       interval = interval * time.Minute

       for {
              MailTestResults = nil // Ergebnisse für jeden Durchlauf zurücksetzen
              fmt.Printf("Starte Mail-Tests (%s Intervall)...\n", interval)
              if err := RunMailTests(cfg); err != nil {
                     fmt.Fprintf(os.Stderr, "Testfehler: %v\n", err)
              } else {
                     fmt.Println("Alle Tests abgeschlossen.")
              }
              time.Sleep(interval)
       }
}
