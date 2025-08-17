package main

import (
	"fmt"
	"os"
)


func main() {
       fmt.Println("Mail Server Tester gestartet.")
       cfg, err := LoadConfig("config.yaml")
       if err != nil {
	       fmt.Fprintf(os.Stderr, "Fehler beim Laden der Konfiguration: %v\n", err)
	       os.Exit(1)
       }
       if err := RunMailTests(cfg); err != nil {
	       fmt.Fprintf(os.Stderr, "Testfehler: %v\n", err)
	       os.Exit(1)
       }
       fmt.Println("Alle Tests abgeschlossen.")
}
