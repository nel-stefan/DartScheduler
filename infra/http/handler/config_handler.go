package handler

import "net/http"

// ConfigHandler serves read-only application configuration to the frontend.
type ConfigHandler struct {
	appTitle     string
	clubName     string
	primaryColor string
}

func NewConfigHandler(appTitle, clubName, primaryColor string) *ConfigHandler {
	return &ConfigHandler{appTitle: appTitle, clubName: clubName, primaryColor: primaryColor}
}

func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"appTitle":     h.appTitle,
		"clubName":     h.clubName,
		"primaryColor": h.primaryColor,
	})
}
