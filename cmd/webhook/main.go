package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/casbin/casbin-admission-webhook/pkg/config"
	"github.com/casbin/casbin-admission-webhook/pkg/webhook"
	"github.com/casbin/casbin/v2"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()

	klog.Info("Starting Casbin Admission Webhook")

	// Load configuration
	cfg := config.LoadConfig()
	klog.Infof("Configuration loaded: port=%d, model=%s, policy=%s", cfg.Port, cfg.ModelFile, cfg.PolicyFile)

	// Initialize Casbin enforcer
	enforcer, err := casbin.NewEnforcer(cfg.ModelFile, cfg.PolicyFile)
	if err != nil {
		klog.Fatalf("Failed to create Casbin enforcer: %v", err)
	}
	klog.Info("Casbin enforcer initialized successfully")

	// Create webhook server
	server := webhook.NewServer(enforcer)

	// Setup HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", server.HandleAdmission)
	mux.HandleFunc("/health", server.HandleHealth)
	mux.HandleFunc("/readyz", server.HandleHealth)

	// Load TLS certificates
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		klog.Fatalf("Failed to load TLS certificates: %v", err)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create HTTPS server
	httpServer := &http.Server{
		Addr:      fmt.Sprintf(":%d", cfg.Port),
		TLSConfig: tlsConfig,
		Handler:   mux,
	}

	// Start server in a goroutine
	go func() {
		klog.Infof("Starting HTTPS server on port %d", cfg.Port)
		if err := httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			klog.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	klog.Info("Shutting down webhook server")
	if err := httpServer.Close(); err != nil {
		klog.Errorf("Error closing server: %v", err)
	}
}
