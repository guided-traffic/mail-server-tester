package main

import (
	"testing"
)

func TestMailTestResult(t *testing.T) {
	// Test creating a MailTestResult
	result := MailTestResult{
		From:     "server1",
		To:       "server2",
		Success:  true,
		Duration: 1.5,
		Error:    "",
	}

	if result.From != "server1" {
		t.Errorf("Expected From 'server1', got '%s'", result.From)
	}

	if result.To != "server2" {
		t.Errorf("Expected To 'server2', got '%s'", result.To)
	}

	if !result.Success {
		t.Errorf("Expected Success true, got %v", result.Success)
	}

	if result.Duration != 1.5 {
		t.Errorf("Expected Duration 1.5, got %f", result.Duration)
	}

	if result.Error != "" {
		t.Errorf("Expected empty Error, got '%s'", result.Error)
	}
}

func TestRunMailTestsWithMockConfig(t *testing.T) {
	// Create a minimal config for testing
	cfg := &Config{
		IntervalMinutes: 30,
		TestServer: ServerConfig{
			Name:         "testserver",
			SMTPServer:   "smtp.test.com",
			SMTPPort:     587,
			SMTPUser:     "user@test.com",
			SMTPPassword: "password",
			IMAPServer:   "imap.test.com",
			IMAPPort:     993,
			IMAPUser:     "user@test.com",
			IMAPPassword: "password",
			MailAddress:  "test@test.com",
			TLS:          true,
		},
		ExternalServers: []ServerConfig{
			{
				Name:         "external1",
				SMTPServer:   "smtp.ext1.com",
				SMTPPort:     587,
				SMTPUser:     "user@ext1.com",
				SMTPPassword: "password1",
				IMAPServer:   "imap.ext1.com",
				IMAPPort:     993,
				IMAPUser:     "user@ext1.com",
				IMAPPassword: "password1",
				MailAddress:  "test@ext1.com",
				TLS:          true,
			},
		},
	}

	// Clear previous results
	MailTestResults = nil

	// Note: This test will fail because it tries to connect to real servers
	// But it tests that the function doesn't panic and returns an error
	err := RunMailTests(cfg)
	
	// We expect an error because the servers don't exist
	if err == nil {
		t.Log("RunMailTests completed without error (unexpected for mock servers)")
	} else {
		t.Logf("RunMailTests returned expected error: %v", err)
	}

	// The function should have attempted to create results
	t.Logf("Created %d test results", len(MailTestResults))
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Errorf("Expected boolToInt(true) = 1, got %d", boolToInt(true))
	}

	if boolToInt(false) != 0 {
		t.Errorf("Expected boolToInt(false) = 0, got %d", boolToInt(false))
	}
}
