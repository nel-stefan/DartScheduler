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

		r.Post("/schedule/generate", schedH.Generate)
		r.Get("/schedule", schedH.Get)
		r.Get("/schedule/evening/{id}", schedH.GetEvening)

		r.Put("/matches/{id}/score", scoreH.Submit)

		r.Get("/stats", statsH.Get)

		r.Get("/export/excel", exportH.Excel)
		r.Get("/export/pdf", exportH.PDF)
	})

	// SPA fallback: serve Angular app for all non-API routes.
	r.Handle("/*", web.SPAHandler())

	return r
}
