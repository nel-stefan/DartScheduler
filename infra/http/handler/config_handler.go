package handler

import "net/http"

// ConfigHandler serves read-only application configuration to the frontend.
type ConfigHandler struct {
	appTitle string
	clubName string
}

func NewConfigHandler(appTitle, clubName string) *ConfigHandler {
	return &ConfigHandler{appTitle: appTitle, clubName: clubName}
}

func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"appTitle": h.appTitle,
		"clubName": h.clubName,
	})
}
