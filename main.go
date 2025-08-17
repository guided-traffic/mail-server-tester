package main

import (
	"fmt"
	"os"
)




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
       if err := RunMailTests(cfg); err != nil {
              fmt.Fprintf(os.Stderr, "Testfehler: %v\n", err)
              os.Exit(1)
       }
       fmt.Println("Alle Tests abgeschlossen.")
}
