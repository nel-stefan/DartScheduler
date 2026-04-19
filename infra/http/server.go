// Package http registers all API routes and mounts the Angular SPA handler.
package http

import (
	"net/http"

	"DartScheduler/infra/http/handler"
	mw "DartScheduler/infra/http/middleware"
	"DartScheduler/web"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	playerH *handler.PlayerHandler,
	schedH *handler.ScheduleHandler,
	scoreH *handler.ScoreHandler,
	statsH *handler.StatsHandler,
	exportH *handler.ExportHandler,
	systemH *handler.SystemHandler,
	eveningStatH *handler.EveningStatHandler,
	seasonStatH *handler.SeasonStatHandler,
	configH *handler.ConfigHandler,
	progressH *handler.ProgressHandler,
	playerListH *handler.PlayerListHandler,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	allowedOrigin string,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(mw.Logger)
	r.Use(mw.CORSWithOrigin(allowedOrigin))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		// Public — no authentication required
		r.Get("/config", configH.GetConfig)
		r.Post("/auth/login", authH.Login)
		r.Get("/auth/me", authH.Me)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(jwtSecret))

			// viewer+ — any authenticated identity
			r.Get("/schedules", schedH.List)
			r.Get("/schedules/{id}", schedH.GetByID)
			r.Get("/schedule", schedH.Get)
			r.Get("/schedule/evening/{id}", schedH.GetEvening)
			r.Get("/players", playerH.List)
			r.Get("/player-lists", playerListH.List)

			// maintainer+ — score entry and absence reporting
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireRole("maintainer", "admin"))
				r.Put("/matches/{id}/score", scoreH.Submit)
				r.Post("/evenings/{id}/report-absent", scoreH.ReportAbsent)
				r.Get("/evenings/{id}/player-stats", eveningStatH.GetByEvening)
				r.Put("/evenings/{id}/player-stats/{playerId}", eveningStatH.Upsert)
			})

			// admin only — everything else
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireRole("admin"))

				r.Post("/import", playerH.Import)
				r.Put("/players/{id}", playerH.Update)
				r.Delete("/players/{id}", playerH.Delete)
				r.Get("/players/{id}/buddies", playerH.GetBuddies)
				r.Put("/players/{id}/buddies", playerH.SetBuddies)

				r.Post("/schedule/generate", schedH.Generate)
				r.Get("/schedules/{id}/info", schedH.GetInfo)
				r.Get("/schedules/{id}/matches", schedH.GetPlayedMatches)
				r.Post("/schedules/import-season", schedH.ImportSeason)
				r.Patch("/schedules/{id}", schedH.RenameSchedule)
				r.Delete("/schedules/{id}", schedH.DeleteSchedule)
				r.Post("/schedules/{id}/regenerate", schedH.RegenerateSchedule)
				r.Post("/schedules/{id}/active", schedH.SetActive)
				r.Post("/schedules/{id}/inhaal-avond", schedH.AddCatchUpEvening)
				r.Delete("/schedules/{id}/evenings/{eveningId}", schedH.DeleteEvening)

				r.Get("/stats", statsH.Get)
				r.Get("/stats/duties", statsH.GetDuties)
				r.Get("/stats/pdf", statsH.StandingsPDF)

				r.Get("/export/excel", exportH.Excel)
				r.Get("/export/pdf", exportH.PDF)
				r.Get("/export/evening/{id}/excel", exportH.EveningExcel)
				r.Get("/export/evening/{id}/pdf", exportH.EveningPDF)
				r.Get("/export/evening/{id}/print", exportH.EveningPrint)

				r.Get("/schedules/{id}/player-stats", seasonStatH.GetBySchedule)
				r.Put("/schedules/{id}/player-stats/{playerId}", seasonStatH.Upsert)

				r.Get("/system/logs", systemH.GetLogs)
				r.Get("/progress", progressH.GetProgress)

				// User management
				r.Get("/users", userH.List)
				r.Post("/users", userH.Create)
				r.Put("/users/{id}", userH.Update)
				r.Delete("/users/{id}", userH.Delete)
			})
		})
	})

	// SPA fallback: serve Angular app for all non-API routes.
	r.Handle("/*", web.SPAHandler())

	return r
}
