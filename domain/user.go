package domain

import (
	"context"
	"time"
)

// Role constants for the three access levels.
const (
	RoleViewer     = "viewer"
	RoleMaintainer = "maintainer"
	RoleAdmin      = "admin"
)

// User represents an application user with authentication credentials.
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

// UserRepository defines persistence operations for User entities.
type UserRepository interface {
	Create(ctx context.Context, user User) error
	FindByUsername(ctx context.Context, username string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	List(ctx context.Context) ([]User, error)
	UpdateRole(ctx context.Context, id, role string) error
	UpdatePassword(ctx context.Context, id, hash string) error
	Delete(ctx context.Context, id string) error
	ExistsAdmin(ctx context.Context) (bool, error)
}
