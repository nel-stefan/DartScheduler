package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type PlayerRepo struct{ db *sql.DB }

func NewPlayerRepo(db *sql.DB) *PlayerRepo { return &PlayerRepo{db: db} }

func (r *PlayerRepo) Save(ctx context.Context, p domain.Player) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO players(id,schedule_id,nr,name,email,sponsor,address,postal_code,city,phone,mobile,member_since,class)
         VALUES(?,NULL,?,?,?,?,?,?,?,?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET nr=excluded.nr, name=excluded.name, email=excluded.email, sponsor=excluded.sponsor,
           address=excluded.address, postal_code=excluded.postal_code, city=excluded.city,
           phone=excluded.phone, mobile=excluded.mobile, member_since=excluded.member_since,
           class=excluded.class`,
		p.ID.String(), p.Nr, p.Name, p.Email, p.Sponsor, p.Address, p.PostalCode, p.City, p.Phone, p.Mobile, p.MemberSince, p.Class)
	return err
}

func (r *PlayerRepo) SaveBatch(ctx context.Context, players []domain.Player) error {
	// Build nr → existing UUID map so we preserve IDs (match references stay valid).
	// Scope the lookup to the same list so players from different lists never share UUIDs.
	var existingRows *sql.Rows
	var err error
	if len(players) > 0 && players[0].ListID != nil {
		existingRows, err = r.db.QueryContext(ctx,
			`SELECT id, nr FROM players WHERE nr != '' AND list_id = ?`,
			players[0].ListID.String())
	} else {
		existingRows, err = r.db.QueryContext(ctx,
			`SELECT id, nr FROM players WHERE nr != '' AND list_id IS NULL`)
	}
	if err != nil {
		return err
	}
	nrToID := make(map[string]string)
	for existingRows.Next() {
		var id, nr string
		if err := existingRows.Scan(&id, &nr); err != nil {
			existingRows.Close()
			return err
		}
		nrToID[nr] = id
	}
	existingRows.Close()
	if err := existingRows.Err(); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO players(id,nr,name,email,sponsor,address,postal_code,city,phone,mobile,member_since,class,list_id)
         VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET
           nr=excluded.nr, name=excluded.name, email=excluded.email, sponsor=excluded.sponsor,
           address=excluded.address, postal_code=excluded.postal_code, city=excluded.city,
           phone=excluded.phone, mobile=excluded.mobile, member_since=excluded.member_since,
           class=excluded.class, list_id=excluded.list_id`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range players {
		id := p.ID.String()
		if p.Nr != "" {
			if existing, ok := nrToID[p.Nr]; ok {
				id = existing // reuse existing UUID to preserve match references
			}
		}
		var listIDStr *string
		if p.ListID != nil {
			s := p.ListID.String()
			listIDStr = &s
		}
		if _, err := stmt.ExecContext(ctx, id, p.Nr, p.Name, p.Email, p.Sponsor,
			p.Address, p.PostalCode, p.City, p.Phone, p.Mobile, p.MemberSince, p.Class, listIDStr); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PlayerRepo) FindByID(ctx context.Context, id domain.PlayerID) (domain.Player, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,nr,name,email,sponsor,address,postal_code,city,phone,mobile,member_since,class,list_id FROM players WHERE id=?`, id.String())
	return scanPlayer(row)
}

func (r *PlayerRepo) FindAll(ctx context.Context) ([]domain.Player, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,nr,name,email,sponsor,address,postal_code,city,phone,mobile,member_since,class,list_id FROM players ORDER BY CAST(nr AS INTEGER), nr, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

func (r *PlayerRepo) Delete(ctx context.Context, id domain.PlayerID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM players WHERE id=?`, id.String())
	return err
}

func (r *PlayerRepo) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM players`)
	return err
}

func (r *PlayerRepo) SaveBuddyPreference(ctx context.Context, bp domain.BuddyPreference) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO buddy_preferences(player_id,buddy_id) VALUES(?,?)`,
		bp.PlayerID.String(), bp.BuddyID.String())
	return err
}

func (r *PlayerRepo) DeleteBuddiesForPlayer(ctx context.Context, id domain.PlayerID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM buddy_preferences WHERE player_id = ?`, id.String())
	return err
}

func (r *PlayerRepo) DeleteAllBuddyPairs(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM buddy_preferences`)
	return err
}

func (r *PlayerRepo) FindBuddiesForPlayer(ctx context.Context, id domain.PlayerID) ([]domain.PlayerID, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT buddy_id FROM buddy_preferences WHERE player_id=?`, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.PlayerID, 0)
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		uid, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid buddy_id %q: %w", s, err)
		}
		out = append(out, uid)
	}
	return out, rows.Err()
}

func (r *PlayerRepo) FindAllBuddyPairs(ctx context.Context) ([]domain.BuddyPreference, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT player_id, buddy_id FROM buddy_preferences`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.BuddyPreference, 0)
	for rows.Next() {
		var ps, bs string
		if err := rows.Scan(&ps, &bs); err != nil {
			return nil, err
		}
		pid, err := uuid.Parse(ps)
		if err != nil {
			return nil, err
		}
		bid, err := uuid.Parse(bs)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.BuddyPreference{PlayerID: pid, BuddyID: bid})
	}
	return out, rows.Err()
}

func (r *PlayerRepo) FindByList(ctx context.Context, listID domain.PlayerListID) ([]domain.Player, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,nr,name,email,sponsor,address,postal_code,city,phone,mobile,member_since,class,list_id
         FROM players WHERE list_id = ?`, listID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlayers(rows)
}

// scanPlayer works for both *sql.Row and *sql.Rows via a shared interface.
type scanner interface {
	Scan(dest ...any) error
}

func scanPlayer(s scanner) (domain.Player, error) {
	var p domain.Player
	var idStr string
	var listIDStr *string
	if err := s.Scan(&idStr, &p.Nr, &p.Name, &p.Email, &p.Sponsor, &p.Address, &p.PostalCode, &p.City, &p.Phone, &p.Mobile, &p.MemberSince, &p.Class, &listIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, domain.ErrNotFound
		}
		return p, err
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		return p, fmt.Errorf("invalid player id %q: %w", idStr, err)
	}
	p.ID = uid
	if listIDStr != nil {
		id, _ := uuid.Parse(*listIDStr)
		p.ListID = &id
	}
	return p, nil
}

func scanPlayers(rows *sql.Rows) ([]domain.Player, error) {
	var players []domain.Player
	for rows.Next() {
		var p domain.Player
		var idStr string
		var listIDStr *string
		if err := rows.Scan(&idStr, &p.Nr, &p.Name, &p.Email, &p.Sponsor,
			&p.Address, &p.PostalCode, &p.City, &p.Phone, &p.Mobile,
			&p.MemberSince, &p.Class, &listIDStr); err != nil {
			return nil, err
		}
		p.ID, _ = uuid.Parse(idStr)
		if listIDStr != nil {
			id, _ := uuid.Parse(*listIDStr)
			p.ListID = &id
		}
		players = append(players, p)
	}
	return players, rows.Err()
}
