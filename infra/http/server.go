// Package http registers all API routes and mounts the Angular SPA handler.
//
// Route overview:
//
//	POST   /api/import                        — import players from Excel
//	GET    /api/players                       — list all players
//	POST   /api/schedule/generate             — generate a new schedule
//	GET    /api/schedule                      — get the current schedule
//	GET    /api/schedule/evening/{id}         — get a single playing evening
//	PUT    /api/matches/{id}/score            — submit a match score
//	GET    /api/stats                         — get standings
//	GET    /api/stats/duties                  — get secretary/counter duty statistics
//	GET    /api/export/excel                  — download schedule as Excel
//	GET    /api/export/pdf                    — download schedule as PDF
//	GET    /api/export/evening/{id}/excel     — download a single evening as match form Excel
//	/*                                        — Angular SPA (fallback)
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
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(mw.Logger)
	r.Use(mw.CORS)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Post("/import", playerH.Import)
		r.Get("/players", playerH.List)
		r.Put("/players/{id}", playerH.Update)
		r.Delete("/players/{id}", playerH.Delete)
		r.Get("/players/{id}/buddies", playerH.GetBuddies)
		r.Put("/players/{id}/buddies", playerH.SetBuddies)

		r.Post("/schedule/generate", schedH.Generate)
		r.Get("/schedule", schedH.Get)
		r.Get("/schedule/evening/{id}", schedH.GetEvening)
		r.Get("/schedules", schedH.List)
		r.Get("/schedules/{id}", schedH.GetByID)
		r.Get("/schedules/{id}/info", schedH.GetInfo)
		r.Post("/schedules/import-season", schedH.ImportSeason)
		r.Delete("/schedules/{id}", schedH.DeleteSchedule)
		r.Post("/schedules/{id}/inhaal-avond", schedH.AddCatchUpEvening)
		r.Delete("/schedules/{id}/evenings/{eveningId}", schedH.DeleteEvening)

		r.Put("/matches/{id}/score", scoreH.Submit)
		r.Post("/evenings/{id}/report-absent", scoreH.ReportAbsent)

		r.Get("/stats", statsH.Get)
		r.Get("/stats/duties", statsH.GetDuties)

		r.Get("/export/excel", exportH.Excel)
		r.Get("/export/pdf", exportH.PDF)
		r.Get("/export/evening/{id}/excel", exportH.EveningExcel)
		r.Get("/export/evening/{id}/print", exportH.EveningPrint)

		r.Get("/system/logs", systemH.GetLogs)
	})

	// SPA fallback: serve Angular app for all non-API routes.
	r.Handle("/*", web.SPAHandler())

	return r
}
