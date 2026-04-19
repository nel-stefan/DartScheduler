package main

import (
	"log"
	"os"
)

// AppConfig holds all runtime configuration for the server, loaded from
// environment variables. Every field has a safe default so existing
// deployments that set no variables continue to work unchanged.
type AppConfig struct {
	// Port is the TCP port the HTTP server listens on.
	Port string

	// DatabaseType selects the storage backend. Currently only "sqlite" is
	// supported; the field exists for future extensibility.
	DatabaseType string

	// DatabasePath is the filesystem path to the SQLite database file.
	DatabasePath string

	// ClubName is printed as the heading on every match-form export
	// (Excel, PDF, HTML). Set CLUB_NAME to your club's name.
	ClubName string

	// AppTitle is displayed in the Angular toolbar and browser tab.
	AppTitle string

	// LogoPath is an optional path to a PNG or JPEG logo image. When set,
	// the logo is rendered at the top of PDF and HTML match forms.
	// Leave empty to omit the logo.
	LogoPath string

	// AllowedOrigin is the value of the Access-Control-Allow-Origin header.
	// Use "*" for development; set to your frontend origin in production.
	AllowedOrigin string

	// PrimaryColor is an optional CSS colour (e.g. "#e65100") that overrides
	// the toolbar background. Leave empty to use the default theme colour.
	// Useful for distinguishing demo from production deployments.
	PrimaryColor string

	// JWTSecret is the HS256 signing secret for JWT tokens.
	// Set via JWT_SECRET env var. Falls back to an insecure default in development.
	JWTSecret string
}

func loadConfig() AppConfig {
	cfg := AppConfig{
		Port:          "8080",
		DatabaseType:  "sqlite",
		DatabasePath:  "dartscheduler.db",
		ClubName:      "DARTCLUB GROLZICHT",
		AppTitle:      "DartScheduler",
		AllowedOrigin: "*",
	}
	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
	}
	if v := os.Getenv("DATABASE_TYPE"); v != "" {
		cfg.DatabaseType = v
	}
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("CLUB_NAAM"); v != "" {
		cfg.ClubName = v
	} else if v := os.Getenv("CLUB_NAME"); v != "" {
		cfg.ClubName = v
	}
	if v := os.Getenv("APP_TITLE"); v != "" {
		cfg.AppTitle = v
	}
	if v := os.Getenv("LOGO_PATH"); v != "" {
		cfg.LogoPath = v
	}
	if v := os.Getenv("ALLOWED_ORIGIN"); v != "" {
		cfg.AllowedOrigin = v
	}
	if v := os.Getenv("PRIMARY_COLOR"); v != "" {
		cfg.PrimaryColor = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWTSecret = v
	} else {
		log.Println("[WARN] JWT_SECRET not set — using insecure default (development only)")
		cfg.JWTSecret = "dart-scheduler-dev-secret-change-in-production"
	}
	return cfg
}
