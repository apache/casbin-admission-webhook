package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test default values
	cfg := LoadConfig()
	if cfg.Port != 8443 {
		t.Errorf("Expected default port 8443, got %d", cfg.Port)
	}
	if cfg.CertFile != "/etc/webhook/certs/tls.crt" {
		t.Errorf("Expected default cert file, got %s", cfg.CertFile)
	}
	if cfg.KeyFile != "/etc/webhook/certs/tls.key" {
		t.Errorf("Expected default key file, got %s", cfg.KeyFile)
	}
	if cfg.ModelFile != "/etc/webhook/casbin/model.conf" {
		t.Errorf("Expected default model file, got %s", cfg.ModelFile)
	}
	if cfg.PolicyFile != "/etc/webhook/casbin/policy.csv" {
		t.Errorf("Expected default policy file, got %s", cfg.PolicyFile)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("WEBHOOK_PORT", "9443")
	os.Setenv("TLS_CERT_FILE", "/custom/cert.crt")
	os.Setenv("TLS_KEY_FILE", "/custom/key.key")
	os.Setenv("CASBIN_MODEL_FILE", "/custom/model.conf")
	os.Setenv("CASBIN_POLICY_FILE", "/custom/policy.csv")
	defer func() {
		os.Unsetenv("WEBHOOK_PORT")
		os.Unsetenv("TLS_CERT_FILE")
		os.Unsetenv("TLS_KEY_FILE")
		os.Unsetenv("CASBIN_MODEL_FILE")
		os.Unsetenv("CASBIN_POLICY_FILE")
	}()

	cfg := LoadConfig()
	if cfg.Port != 9443 {
		t.Errorf("Expected port 9443, got %d", cfg.Port)
	}
	if cfg.CertFile != "/custom/cert.crt" {
		t.Errorf("Expected custom cert file, got %s", cfg.CertFile)
	}
	if cfg.KeyFile != "/custom/key.key" {
		t.Errorf("Expected custom key file, got %s", cfg.KeyFile)
	}
	if cfg.ModelFile != "/custom/model.conf" {
		t.Errorf("Expected custom model file, got %s", cfg.ModelFile)
	}
	if cfg.PolicyFile != "/custom/policy.csv" {
		t.Errorf("Expected custom policy file, got %s", cfg.PolicyFile)
	}
}

func TestLoadConfigInvalidPort(t *testing.T) {
	os.Setenv("WEBHOOK_PORT", "invalid")
	defer os.Unsetenv("WEBHOOK_PORT")

	cfg := LoadConfig()
	if cfg.Port != 8443 {
		t.Errorf("Expected default port 8443 when invalid port specified, got %d", cfg.Port)
	}
}
