package main

import (
	"context"
	"database/sql"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"DartScheduler/domain"
	apphttp "DartScheduler/infra/http"
	"DartScheduler/infra/http/handler"
	"DartScheduler/infra/logbuf"
	"DartScheduler/infra/postgres"
	"DartScheduler/infra/sqlite"
	"DartScheduler/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := loadConfig()

	logBuf := logbuf.New(200)
	log.SetOutput(io.MultiWriter(os.Stderr, logBuf))
	log.Printf("[INFO] config: port=%s db_type=%s db_path=%s club=%q title=%q logo=%q cors=%q",
		cfg.Port, cfg.DatabaseType, cfg.DatabasePath, cfg.ClubName, cfg.AppTitle, cfg.LogoPath, cfg.AllowedOrigin)

	ctx := context.Background()

	// Repositories — wired based on DATABASE_TYPE / DATABASE_URL.
	var (
		playerRepo     domain.PlayerRepository
		scheduleRepo   domain.ScheduleRepository
		eveningRepo    domain.EveningRepository
		matchRepo      domain.MatchRepository
		eveningStatRepo domain.EveningPlayerStatRepository
		seasonStatRepo  domain.SeasonPlayerStatRepository
		playerListRepo  domain.PlayerListRepository

		// held for explicit Close on shutdown
		sqliteConn *sql.DB
		pgPool     *pgxpool.Pool
	)

	switch cfg.DatabaseType {
	case "postgres":
		if cfg.DatabaseURL == "" {
			log.Fatal("DATABASE_URL or POSTGRES_DSN must be set when DATABASE_TYPE=postgres")
		}
		pool, err := postgres.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("open postgres: %v", err)
		}
		pgPool = pool
		playerRepo = postgres.NewPlayerRepo(pool)
		scheduleRepo = postgres.NewScheduleRepo(pool)
		eveningRepo = postgres.NewEveningRepo(pool)
		matchRepo = postgres.NewMatchRepo(pool)
		eveningStatRepo = postgres.NewEveningPlayerStatRepo(pool)
		seasonStatRepo = postgres.NewSeasonPlayerStatRepo(pool)
		playerListRepo = postgres.NewPlayerListRepo(pool)

	default: // "sqlite" or unrecognised value
		db, err := sqlite.Open(cfg.DatabasePath)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		sqliteConn = db
		playerRepo = sqlite.NewPlayerRepo(db)
		scheduleRepo = sqlite.NewScheduleRepo(db)
		eveningRepo = sqlite.NewEveningRepo(db)
		matchRepo = sqlite.NewMatchRepo(db)
		eveningStatRepo = sqlite.NewEveningPlayerStatRepo(db)
		seasonStatRepo = sqlite.NewSeasonPlayerStatRepo(db)
		playerListRepo = sqlite.NewPlayerListRepo(db)
	}

	defer func() {
		if sqliteConn != nil {
			sqliteConn.Close()
		}
		if pgPool != nil {
			pgPool.Close()
		}
	}()

	// Use cases
	playerUC := usecase.NewPlayerUseCase(playerRepo, matchRepo, playerListRepo)
	scheduleUC := usecase.NewScheduleUseCase(playerRepo, scheduleRepo, eveningRepo, matchRepo)
	scoreUC := usecase.NewScoreUseCase(matchRepo, eveningRepo, seasonStatRepo)
	exportUC := usecase.NewExportUseCase(scheduleRepo, eveningRepo, matchRepo, playerRepo)

	// Log database summary at startup
	if players, err := playerUC.ListPlayers(ctx); err == nil {
		schedules, _ := scheduleUC.ListSchedules(ctx)
		log.Printf("[INFO] database: %d spelers, %d seizoenen", len(players), len(schedules))
	}

	// Handlers
	playerH := handler.NewPlayerHandler(playerUC)
	schedH := handler.NewScheduleHandler(scheduleUC)
	scoreH := handler.NewScoreHandler(scoreUC)
	statsH := handler.NewStatsHandler(playerRepo, scheduleRepo, scoreUC)
	exportH := handler.NewExportHandler(exportUC, cfg.ClubName, cfg.LogoPath)
	systemH := handler.NewSystemHandler(logBuf)
	eveningStatH := handler.NewEveningStatHandler(eveningStatRepo)
	seasonStatH := handler.NewSeasonStatHandler(seasonStatRepo)
	configH := handler.NewConfigHandler(cfg.AppTitle, cfg.ClubName, cfg.PrimaryColor)
	progressH := handler.NewProgressHandler()
	playerListH := handler.NewPlayerListHandler(playerUC)

	router := apphttp.NewRouter(playerH, schedH, scoreH, statsH, exportH, systemH, eveningStatH, seasonStatH, configH, progressH, playerListH, cfg.AllowedOrigin)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("[INFO] listening on :%s", cfg.Port)
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

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
