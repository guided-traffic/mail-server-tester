package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test valid config
	configContent := `interval_minutes: 30
testserver:
  name: testserver
  smtp_server: smtp.test.com
  smtp_port: 587
  smtp_user: user@test.com
  smtp_password: password
  imap_server: imap.test.com
  imap_port: 993
  imap_user: user@test.com
  imap_password: password
  mail_address: test@test.com
  tls: true
  skip_cert_verify: false
external_servers:
  - name: external1
    smtp_server: smtp.ext1.com
    smtp_port: 587
    smtp_user: user@ext1.com
    smtp_password: password1
    imap_server: imap.ext1.com
    imap_port: 993
    imap_user: user@ext1.com
    imap_password: password1
    mail_address: test@ext1.com
    tls: true
    skip_cert_verify: true`

	// Create temporary config file
	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpFile.Close()

	// Test loading the config
	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify config values
	if cfg.IntervalMinutes != 30 {
		t.Errorf("Expected IntervalMinutes 30, got %d", cfg.IntervalMinutes)
	}

	if cfg.TestServer.Name != "testserver" {
		t.Errorf("Expected TestServer.Name 'testserver', got '%s'", cfg.TestServer.Name)
	}

	if cfg.TestServer.MailAddress != "test@test.com" {
		t.Errorf("Expected TestServer.MailAddress 'test@test.com', got '%s'", cfg.TestServer.MailAddress)
	}

	if len(cfg.ExternalServers) != 1 {
		t.Errorf("Expected 1 external server, got %d", len(cfg.ExternalServers))
	}

	if cfg.ExternalServers[0].MailAddress != "test@ext1.com" {
		t.Errorf("Expected external server MailAddress 'test@ext1.com', got '%s'", cfg.ExternalServers[0].MailAddress)
	}
}

func TestLoadConfigInvalidFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create temporary file with invalid YAML
	tmpFile, err := os.CreateTemp("", "invalid_config_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	invalidYAML := `invalid: yaml: content: [unclosed`
	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
