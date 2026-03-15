// Package http registreert alle API-routes en monteert de Angular SPA-handler.
//
// Routeoverzicht:
//
//	POST   /api/import                  — spelers importeren vanuit Excel
//	GET    /api/players                 — alle spelers ophalen
//	POST   /api/schedule/generate       — nieuw schema genereren
//	GET    /api/schedule                — huidig schema ophalen
//	GET    /api/schedule/evening/{id}   — één speelavond ophalen
//	PUT    /api/matches/{id}/score      — score invoeren
//	GET    /api/stats                   — ranglijst ophalen
//	GET    /api/stats/duties              — schrijver/teller statistieken
//	GET    /api/export/excel                    — schema als Excel downloaden
//	GET    /api/export/pdf                      — schema als PDF downloaden
//	GET    /api/export/evening/{id}/excel       — één avond als wedstrijdformulier Excel
//	/*                                          — Angular SPA (fallback)
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
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(mw.Logger)
	r.Use(mw.CORS)

	r.Route("/api", func(r chi.Router) {
		r.Post("/import", playerH.Import)
		r.Get("/players", playerH.List)
		r.Put("/players/{id}", playerH.Update)
		r.Get("/players/{id}/buddies", playerH.GetBuddies)
		r.Put("/players/{id}/buddies", playerH.SetBuddies)

		r.Post("/schedule/generate", schedH.Generate)
		r.Get("/schedule", schedH.Get)
		r.Get("/schedule/evening/{id}", schedH.GetEvening)
		r.Get("/schedules", schedH.List)
		r.Get("/schedules/{id}", schedH.GetByID)
		r.Post("/schedules/import-season", schedH.ImportSeason)

		r.Put("/matches/{id}/score", scoreH.Submit)

		r.Get("/stats", statsH.Get)
		r.Get("/stats/duties", statsH.GetDuties)

		r.Get("/export/excel", exportH.Excel)
		r.Get("/export/pdf", exportH.PDF)
		r.Get("/export/evening/{id}/excel", exportH.EveningExcel)
	})

	// SPA fallback: serve Angular app for all non-API routes.
	r.Handle("/*", web.SPAHandler())

	return r
}
