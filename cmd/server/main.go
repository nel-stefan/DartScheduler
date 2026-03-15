package main

import (
	"log"
	"net/http"
	"os"

	apphttp "DartScheduler/infra/http"
	"DartScheduler/infra/http/handler"
	"DartScheduler/infra/sqlite"
	"DartScheduler/usecase"
)

func main() {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "dartscheduler.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Repositories
	playerRepo := sqlite.NewPlayerRepo(db)
	scheduleRepo := sqlite.NewScheduleRepo(db)
	eveningRepo := sqlite.NewEveningRepo(db)
	matchRepo := sqlite.NewMatchRepo(db)

	// Use cases
	playerUC := usecase.NewPlayerUseCase(playerRepo, matchRepo)
	scheduleUC := usecase.NewScheduleUseCase(playerRepo, scheduleRepo, eveningRepo, matchRepo)
	scoreUC := usecase.NewScoreUseCase(matchRepo)
	exportUC := usecase.NewExportUseCase(scheduleRepo, eveningRepo, matchRepo, playerRepo)

	// Handlers
	playerH := handler.NewPlayerHandler(playerUC)
	schedH := handler.NewScheduleHandler(scheduleUC)
	scoreH := handler.NewScoreHandler(scoreUC)
	statsH := handler.NewStatsHandler(playerRepo, scoreUC)
	exportH := handler.NewExportHandler(exportUC)

	router := apphttp.NewRouter(playerH, schedH, scoreH, statsH, exportH)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("server: %v", err)
	}
}
