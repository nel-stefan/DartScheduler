package postgres

import (
	"context"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlayerRepo implements domain.PlayerRepository using PostgreSQL.
type PlayerRepo struct{ pool *pgxpool.Pool }

func NewPlayerRepo(pool *pgxpool.Pool) *PlayerRepo { return &PlayerRepo{pool: pool} }

func (r *PlayerRepo) Save(ctx context.Context, p domain.Player) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO players(id, schedule_id, nr, name, email, sponsor, address, postal_code, city, phone, mobile, member_since, class)
		 VALUES($1, NULL, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		 ON CONFLICT(id) DO UPDATE SET
		   nr           = EXCLUDED.nr,
		   name         = EXCLUDED.name,
		   email        = EXCLUDED.email,
		   sponsor      = EXCLUDED.sponsor,
		   address      = EXCLUDED.address,
		   postal_code  = EXCLUDED.postal_code,
		   city         = EXCLUDED.city,
		   phone        = EXCLUDED.phone,
		   mobile       = EXCLUDED.mobile,
		   member_since = EXCLUDED.member_since,
		   class        = EXCLUDED.class`,
		p.ID, p.Nr, p.Name, p.Email, p.Sponsor, p.Address, p.PostalCode, p.City, p.Phone, p.Mobile, p.MemberSince, p.Class)
	return err
}

func (r *PlayerRepo) SaveBatch(ctx context.Context, players []domain.Player) error {
	if len(players) == 0 {
		return nil
	}

	// Build nr → existing UUID map scoped to the same list so we preserve IDs.
	nrToID := make(map[string]uuid.UUID)
	var rows pgx.Rows
	var err error
	if players[0].ListID != nil {
		rows, err = r.pool.Query(ctx,
			`SELECT id, nr FROM players WHERE nr != '' AND list_id = $1`,
			players[0].ListID)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, nr FROM players WHERE nr != '' AND list_id IS NULL`)
	}
	if err != nil {
		return err
	}
	for rows.Next() {
		var id uuid.UUID
		var nr string
		if err := rows.Scan(&id, &nr); err != nil {
			rows.Close()
			return err
		}
		nrToID[nr] = id
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, p := range players {
		id := p.ID
		if p.Nr != "" {
			if existing, ok := nrToID[p.Nr]; ok {
				id = existing // reuse existing UUID to preserve match references
			}
		}
		var listID *uuid.UUID
		if p.ListID != nil {
			v := *p.ListID
			listID = &v
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO players(id, nr, name, email, sponsor, address, postal_code, city, phone, mobile, member_since, class, list_id)
			 VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			 ON CONFLICT(id) DO UPDATE SET
			   nr           = EXCLUDED.nr,
			   name         = EXCLUDED.name,
			   email        = EXCLUDED.email,
			   sponsor      = EXCLUDED.sponsor,
			   address      = EXCLUDED.address,
			   postal_code  = EXCLUDED.postal_code,
			   city         = EXCLUDED.city,
			   phone        = EXCLUDED.phone,
			   mobile       = EXCLUDED.mobile,
			   member_since = EXCLUDED.member_since,
			   class        = EXCLUDED.class,
			   list_id      = EXCLUDED.list_id`,
			id, p.Nr, p.Name, p.Email, p.Sponsor,
			p.Address, p.PostalCode, p.City, p.Phone, p.Mobile, p.MemberSince, p.Class, listID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PlayerRepo) FindByID(ctx context.Context, id domain.PlayerID) (domain.Player, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, nr, name, email, sponsor, address, postal_code, city, phone, mobile, member_since, class, list_id
		 FROM players WHERE id = $1`, id)
	return scanPlayer(row)
}

func (r *PlayerRepo) FindAll(ctx context.Context) ([]domain.Player, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, nr, name, email, sponsor, address, postal_code, city, phone, mobile, member_since, class, list_id
		 FROM players
		 ORDER BY
		   CASE WHEN nr ~ '^[0-9]+$' THEN nr::INTEGER ELSE NULL END NULLS LAST,
		   nr, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func (r *PlayerRepo) Delete(ctx context.Context, id domain.PlayerID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM players WHERE id = $1`, id)
	return err
}

func (r *PlayerRepo) DeleteAll(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM players`)
	return err
}

func (r *PlayerRepo) FindByList(ctx context.Context, listID domain.PlayerListID) ([]domain.Player, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, nr, name, email, sponsor, address, postal_code, city, phone, mobile, member_since, class, list_id
		 FROM players WHERE list_id = $1`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func (r *PlayerRepo) SaveBuddyPreference(ctx context.Context, bp domain.BuddyPreference) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO buddy_preferences(player_id, buddy_id) VALUES($1, $2)
		 ON CONFLICT DO NOTHING`,
		bp.PlayerID, bp.BuddyID)
	return err
}

func (r *PlayerRepo) DeleteBuddiesForPlayer(ctx context.Context, id domain.PlayerID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM buddy_preferences WHERE player_id = $1`, id)
	return err
}

func (r *PlayerRepo) DeleteSpecificBuddyPair(ctx context.Context, playerID, buddyID domain.PlayerID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM buddy_preferences WHERE player_id = $1 AND buddy_id = $2`,
		playerID, buddyID)
	return err
}

func (r *PlayerRepo) DeleteAllBuddyPairs(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM buddy_preferences`)
	return err
}

func (r *PlayerRepo) FindBuddiesForPlayer(ctx context.Context, id domain.PlayerID) ([]domain.PlayerID, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT buddy_id FROM buddy_preferences WHERE player_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.PlayerID, 0)
	for rows.Next() {
		var buddyID uuid.UUID
		if err := rows.Scan(&buddyID); err != nil {
			return nil, fmt.Errorf("scan buddy_id: %w", err)
		}
		out = append(out, buddyID)
	}
	return out, rows.Err()
}

func (r *PlayerRepo) FindAllBuddyPairs(ctx context.Context) ([]domain.BuddyPreference, error) {
	rows, err := r.pool.Query(ctx, `SELECT player_id, buddy_id FROM buddy_preferences`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.BuddyPreference, 0)
	for rows.Next() {
		var pid, bid uuid.UUID
		if err := rows.Scan(&pid, &bid); err != nil {
			return nil, err
		}
		out = append(out, domain.BuddyPreference{PlayerID: pid, BuddyID: bid})
	}
	return out, rows.Err()
}

func scanPlayer(s pgxScanner) (domain.Player, error) {
	var p domain.Player
	var listID *uuid.UUID
	if err := s.Scan(&p.ID, &p.Nr, &p.Name, &p.Email, &p.Sponsor,
		&p.Address, &p.PostalCode, &p.City, &p.Phone, &p.Mobile,
		&p.MemberSince, &p.Class, &listID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, domain.ErrNotFound
		}
		return p, fmt.Errorf("scan player: %w", err)
	}
	p.ListID = listID
	return p, nil
}

func scanPlayers(rows pgx.Rows) ([]domain.Player, error) {
	var players []domain.Player
	for rows.Next() {
		var p domain.Player
		var listID *uuid.UUID
		if err := rows.Scan(&p.ID, &p.Nr, &p.Name, &p.Email, &p.Sponsor,
			&p.Address, &p.PostalCode, &p.City, &p.Phone, &p.Mobile,
			&p.MemberSince, &p.Class, &listID); err != nil {
			return nil, err
		}
		p.ListID = listID
		players = append(players, p)
	}
	return players, rows.Err()
}
