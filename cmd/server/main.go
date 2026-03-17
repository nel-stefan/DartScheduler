package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{Addr: ":" + port, Handler: router}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
