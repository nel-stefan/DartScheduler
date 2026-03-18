package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apphttp "DartScheduler/infra/http"
	"DartScheduler/infra/http/handler"
	"DartScheduler/infra/logbuf"
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

	logBuf := logbuf.New(200)
	log.SetOutput(io.MultiWriter(os.Stderr, logBuf))

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
	exportH  := handler.NewExportHandler(exportUC)
	systemH  := handler.NewSystemHandler(logBuf)

	router := apphttp.NewRouter(playerH, schedH, scoreH, statsH, exportH, systemH)

	srv := &http.Server{Addr: ":" + port, Handler: router}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-serveErr:
		log.Fatalf("listen: %v", err)
	case <-quit:
	}
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
