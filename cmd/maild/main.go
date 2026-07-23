// Package main is the entry point for the maild local mail composer.
//
// maild listens on 127.0.0.1:8080 and provides a browser-based UI
// for composing and sending emails via the Resend API using your own
// domain identity.
package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"maild/internal/config"
	"maild/internal/provider"
	"maild/internal/server"
)

func main() {
	// CLI flags.
	configPath := flag.String("config", "config.toml", "Path to TOML configuration file")
	webDir := flag.String("web", "web", "Path to web assets directory")
	flag.Parse()

	// If the first positional arg looks like a config path (legacy support), use it.
	if flag.NArg() > 0 && *configPath == "config.toml" {
		*configPath = flag.Arg(0)
	}

	// Resolve web directory.
	absWebDir, err := filepath.Abs(*webDir)
	if err != nil {
		log.Fatalf("ERROR resolve web directory: %v", err)
	}
	if _, err := os.Stat(absWebDir); os.IsNotExist(err) {
		cfgDir := filepath.Dir(*configPath)
		alt := filepath.Join(cfgDir, *webDir)
		if _, err2 := os.Stat(alt); err2 == nil {
			absWebDir = alt
		} else {
			log.Fatalf("ERROR web directory not found: %s (or %s)", absWebDir, alt)
		}
	}

	// Load configuration.
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("ERROR load config: %v", err)
	}

	// Initialize the Resend mail provider.
	p := provider.NewResendProvider(cfg.Resend.APIKey)

	// Create the HTTP server.
	srv := server.New(cfg, p, absWebDir)

	// Startup log sequence.
	log.Printf("INFO  maild starting")
	log.Printf("INFO  listen: %s", cfg.Server.Address)
	log.Printf("INFO  web root: %s", absWebDir)
	log.Printf("INFO  provider: resend")
	log.Printf("INFO  ready")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("ERROR server: %v", err)
	}
}
