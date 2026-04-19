package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"DartScheduler/domain"
)

// UserRepo implements domain.UserRepository using SQLite.
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u domain.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users(id, username, password_hash, role, created_at) VALUES(?,?,?,?,?)`,
		u.ID, u.Username, u.PasswordHash, u.Role, u.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (domain.User, error) {
	var u domain.User
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`,
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &createdAt)
	if err == sql.ErrNoRows {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	var createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &createdAt)
	if err == sql.ErrNoRows {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, password_hash, role, created_at FROM users ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []domain.User
	for rows.Next() {
		var u domain.User
		var createdAt string
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &createdAt); err != nil {
			return nil, err
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) UpdateRole(ctx context.Context, id, role string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET role = ? WHERE id = ?`, role, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *UserRepo) ExistsAdmin(ctx context.Context) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	return count > 0, err
}
