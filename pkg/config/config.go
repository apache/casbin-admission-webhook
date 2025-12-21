package config

import (
	"os"
	"strconv"
)

// Config holds the configuration for the admission webhook
type Config struct {
	// Server configuration
	Port     int
	CertFile string
	KeyFile  string

	// Casbin configuration
	ModelFile  string
	PolicyFile string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	port := 8443
	if p := os.Getenv("WEBHOOK_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	return &Config{
		Port:       port,
		CertFile:   getEnv("TLS_CERT_FILE", "/etc/webhook/certs/tls.crt"),
		KeyFile:    getEnv("TLS_KEY_FILE", "/etc/webhook/certs/tls.key"),
		ModelFile:  getEnv("CASBIN_MODEL_FILE", "/etc/webhook/casbin/model.conf"),
		PolicyFile: getEnv("CASBIN_POLICY_FILE", "/etc/webhook/casbin/policy.csv"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
