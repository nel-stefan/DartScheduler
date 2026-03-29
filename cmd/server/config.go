package main

import "os"

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
}

func loadConfig() AppConfig {
	cfg := AppConfig{
		Port:         "8080",
		DatabaseType: "sqlite",
		DatabasePath: "dartscheduler.db",
		ClubName:     "DARTCLUB GROLZICHT",
		AppTitle:     "DartScheduler",
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
	if v := os.Getenv("CLUB_NAME"); v != "" {
		cfg.ClubName = v
	}
	if v := os.Getenv("APP_TITLE"); v != "" {
		cfg.AppTitle = v
	}
	if v := os.Getenv("LOGO_PATH"); v != "" {
		cfg.LogoPath = v
	}
	return cfg
}
