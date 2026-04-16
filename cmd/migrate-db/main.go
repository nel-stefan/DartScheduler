// Command migrate-db migrates data from a SQLite database to a PostgreSQL database.
//
// Usage:
//
//	migrate-db --sqlite-path ./dartscheduler.db \
//	           --postgres-dsn "postgres://dart:secret@localhost:5432/dartscheduler?sslmode=disable"
//
// The command is idempotent: running it multiple times is safe because every
// INSERT uses ON CONFLICT DO NOTHING.  Rows already present in PostgreSQL are
// skipped without error.
//
// Migration order (respects FK constraints):
//  1. applied_migrations
//  2. player_lists
//  3. schedules
//  4. evenings
//  5. players
//  6. buddy_preferences
//  7. matches
//  8. season_player_stats
//  9. evening_player_stats
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"DartScheduler/infra/postgres"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func main() {
	sqlitePath := flag.String("sqlite-path", "dartscheduler.db", "Path to the SQLite database file")
	postgresDSN := flag.String("postgres-dsn", "", "PostgreSQL connection string (DSN)")
	flag.Parse()

	if *postgresDSN == "" {
		log.Fatal("--postgres-dsn is required")
	}

	ctx := context.Background()

	log.Printf("Opening SQLite: %s", *sqlitePath)
	src, err := sql.Open("sqlite", *sqlitePath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer src.Close()
	src.SetMaxOpenConns(1)

	if err := src.PingContext(ctx); err != nil {
		log.Fatalf("ping sqlite: %v", err)
	}

	log.Printf("Connecting to PostgreSQL and running schema migration...")
	// Run schema migration via pool first.
	pool, err := postgres.Open(ctx, *postgresDSN)
	if err != nil {
		log.Fatalf("postgres schema migration: %v", err)
	}
	pool.Close()

	// Open a database/sql handle backed by pgx for row-by-row inserts.
	connCfg, err := pgx.ParseConfig(*postgresDSN)
	if err != nil {
		log.Fatalf("parse postgres DSN: %v", err)
	}
	dst := stdlib.OpenDB(*connCfg)
	defer dst.Close()

	if err := dst.PingContext(ctx); err != nil {
		log.Fatalf("ping postgres: %v", err)
	}

	log.Println("Starting migration...")

	steps := []struct {
		name string
		fn   func(ctx context.Context, src, dst *sql.DB) (int, error)
	}{
		{"applied_migrations", migrateAppliedMigrations},
		{"player_lists", migratePlayerLists},
		{"schedules", migrateSchedules},
		{"evenings", migrateEvenings},
		{"players", migratePlayers},
		{"buddy_preferences", migrateBuddyPreferences},
		{"matches", migrateMatches},
		{"season_player_stats", migrateSeasonPlayerStats},
		{"evening_player_stats", migrateEveningPlayerStats},
	}

	total := 0
	for _, step := range steps {
		n, err := step.fn(ctx, src, dst)
		if err != nil {
			log.Fatalf("migrate %s: %v", step.name, err)
		}
		log.Printf("  %-30s %d rows", step.name, n)
		total += n
	}

	log.Printf("Migration complete — %d rows inserted into PostgreSQL.", total)
}

// --- table migrations --------------------------------------------------------

func migrateAppliedMigrations(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx, `SELECT name, applied_at FROM applied_migrations`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var name, appliedAt string
		if err := rows.Scan(&name, &appliedAt); err != nil {
			return n, err
		}
		ts, _ := parseTimestamp(appliedAt)
		res, err := dst.ExecContext(ctx,
			`INSERT INTO applied_migrations(name, applied_at) VALUES($1, $2) ON CONFLICT DO NOTHING`,
			name, ts)
		if err != nil {
			return n, fmt.Errorf("insert applied_migrations %q: %w", name, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migratePlayerLists(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx, `SELECT id, name, created_at FROM player_lists`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, name, createdAt string
		if err := rows.Scan(&id, &name, &createdAt); err != nil {
			return n, err
		}
		ts, _ := parseTimestamp(createdAt)
		res, err := dst.ExecContext(ctx,
			`INSERT INTO player_lists(id, name, created_at) VALUES($1, $2, $3) ON CONFLICT DO NOTHING`,
			id, name, ts)
		if err != nil {
			return n, fmt.Errorf("insert player_list %q: %w", id, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migrateSchedules(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx,
		`SELECT id, competition_name, season, active, created_at, player_list_id FROM schedules`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, name, season, createdAt string
		var active int
		var listID *string
		if err := rows.Scan(&id, &name, &season, &active, &createdAt, &listID); err != nil {
			return n, err
		}
		ts, _ := parseTimestamp(createdAt)
		res, err := dst.ExecContext(ctx,
			`INSERT INTO schedules(id, competition_name, season, active, created_at, player_list_id)
			 VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`,
			id, name, season, active != 0, ts, listID)
		if err != nil {
			return n, fmt.Errorf("insert schedule %q: %w", id, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migrateEvenings(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx,
		`SELECT id, schedule_id, number, date, is_inhaal_avond FROM evenings`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, scheduleID, date string
		var number, isInhaal int
		if err := rows.Scan(&id, &scheduleID, &number, &date, &isInhaal); err != nil {
			return n, err
		}
		ts, _ := parseTimestamp(date)
		res, err := dst.ExecContext(ctx,
			`INSERT INTO evenings(id, schedule_id, number, date, is_inhaal_avond)
			 VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`,
			id, scheduleID, number, ts, isInhaal != 0)
		if err != nil {
			return n, fmt.Errorf("insert evening %q: %w", id, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migratePlayers(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx,
		`SELECT id, schedule_id, list_id, nr, name, email, sponsor,
		        address, postal_code, city, phone, mobile, member_since, class
		 FROM players`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id, nr, name, email, sponsor, address, postalCode, city, phone, mobile, memberSince, class string
		var scheduleID, listID *string
		if err := rows.Scan(&id, &scheduleID, &listID, &nr, &name, &email, &sponsor,
			&address, &postalCode, &city, &phone, &mobile, &memberSince, &class); err != nil {
			return n, err
		}
		res, err := dst.ExecContext(ctx,
			`INSERT INTO players(id, schedule_id, list_id, nr, name, email, sponsor,
			                     address, postal_code, city, phone, mobile, member_since, class)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) ON CONFLICT DO NOTHING`,
			id, scheduleID, listID, nr, name, email, sponsor,
			address, postalCode, city, phone, mobile, memberSince, class)
		if err != nil {
			return n, fmt.Errorf("insert player %q: %w", id, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migrateBuddyPreferences(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx, `SELECT player_id, buddy_id FROM buddy_preferences`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var playerID, buddyID string
		if err := rows.Scan(&playerID, &buddyID); err != nil {
			return n, err
		}
		res, err := dst.ExecContext(ctx,
			`INSERT INTO buddy_preferences(player_id, buddy_id) VALUES($1, $2) ON CONFLICT DO NOTHING`,
			playerID, buddyID)
		if err != nil {
			return n, fmt.Errorf("insert buddy_preference (%s,%s): %w", playerID, buddyID, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migrateMatches(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx,
		`SELECT id, evening_id, player_a, player_b, score_a, score_b, played,
		        leg1_winner, leg1_turns, leg2_winner, leg2_turns,
		        leg3_winner, leg3_turns,
		        reported_by, reschedule_date, secretary_nr, counter_nr,
		        player_a_180s, player_b_180s, player_a_highest_finish, player_b_highest_finish,
		        played_date
		 FROM matches`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var (
			id, eveningID, playerA, playerB                    string
			scoreA, scoreB                                     *int
			played, leg1Turns, leg2Turns, leg3Turns            int
			leg1Winner, leg2Winner, leg3Winner                 string
			reportedBy, rescheduleDate, secretaryNr, counterNr string
			pA180s, pB180s, pAHighest, pBHighest               int
			playedDate                                         string
		)
		if err := rows.Scan(
			&id, &eveningID, &playerA, &playerB, &scoreA, &scoreB, &played,
			&leg1Winner, &leg1Turns, &leg2Winner, &leg2Turns,
			&leg3Winner, &leg3Turns,
			&reportedBy, &rescheduleDate, &secretaryNr, &counterNr,
			&pA180s, &pB180s, &pAHighest, &pBHighest,
			&playedDate,
		); err != nil {
			return n, err
		}
		res, err := dst.ExecContext(ctx,
			`INSERT INTO matches(
				id, evening_id, player_a, player_b, score_a, score_b, played,
				leg1_winner, leg1_turns, leg2_winner, leg2_turns,
				leg3_winner, leg3_turns,
				reported_by, reschedule_date, secretary_nr, counter_nr,
				player_a_180s, player_b_180s, player_a_highest_finish, player_b_highest_finish,
				played_date)
			 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
			 ON CONFLICT DO NOTHING`,
			id, eveningID, playerA, playerB, scoreA, scoreB, played != 0,
			leg1Winner, leg1Turns, leg2Winner, leg2Turns,
			leg3Winner, leg3Turns,
			reportedBy, rescheduleDate, secretaryNr, counterNr,
			pA180s, pB180s, pAHighest, pBHighest,
			playedDate)
		if err != nil {
			return n, fmt.Errorf("insert match %q: %w", id, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migrateSeasonPlayerStats(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx,
		`SELECT schedule_id, player_id, one_eighties, highest_finish FROM season_player_stats`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var scheduleID, playerID string
		var oneEighties, highestFinish int
		if err := rows.Scan(&scheduleID, &playerID, &oneEighties, &highestFinish); err != nil {
			return n, err
		}
		res, err := dst.ExecContext(ctx,
			`INSERT INTO season_player_stats(schedule_id, player_id, one_eighties, highest_finish)
			 VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
			scheduleID, playerID, oneEighties, highestFinish)
		if err != nil {
			return n, fmt.Errorf("insert season_player_stat (%s,%s): %w", scheduleID, playerID, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

func migrateEveningPlayerStats(ctx context.Context, src, dst *sql.DB) (int, error) {
	rows, err := src.QueryContext(ctx,
		`SELECT evening_id, player_id, one_eighties, highest_finish FROM evening_player_stats`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var eveningID, playerID string
		var oneEighties, highestFinish int
		if err := rows.Scan(&eveningID, &playerID, &oneEighties, &highestFinish); err != nil {
			return n, err
		}
		res, err := dst.ExecContext(ctx,
			`INSERT INTO evening_player_stats(evening_id, player_id, one_eighties, highest_finish)
			 VALUES($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
			eveningID, playerID, oneEighties, highestFinish)
		if err != nil {
			return n, fmt.Errorf("insert evening_player_stat (%s,%s): %w", eveningID, playerID, err)
		}
		ra, _ := res.RowsAffected()
		n += int(ra)
	}
	return n, rows.Err()
}

// parseTimestamp tries several SQLite timestamp formats and returns a UTC time.
func parseTimestamp(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999Z07:00",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Now().UTC(), fmt.Errorf("unrecognised timestamp %q", s)
}
