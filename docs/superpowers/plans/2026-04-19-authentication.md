# Authentication & RBAC Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add JWT-based authentication with three roles (viewer/maintainer/admin), network trust bypass for 192.168.x.x, and an in-app user management page to DartScheduler.

**Architecture:** Go backend gains a `users` SQLite table, JWT middleware, and new `/api/auth/` + `/api/users/` endpoints grouped by role. Angular frontend gains an AuthService, route guards, a login page, and a gebruikers management page. Network requests from 192.168.0.0/16 automatically receive admin identity without a token.

**Tech Stack:** Go (`golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`), Angular 19 (signals, functional guards, `APP_INITIALIZER`), Angular Material.

---

## File Map

### New Go files
- `domain/user.go` — User entity, role constants, UserRepository interface
- `infra/sqlite/user_repo.go` — SQLite UserRepository
- `usecase/auth_usecase.go` — Login, user CRUD, SeedAdmin
- `infra/http/middleware/auth.go` — Auth + RequireRole middleware, ResolveIdentity
- `infra/http/middleware/auth_test.go` — Middleware tests
- `infra/http/handler/auth_handler.go` — POST /api/auth/login, GET /api/auth/me
- `infra/http/handler/user_handler.go` — User CRUD handlers

### Modified Go files
- `domain/errors.go` — add ErrInvalidCredentials
- `infra/sqlite/schema.sql` — add users table
- `infra/sqlite/db.go` — add users table alteration
- `usecase/dto.go` — add auth DTOs (LoginOutput, CreateUserInput, UpdateUserInput, UserDTO)
- `usecase/auth_usecase_test.go` — New file with auth usecase tests
- `infra/http/server.go` — restructure routes into role-based groups, add authH/userH
- `cmd/server/config.go` — add JWTSecret field
- `cmd/server/main.go` — wire auth components, add seed-admin subcommand

### New Angular files
- `frontend/src/app/services/auth.service.ts`
- `frontend/src/app/interceptors/auth.interceptor.ts`
- `frontend/src/app/guards/auth.guard.ts`
- `frontend/src/app/guards/role.guard.ts`
- `frontend/src/app/components/login/login.component.ts`
- `frontend/src/app/components/gebruikers/gebruikers.component.ts`

### Modified Angular files
- `frontend/src/app/app.config.ts` — add APP_INITIALIZER + withInterceptors
- `frontend/src/app/app.routes.ts` — add guards + login + gebruikers routes
- `frontend/src/app/mobile/mobile.routes.ts` — add guards
- `frontend/src/app/app.component.ts` — inject AuthService
- `frontend/src/app/app.component.html` — role-based nav + logout button
- `frontend/src/app/app.component.spec.ts` — add APP_INITIALIZER mock

---

## Task 1: Add JWT dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add the golang-jwt library**

```bash
cd /path/to/DartScheduler
go get github.com/golang-jwt/jwt/v5
```

Expected output: line added to go.mod like `github.com/golang-jwt/jwt/v5 v5.x.x`

- [ ] **Step 2: Verify go.mod**

```bash
grep "golang-jwt" go.mod
```

Expected: `github.com/golang-jwt/jwt/v5 v5.x.x`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add golang-jwt dependency"
```

---

## Task 2: Domain User entity + error sentinel

**Files:**
- Create: `domain/user.go`
- Modify: `domain/errors.go`

- [ ] **Step 1: Add ErrInvalidCredentials to domain/errors.go**

Add after the last existing `var` block in `domain/errors.go`:

```go
// ErrInvalidCredentials is returned when username or password is incorrect.
ErrInvalidCredentials = errors.New("invalid credentials")
```

The full vars block should look like:
```go
var (
    ErrNotFound                    = errors.New("not found")
    ErrAlreadyExists               = errors.New("already exists")
    ErrInvalidInput                = errors.New("invalid input")
    ErrScheduleConflict            = errors.New("schedule conflict")
    ErrMatchAlreadyPlayed          = errors.New("match already played")
    ErrScheduleConstraintViolation = errors.New("schema voldoet niet aan harde constraints")
    ErrInvalidCredentials          = errors.New("invalid credentials")
)
```

- [ ] **Step 2: Create domain/user.go**

```go
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
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./domain/...
```

Expected: no output (no errors)

- [ ] **Step 4: Commit**

```bash
git add domain/user.go domain/errors.go
git commit -m "feat: add User domain entity, UserRepository interface, ErrInvalidCredentials"
```

---

## Task 3: SQLite users table + UserRepo

**Files:**
- Modify: `infra/sqlite/schema.sql`
- Modify: `infra/sqlite/db.go`
- Create: `infra/sqlite/user_repo.go`

- [ ] **Step 1: Add users table to schema.sql**

Add after the `applied_migrations` table definition in `infra/sqlite/schema.sql`:

```sql
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK(role IN ('viewer', 'maintainer', 'admin')),
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
);
```

- [ ] **Step 2: Add alteration to db.go for existing databases**

In `infra/sqlite/db.go`, add to the `alterations` slice (after the last existing entry):

```go
`CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK(role IN ('viewer', 'maintainer', 'admin')),
    created_at    TEXT NOT NULL DEFAULT (datetime('now'))
)`,
```

- [ ] **Step 3: Create infra/sqlite/user_repo.go**

```go
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
```

- [ ] **Step 4: Verify compilation**

```bash
go build ./infra/sqlite/...
```

Expected: no output

- [ ] **Step 5: Commit**

```bash
git add infra/sqlite/schema.sql infra/sqlite/db.go infra/sqlite/user_repo.go
git commit -m "feat: add users table to SQLite schema and UserRepo implementation"
```

---

## Task 4: Auth use case + DTOs + tests

**Files:**
- Modify: `usecase/dto.go`
- Create: `usecase/auth_usecase.go`
- Create: `usecase/auth_usecase_test.go`

- [ ] **Step 1: Write the failing tests first**

Create `usecase/auth_usecase_test.go`:

```go
package usecase_test

import (
	"context"
	"errors"
	"testing"

	"DartScheduler/domain"
	"DartScheduler/usecase"
)

// stubUserRepo is an in-memory UserRepository for testing.
type stubUserRepo struct {
	users map[string]domain.User
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: make(map[string]domain.User)}
}

func (r *stubUserRepo) Create(_ context.Context, u domain.User) error {
	for _, existing := range r.users {
		if existing.Username == u.Username {
			return domain.ErrAlreadyExists
		}
	}
	r.users[u.ID] = u
	return nil
}

func (r *stubUserRepo) FindByUsername(_ context.Context, username string) (domain.User, error) {
	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrNotFound
}

func (r *stubUserRepo) FindByID(_ context.Context, id string) (domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *stubUserRepo) List(_ context.Context) ([]domain.User, error) {
	out := make([]domain.User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, u)
	}
	return out, nil
}

func (r *stubUserRepo) UpdateRole(_ context.Context, id, role string) error {
	u, ok := r.users[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.Role = role
	r.users[id] = u
	return nil
}

func (r *stubUserRepo) UpdatePassword(_ context.Context, id, hash string) error {
	u, ok := r.users[id]
	if !ok {
		return domain.ErrNotFound
	}
	u.PasswordHash = hash
	r.users[id] = u
	return nil
}

func (r *stubUserRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.users[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.users, id)
	return nil
}

func (r *stubUserRepo) ExistsAdmin(_ context.Context) (bool, error) {
	for _, u := range r.users {
		if u.Role == "admin" {
			return true, nil
		}
	}
	return false, nil
}

const testJWTSecret = "test-secret-key-not-for-production"

func TestAuthUseCase_Login_Success(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)

	if _, err := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Username: "alice", Password: "secret123", Role: "admin",
	}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	out, err := uc.Login(context.Background(), "alice", "secret123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if out.Token == "" {
		t.Error("expected non-empty token")
	}
	if out.Username != "alice" {
		t.Errorf("username: got %q, want %q", out.Username, "alice")
	}
	if out.Role != "admin" {
		t.Errorf("role: got %q, want %q", out.Role, "admin")
	}
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)
	_, _ = uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Username: "alice", Password: "secret123", Role: "viewer",
	})

	_, err := uc.Login(context.Background(), "alice", "wrongpassword")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUseCase_Login_UnknownUser(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)

	_, err := uc.Login(context.Background(), "nobody", "password")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthUseCase_CreateUser_DuplicateUsername(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)
	_, _ = uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Username: "alice", Password: "pass", Role: "viewer",
	})

	_, err := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Username: "alice", Password: "other", Role: "maintainer",
	})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAuthUseCase_CreateUser_InvalidRole(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)

	_, err := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Username: "bob", Password: "pass", Role: "superuser",
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAuthUseCase_DeleteUser_CannotDeleteSelf(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)
	dto, _ := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Username: "admin", Password: "pass", Role: "admin",
	})

	err := uc.DeleteUser(context.Background(), dto.ID, dto.ID)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestAuthUseCase_SeedAdmin_WhenEmpty(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)

	if err := uc.SeedAdmin(context.Background(), "admin", "adminpass"); err != nil {
		t.Fatalf("SeedAdmin: %v", err)
	}
	users, _ := repo.List(context.Background())
	if len(users) != 1 || users[0].Role != "admin" {
		t.Errorf("expected 1 admin user, got %+v", users)
	}
}

func TestAuthUseCase_SeedAdmin_AlreadyExists(t *testing.T) {
	repo := newStubUserRepo()
	uc := usecase.NewAuthUseCase(repo, testJWTSecret)
	_ = uc.SeedAdmin(context.Background(), "admin", "adminpass")

	err := uc.SeedAdmin(context.Background(), "admin2", "adminpass")
	if err == nil {
		t.Error("expected error when admin already exists")
	}
}
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
go test ./usecase/... -run TestAuthUseCase -v 2>&1 | head -20
```

Expected: compile error — `usecase.NewAuthUseCase` not defined yet.

- [ ] **Step 3: Add auth DTOs to usecase/dto.go**

Append to the end of `usecase/dto.go` (after all existing types):

```go
// --- Auth ---

// LoginOutput is returned by AuthUseCase.Login.
type LoginOutput struct {
	Token    string
	Username string
	Role     string
}

// CreateUserInput is the input for AuthUseCase.CreateUser.
type CreateUserInput struct {
	Username string
	Password string
	Role     string
}

// UpdateUserInput is the input for AuthUseCase.UpdateUser.
// Empty fields are ignored (no update performed for that field).
type UpdateUserInput struct {
	Role     string
	Password string
}

// UserDTO is the output representation of a User (never includes PasswordHash).
type UserDTO struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}
```

- [ ] **Step 4: Create usecase/auth_usecase.go**

```go
package usecase

import (
	"context"
	"fmt"
	"time"

	"DartScheduler/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication and user management.
type AuthUseCase struct {
	users     domain.UserRepository
	jwtSecret string
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(users domain.UserRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{users: users, jwtSecret: jwtSecret}
}

// Login verifies credentials and returns a signed JWT token valid for 7 days.
func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (LoginOutput, error) {
	user, err := uc.users.FindByUsername(ctx, username)
	if err != nil {
		return LoginOutput{}, domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginOutput{}, domain.ErrInvalidCredentials
	}
	token, err := uc.generateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("generate token: %w", err)
	}
	return LoginOutput{Token: token, Username: user.Username, Role: user.Role}, nil
}

// CreateUser creates a new user with a bcrypt-hashed password.
func (uc *AuthUseCase) CreateUser(ctx context.Context, in CreateUserInput) (UserDTO, error) {
	if in.Username == "" || in.Password == "" {
		return UserDTO{}, domain.ErrInvalidInput
	}
	if in.Role != domain.RoleViewer && in.Role != domain.RoleMaintainer && in.Role != domain.RoleAdmin {
		return UserDTO{}, domain.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
	if err != nil {
		return UserDTO{}, fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{
		ID:           uuid.New().String(),
		Username:     in.Username,
		PasswordHash: string(hash),
		Role:         in.Role,
		CreatedAt:    time.Now(),
	}
	if err := uc.users.Create(ctx, user); err != nil {
		return UserDTO{}, err
	}
	return UserDTO{ID: user.ID, Username: user.Username, Role: user.Role, CreatedAt: user.CreatedAt}, nil
}

// ListUsers returns all users without password hashes.
func (uc *AuthUseCase) ListUsers(ctx context.Context) ([]UserDTO, error) {
	users, err := uc.users.List(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]UserDTO, len(users))
	for i, u := range users {
		dtos[i] = UserDTO{ID: u.ID, Username: u.Username, Role: u.Role, CreatedAt: u.CreatedAt}
	}
	return dtos, nil
}

// UpdateUser updates a user's role and/or password. Empty fields are skipped.
func (uc *AuthUseCase) UpdateUser(ctx context.Context, id string, in UpdateUserInput) error {
	if in.Role != "" {
		if in.Role != domain.RoleViewer && in.Role != domain.RoleMaintainer && in.Role != domain.RoleAdmin {
			return domain.ErrInvalidInput
		}
		if err := uc.users.UpdateRole(ctx, id, in.Role); err != nil {
			return err
		}
	}
	if in.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}
		if err := uc.users.UpdatePassword(ctx, id, string(hash)); err != nil {
			return err
		}
	}
	return nil
}

// DeleteUser deletes a user. Returns ErrInvalidInput if requestorID == id (cannot delete self).
func (uc *AuthUseCase) DeleteUser(ctx context.Context, id, requestorID string) error {
	if id == requestorID {
		return fmt.Errorf("%w: cannot delete own account", domain.ErrInvalidInput)
	}
	return uc.users.Delete(ctx, id)
}

// SeedAdmin creates the first admin user. Returns an error if any admin already exists.
func (uc *AuthUseCase) SeedAdmin(ctx context.Context, username, password string) error {
	exists, err := uc.users.ExistsAdmin(ctx)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("an admin user already exists")
	}
	_, err = uc.CreateUser(ctx, CreateUserInput{Username: username, Password: password, Role: domain.RoleAdmin})
	return err
}

func (uc *AuthUseCase) generateToken(userID, username, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
```

- [ ] **Step 5: Run tests — verify they pass**

```bash
go test ./usecase/... -run TestAuthUseCase -v
```

Expected: all 8 tests PASS

- [ ] **Step 6: Commit**

```bash
git add usecase/dto.go usecase/auth_usecase.go usecase/auth_usecase_test.go
git commit -m "feat: add AuthUseCase with login, user CRUD, and seed-admin"
```

---

## Task 5: Auth middleware + tests

**Files:**
- Create: `infra/http/middleware/auth.go`
- Create: `infra/http/middleware/auth_test.go`

- [ ] **Step 1: Write the failing middleware tests**

Create `infra/http/middleware/auth_test.go`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "DartScheduler/infra/http/middleware"

	"github.com/golang-jwt/jwt/v5"
)

func makeJWT(secret, userID, username, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":      userID,
		"username": username,
		"role":     role,
		"exp":      exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestResolveIdentity_LocalNetwork(t *testing.T) {
	tests := []struct {
		addr    string
		isLocal bool
	}{
		{"192.168.1.100:54321", true},
		{"192.168.0.1:80", true},
		{"10.0.0.1:8080", false},
		{"127.0.0.1:8080", false},
		{"8.8.8.8:443", false},
	}
	for _, tc := range tests {
		r := &http.Request{RemoteAddr: tc.addr, Header: http.Header{}}
		id, ok := mw.ResolveIdentity(r, "secret")
		if tc.isLocal {
			if !ok || id.Role != "admin" || id.Username != "lokaal netwerk" {
				t.Errorf("addr=%q: expected local-admin identity, got ok=%v id=%+v", tc.addr, ok, id)
			}
		} else {
			// non-local without JWT → not ok
			if ok {
				t.Errorf("addr=%q: expected ok=false for non-local without JWT", tc.addr)
			}
		}
	}
}

func TestResolveIdentity_ValidJWT(t *testing.T) {
	secret := "test-secret"
	token := makeJWT(secret, "user-1", "alice", "admin", time.Now().Add(time.Hour))
	r := &http.Request{
		RemoteAddr: "10.0.0.1:1234",
		Header:     http.Header{"Authorization": {"Bearer " + token}},
	}
	id, ok := mw.ResolveIdentity(r, secret)
	if !ok {
		t.Fatal("expected ok=true for valid JWT")
	}
	if id.Username != "alice" || id.Role != "admin" || id.UserID != "user-1" {
		t.Errorf("unexpected identity: %+v", id)
	}
}

func TestResolveIdentity_ExpiredJWT(t *testing.T) {
	secret := "test-secret"
	token := makeJWT(secret, "user-1", "alice", "admin", time.Now().Add(-time.Hour))
	r := &http.Request{
		RemoteAddr: "10.0.0.1:1234",
		Header:     http.Header{"Authorization": {"Bearer " + token}},
	}
	_, ok := mw.ResolveIdentity(r, secret)
	if ok {
		t.Error("expected ok=false for expired token")
	}
}

func TestResolveIdentity_WrongSecret(t *testing.T) {
	token := makeJWT("secret-a", "user-1", "alice", "admin", time.Now().Add(time.Hour))
	r := &http.Request{
		RemoteAddr: "10.0.0.1:1234",
		Header:     http.Header{"Authorization": {"Bearer " + token}},
	}
	_, ok := mw.ResolveIdentity(r, "secret-b")
	if ok {
		t.Error("expected ok=false for wrong secret")
	}
}

func TestAuth_Middleware_LocalNetworkGetsAdmin(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := mw.IdentityFromContext(r.Context())
		if !ok {
			http.Error(w, "no identity", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(id.Role))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.5:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	if rr.Body.String() != "admin" {
		t.Errorf("role: got %q, want %q", rr.Body.String(), "admin")
	}
}

func TestAuth_Middleware_NoAuth_Returns401(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestRequireRole_BlocksWrongRole(t *testing.T) {
	secret := "secret"
	token := makeJWT(secret, "u1", "bob", "viewer", time.Now().Add(time.Hour))
	handler := mw.Auth(secret)(mw.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestRequireRole_AllowsCorrectRole(t *testing.T) {
	secret := "secret"
	token := makeJWT(secret, "u1", "bob", "maintainer", time.Now().Add(time.Hour))
	handler := mw.Auth(secret)(mw.RequireRole("maintainer", "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
}
```

- [ ] **Step 2: Run tests — verify they fail**

```bash
go test ./infra/http/middleware/... -v 2>&1 | head -10
```

Expected: compile error — `mw.ResolveIdentity`, `mw.Auth`, `mw.RequireRole` not defined yet.

- [ ] **Step 3: Create infra/http/middleware/auth.go**

```go
package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const identityKey contextKey = "identity"

// Identity holds the resolved request identity injected into context by Auth middleware.
type Identity struct {
	UserID   string
	Username string
	Role     string
}

// IdentityFromContext retrieves the Identity set by Auth middleware.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey).(Identity)
	return id, ok
}

// Auth middleware resolves request identity from network trust (192.168.0.0/16) or JWT Bearer token.
// Returns 401 if neither resolves a valid identity.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := resolveIdentity(r, jwtSecret)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), identityKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware allows only requests whose identity role is in the given list.
// Must be used after Auth middleware.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok || !allowed[identity.Role] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ResolveIdentity resolves identity from network trust or JWT.
// Exported so auth_handler can use it for /api/auth/me without going through middleware.
func ResolveIdentity(r *http.Request, jwtSecret string) (Identity, bool) {
	return resolveIdentity(r, jwtSecret)
}

func resolveIdentity(r *http.Request, jwtSecret string) (Identity, bool) {
	if isLocalNetwork(r) {
		return Identity{UserID: "local-network", Username: "lokaal netwerk", Role: "admin"}, true
	}
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return Identity{}, false
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return Identity{}, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Identity{}, false
	}
	sub, _ := claims["sub"].(string)
	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)
	if sub == "" || role == "" {
		return Identity{}, false
	}
	return Identity{UserID: sub, Username: username, Role: role}, true
}

// isLocalNetwork reports whether the request's remote address is in 192.168.0.0/16.
// Only RemoteAddr is consulted — X-Forwarded-For is never trusted.
func isLocalNetwork(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return strings.HasPrefix(host, "192.168.")
}
```

- [ ] **Step 4: Run tests — verify they pass**

```bash
go test ./infra/http/middleware/... -v
```

Expected: all 8 tests PASS

- [ ] **Step 5: Commit**

```bash
git add infra/http/middleware/auth.go infra/http/middleware/auth_test.go
git commit -m "feat: add Auth middleware with network trust bypass and JWT validation"
```

---

## Task 6: Auth handler + User handler

**Files:**
- Create: `infra/http/handler/auth_handler.go`
- Create: `infra/http/handler/user_handler.go`

- [ ] **Step 1: Create infra/http/handler/auth_handler.go**

```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"DartScheduler/domain"
	mw "DartScheduler/infra/http/middleware"
	"DartScheduler/usecase"
)

// AuthHandler handles login and identity probe requests.
type AuthHandler struct {
	authUC    *usecase.AuthUseCase
	jwtSecret string
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUC *usecase.AuthUseCase, jwtSecret string) *AuthHandler {
	return &AuthHandler{authUC: authUC, jwtSecret: jwtSecret}
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	out, err := h.authUC.Login(r.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{
		"token":    out.Token,
		"username": out.Username,
		"role":     out.Role,
	})
}

// Me handles GET /api/auth/me — resolves identity from network trust or JWT without middleware.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	identity, ok := mw.ResolveIdentity(r, h.jwtSecret)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]string{
		"username": identity.Username,
		"role":     identity.Role,
	})
}
```

- [ ] **Step 2: Create infra/http/handler/user_handler.go**

```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"DartScheduler/domain"
	mw "DartScheduler/infra/http/middleware"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
)

// UserHandler handles user management endpoints (admin only).
type UserHandler struct {
	authUC *usecase.AuthUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(authUC *usecase.AuthUseCase) *UserHandler {
	return &UserHandler{authUC: authUC}
}

// List handles GET /api/users.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.authUC.ListUsers(r.Context())
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, users)
}

// Create handles POST /api/users.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto, err := h.authUC.CreateUser(r.Context(), usecase.CreateUserInput{
		Username: body.Username,
		Password: body.Password,
		Role:     body.Role,
	})
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, dto)
}

// Update handles PUT /api/users/:id — updates role and/or password.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.authUC.UpdateUser(r.Context(), id, usecase.UpdateUserInput{
		Role:     body.Role,
		Password: body.Password,
	}); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/users/:id.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	identity, _ := mw.IdentityFromContext(r.Context())
	if err := h.authUC.DeleteUser(r.Context(), id, identity.UserID); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./infra/http/handler/...
```

Expected: no output

- [ ] **Step 4: Commit**

```bash
git add infra/http/handler/auth_handler.go infra/http/handler/user_handler.go
git commit -m "feat: add AuthHandler (login/me) and UserHandler (CRUD)"
```

---

## Task 7: Wire backend — config, server routes, main, seed command

**Files:**
- Modify: `cmd/server/config.go`
- Modify: `infra/http/server.go`
- Modify: `cmd/server/main.go`

- [ ] **Step 1: Add JWTSecret to config.go**

In `cmd/server/config.go`, add to the `AppConfig` struct after `PrimaryColor`:

```go
// JWTSecret is the HS256 signing secret for JWT tokens.
// Set via JWT_SECRET env var. Falls back to an insecure default in development.
JWTSecret string
```

In `loadConfig()`, add after the PrimaryColor block:

```go
if v := os.Getenv("JWT_SECRET"); v != "" {
    cfg.JWTSecret = v
} else {
    log.Println("[WARN] JWT_SECRET not set — using insecure default (development only)")
    cfg.JWTSecret = "dart-scheduler-dev-secret-change-in-production"
}
```

- [ ] **Step 2: Replace infra/http/server.go with role-grouped routes**

Replace the entire contents of `infra/http/server.go`:

```go
// Package http registers all API routes and mounts the Angular SPA handler.
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
	systemH *handler.SystemHandler,
	eveningStatH *handler.EveningStatHandler,
	seasonStatH *handler.SeasonStatHandler,
	configH *handler.ConfigHandler,
	progressH *handler.ProgressHandler,
	playerListH *handler.PlayerListHandler,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	allowedOrigin string,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(mw.Logger)
	r.Use(mw.CORSWithOrigin(allowedOrigin))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		// Public — no authentication required
		r.Get("/config", configH.GetConfig)
		r.Post("/auth/login", authH.Login)
		r.Get("/auth/me", authH.Me)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(jwtSecret))

			// viewer+ — any authenticated identity
			r.Get("/schedules", schedH.List)
			r.Get("/schedules/{id}", schedH.GetByID)
			r.Get("/schedule", schedH.Get)
			r.Get("/schedule/evening/{id}", schedH.GetEvening)
			r.Get("/players", playerH.List)
			r.Get("/player-lists", playerListH.List)

			// maintainer+ — score entry and absence reporting
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireRole("maintainer", "admin"))
				r.Put("/matches/{id}/score", scoreH.Submit)
				r.Post("/evenings/{id}/report-absent", scoreH.ReportAbsent)
				r.Get("/evenings/{id}/player-stats", eveningStatH.GetByEvening)
				r.Put("/evenings/{id}/player-stats/{playerId}", eveningStatH.Upsert)
			})

			// admin only — everything else
			r.Group(func(r chi.Router) {
				r.Use(mw.RequireRole("admin"))

				r.Post("/import", playerH.Import)
				r.Put("/players/{id}", playerH.Update)
				r.Delete("/players/{id}", playerH.Delete)
				r.Get("/players/{id}/buddies", playerH.GetBuddies)
				r.Put("/players/{id}/buddies", playerH.SetBuddies)

				r.Post("/schedule/generate", schedH.Generate)
				r.Get("/schedules/{id}/info", schedH.GetInfo)
				r.Get("/schedules/{id}/matches", schedH.GetPlayedMatches)
				r.Post("/schedules/import-season", schedH.ImportSeason)
				r.Patch("/schedules/{id}", schedH.RenameSchedule)
				r.Delete("/schedules/{id}", schedH.DeleteSchedule)
				r.Post("/schedules/{id}/regenerate", schedH.RegenerateSchedule)
				r.Post("/schedules/{id}/active", schedH.SetActive)
				r.Post("/schedules/{id}/inhaal-avond", schedH.AddCatchUpEvening)
				r.Delete("/schedules/{id}/evenings/{eveningId}", schedH.DeleteEvening)

				r.Get("/stats", statsH.Get)
				r.Get("/stats/duties", statsH.GetDuties)
				r.Get("/stats/pdf", statsH.StandingsPDF)

				r.Get("/export/excel", exportH.Excel)
				r.Get("/export/pdf", exportH.PDF)
				r.Get("/export/evening/{id}/excel", exportH.EveningExcel)
				r.Get("/export/evening/{id}/pdf", exportH.EveningPDF)
				r.Get("/export/evening/{id}/print", exportH.EveningPrint)

				r.Get("/schedules/{id}/player-stats", seasonStatH.GetBySchedule)
				r.Put("/schedules/{id}/player-stats/{playerId}", seasonStatH.Upsert)

				r.Get("/system/logs", systemH.GetLogs)
				r.Get("/progress", progressH.GetProgress)

				// User management
				r.Get("/users", userH.List)
				r.Post("/users", userH.Create)
				r.Put("/users/{id}", userH.Update)
				r.Delete("/users/{id}", userH.Delete)
			})
		})
	})

	// SPA fallback: serve Angular app for all non-API routes.
	r.Handle("/*", web.SPAHandler())

	return r
}
```

- [ ] **Step 3: Update cmd/server/main.go — wire auth + seed command**

Replace the entire contents of `cmd/server/main.go`:

```go
package main

import (
	"context"
	"fmt"
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
	if len(os.Args) > 1 && os.Args[1] == "seed-admin" {
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "Usage: server seed-admin <username> <password>\n")
			os.Exit(1)
		}
		runSeedAdmin(os.Args[2], os.Args[3])
		return
	}

	cfg := loadConfig()

	logBuf := logbuf.New(200)
	log.SetOutput(io.MultiWriter(os.Stderr, logBuf))
	log.Printf("[INFO] config: port=%s db_type=%s db_path=%s club=%q title=%q logo=%q cors=%q",
		cfg.Port, cfg.DatabaseType, cfg.DatabasePath, cfg.ClubName, cfg.AppTitle, cfg.LogoPath, cfg.AllowedOrigin)

	db, err := sqlite.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Repositories
	playerRepo := sqlite.NewPlayerRepo(db)
	scheduleRepo := sqlite.NewScheduleRepo(db)
	eveningRepo := sqlite.NewEveningRepo(db)
	matchRepo := sqlite.NewMatchRepo(db)
	eveningStatRepo := sqlite.NewEveningPlayerStatRepo(db)
	seasonStatRepo := sqlite.NewSeasonPlayerStatRepo(db)
	playerListRepo := sqlite.NewPlayerListRepo(db)
	userRepo := sqlite.NewUserRepo(db)

	// Use cases
	playerUC := usecase.NewPlayerUseCase(playerRepo, matchRepo, playerListRepo)
	scheduleUC := usecase.NewScheduleUseCase(playerRepo, scheduleRepo, eveningRepo, matchRepo)
	scoreUC := usecase.NewScoreUseCase(matchRepo, eveningRepo, seasonStatRepo)
	exportUC := usecase.NewExportUseCase(scheduleRepo, eveningRepo, matchRepo, playerRepo)
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWTSecret)

	// Log database summary at startup
	if players, err := playerUC.ListPlayers(context.Background()); err == nil {
		schedules, _ := scheduleUC.ListSchedules(context.Background())
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
	authH := handler.NewAuthHandler(authUC, cfg.JWTSecret)
	userH := handler.NewUserHandler(authUC)

	router := apphttp.NewRouter(
		playerH, schedH, scoreH, statsH, exportH, systemH,
		eveningStatH, seasonStatH, configH, progressH, playerListH,
		authH, userH,
		cfg.AllowedOrigin, cfg.JWTSecret,
	)

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}

func runSeedAdmin(username, password string) {
	cfg := loadConfig()
	db, err := sqlite.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	userRepo := sqlite.NewUserRepo(db)
	authUC := usecase.NewAuthUseCase(userRepo, cfg.JWTSecret)

	if err := authUC.SeedAdmin(context.Background(), username, password); err != nil {
		log.Fatalf("seed-admin: %v", err)
	}
	fmt.Printf("Admin user %q created successfully.\n", username)
}
```

- [ ] **Step 4: Build and run all Go tests**

```bash
go build ./cmd/server/ && go test ./...
```

Expected: build succeeds, all existing + new tests pass (no failures)

- [ ] **Step 5: Commit**

```bash
git add cmd/server/config.go infra/http/server.go cmd/server/main.go
git commit -m "feat: wire auth components, restructure API routes by role, add seed-admin command"
```

---

## Task 8: Angular AuthService

**Files:**
- Create: `frontend/src/app/services/auth.service.ts`

- [ ] **Step 1: Create frontend/src/app/services/auth.service.ts**

```typescript
import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';

interface CurrentUser {
  username: string;
  role: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private router = inject(Router);

  private _currentUser = signal<CurrentUser | null>(null);
  readonly currentUser = this._currentUser.asReadonly();
  readonly role = computed(() => this._currentUser()?.role ?? null);
  readonly isLoggedIn = computed(() => this._currentUser() !== null);

  /**
   * Called once at app boot via APP_INITIALIZER.
   * Resolves identity from either network trust or a stored JWT.
   * Always resolves (never throws) — sets currentUser to null on any failure.
   */
  async init(): Promise<void> {
    const token = localStorage.getItem('dart_token');
    if (token && this.isTokenExpired(token)) {
      localStorage.removeItem('dart_token');
    }
    try {
      const user = await firstValueFrom(
        this.http.get<CurrentUser>('/api/auth/me')
      );
      this._currentUser.set(user);
    } catch {
      this._currentUser.set(null);
    }
  }

  async login(username: string, password: string): Promise<void> {
    const resp = await firstValueFrom(
      this.http.post<{ token: string; username: string; role: string }>(
        '/api/auth/login',
        { username, password }
      )
    );
    localStorage.setItem('dart_token', resp.token);
    this._currentUser.set({ username: resp.username, role: resp.role });
  }

  logout(): void {
    localStorage.removeItem('dart_token');
    this._currentUser.set(null);
    this.router.navigate(['/login']);
  }

  private isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }
}
```

- [ ] **Step 2: Commit**

```bash
cd frontend
git add src/app/services/auth.service.ts
git commit -m "feat: add Angular AuthService with JWT localStorage management"
```

---

## Task 9: Angular interceptor + update app.config.ts

**Files:**
- Create: `frontend/src/app/interceptors/auth.interceptor.ts`
- Modify: `frontend/src/app/app.config.ts`

- [ ] **Step 1: Create frontend/src/app/interceptors/auth.interceptor.ts**

```typescript
import { HttpInterceptorFn, HttpErrorResponse } from '@angular/common/http';
import { inject } from '@angular/core';
import { Router } from '@angular/router';
import { catchError, throwError } from 'rxjs';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const token = localStorage.getItem('dart_token');
  if (token && req.url.includes('/api/')) {
    req = req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
  }
  return next(req).pipe(
    catchError((err: unknown) => {
      if (
        err instanceof HttpErrorResponse &&
        err.status === 401 &&
        !req.url.includes('/api/auth/')
      ) {
        localStorage.removeItem('dart_token');
        inject(Router).navigate(['/login']);
      }
      return throwError(() => err);
    })
  );
};
```

- [ ] **Step 2: Replace frontend/src/app/app.config.ts**

```typescript
import { ApplicationConfig, APP_INITIALIZER } from '@angular/core';
import { provideRouter } from '@angular/router';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { MAT_SELECT_CONFIG } from '@angular/material/select';

import { routes } from './app.routes';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { authInterceptor } from './interceptors/auth.interceptor';
import { AuthService } from './services/auth.service';

export const appConfig: ApplicationConfig = {
  providers: [
    provideRouter(routes),
    provideAnimationsAsync(),
    provideHttpClient(withInterceptors([authInterceptor])),
    { provide: MAT_SELECT_CONFIG, useValue: { typeaheadDebounceInterval: 430 } },
    {
      provide: APP_INITIALIZER,
      useFactory: (auth: AuthService) => () => auth.init(),
      deps: [AuthService],
      multi: true,
    },
  ],
};
```

- [ ] **Step 3: Commit**

```bash
git add src/app/interceptors/auth.interceptor.ts src/app/app.config.ts
git commit -m "feat: add auth interceptor (JWT header + 401 redirect) and APP_INITIALIZER"
```

---

## Task 10: Angular guards

**Files:**
- Create: `frontend/src/app/guards/auth.guard.ts`
- Create: `frontend/src/app/guards/role.guard.ts`

- [ ] **Step 1: Create frontend/src/app/guards/auth.guard.ts**

```typescript
import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (auth.isLoggedIn()) return true;
  return router.createUrlTree(['/login']);
};
```

- [ ] **Step 2: Create frontend/src/app/guards/role.guard.ts**

```typescript
import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

/**
 * Returns a CanActivateFn that allows access only when the current user's role
 * is in the provided list. Redirects to '/' on failure.
 */
export function roleGuard(...allowedRoles: string[]): CanActivateFn {
  return () => {
    const auth = inject(AuthService);
    const router = inject(Router);
    const role = auth.role();
    if (role && allowedRoles.includes(role)) return true;
    return router.createUrlTree(['/']);
  };
}
```

- [ ] **Step 3: Commit**

```bash
git add src/app/guards/auth.guard.ts src/app/guards/role.guard.ts
git commit -m "feat: add authGuard and roleGuard for Angular route protection"
```

---

## Task 11: Angular Login component

**Files:**
- Create: `frontend/src/app/components/login/login.component.ts`

- [ ] **Step 1: Create frontend/src/app/components/login/login.component.ts**

```typescript
import { Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { AsyncPipe } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '../../services/auth.service';
import { ConfigService } from '../../services/config.service';

@Component({
  selector: 'app-login',
  imports: [
    FormsModule,
    AsyncPipe,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatProgressSpinnerModule,
  ],
  styles: [`
    .login-page {
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      background: #f5f5f5;
    }
    .login-card { width: 360px; padding: 16px; }
    .login-title { text-align: center; margin-bottom: 24px; }
    .login-field { width: 100%; margin-bottom: 12px; }
    .login-actions { display: flex; justify-content: flex-end; margin-top: 8px; }
    .login-error { color: #c62828; font-size: 13px; margin-bottom: 8px; text-align: center; }
  `],
  template: `
    <div class="login-page">
      <mat-card class="login-card">
        <mat-card-content>
          <div class="login-title">
            <h2>{{ configService.clubName$ | async }}</h2>
            <p style="color:#757575;margin:0">{{ configService.appTitle$ | async }}</p>
          </div>
          @if (error()) {
            <div class="login-error">{{ error() }}</div>
          }
          <mat-form-field class="login-field" appearance="outline">
            <mat-label>Gebruikersnaam</mat-label>
            <input matInput [(ngModel)]="username" (keyup.enter)="submit()" autocomplete="username" />
          </mat-form-field>
          <mat-form-field class="login-field" appearance="outline">
            <mat-label>Wachtwoord</mat-label>
            <input matInput type="password" [(ngModel)]="password" (keyup.enter)="submit()" autocomplete="current-password" />
          </mat-form-field>
          <div class="login-actions">
            <button mat-raised-button color="primary" (click)="submit()" [disabled]="loading()">
              @if (loading()) {
                <mat-spinner diameter="20" style="display:inline-block;margin-right:8px"></mat-spinner>
              }
              Inloggen
            </button>
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
})
export class LoginComponent {
  protected configService = inject(ConfigService);
  private authService = inject(AuthService);
  private router = inject(Router);

  username = '';
  password = '';
  loading = signal(false);
  error = signal('');

  async submit(): Promise<void> {
    if (!this.username || !this.password) return;
    this.loading.set(true);
    this.error.set('');
    try {
      await this.authService.login(this.username, this.password);
      this.router.navigate(['/']);
    } catch {
      this.error.set('Ongeldige gebruikersnaam of wachtwoord.');
    } finally {
      this.loading.set(false);
    }
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add src/app/components/login/login.component.ts
git commit -m "feat: add LoginComponent with username/password form"
```

---

## Task 12: Update Angular routes + mobile routes

**Files:**
- Modify: `frontend/src/app/app.routes.ts`
- Modify: `frontend/src/app/mobile/mobile.routes.ts`

- [ ] **Step 1: Replace frontend/src/app/app.routes.ts**

```typescript
import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';
import { roleGuard } from './guards/role.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () =>
      import('./components/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: '',
    canActivate: [authGuard, roleGuard('viewer', 'maintainer', 'admin')],
    loadComponent: () =>
      import('./components/overview/overview.component').then((m) => m.OverviewComponent),
  },
  {
    path: 'spelers',
    canActivate: [authGuard, roleGuard('admin')],
    loadComponent: () =>
      import('./components/spelers/spelers.component').then((m) => m.SpelersComponent),
  },
  {
    path: 'evening/:id',
    canActivate: [authGuard, roleGuard('maintainer', 'admin')],
    loadComponent: () =>
      import('./components/evening-view/evening-view.component').then((m) => m.EveningViewComponent),
  },
  {
    path: 'standings',
    canActivate: [authGuard, roleGuard('admin')],
    loadComponent: () =>
      import('./components/standings/standings.component').then((m) => m.StandingsComponent),
  },
  {
    path: 'info',
    canActivate: [authGuard, roleGuard('admin')],
    loadComponent: () =>
      import('./components/info/info.component').then((m) => m.InfoComponent),
  },
  {
    path: 'beheer',
    canActivate: [authGuard, roleGuard('admin')],
    loadComponent: () =>
      import('./components/beheer/beheer.component').then((m) => m.BeheerComponent),
  },
  {
    path: 'gebruikers',
    canActivate: [authGuard, roleGuard('admin')],
    loadComponent: () =>
      import('./components/gebruikers/gebruikers.component').then((m) => m.GebruikersComponent),
  },
  {
    path: 'm',
    loadChildren: () => import('./mobile/mobile.routes').then((m) => m.mobileRoutes),
  },
  { path: '**', redirectTo: '' },
];
```

- [ ] **Step 2: Replace frontend/src/app/mobile/mobile.routes.ts**

```typescript
import { Routes } from '@angular/router';
import { authGuard } from '../guards/auth.guard';
import { roleGuard } from '../guards/role.guard';

export const mobileRoutes: Routes = [
  {
    path: '',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./mobile-shell.component').then((m) => m.MobileShellComponent),
    children: [
      { path: '', redirectTo: 'avond', pathMatch: 'full' },
      {
        path: 'avond',
        canActivate: [roleGuard('viewer', 'maintainer', 'admin')],
        loadComponent: () =>
          import('./mobile-avond.component').then((m) => m.MobileAvondComponent),
      },
      {
        path: 'stand',
        canActivate: [roleGuard('admin')],
        loadComponent: () =>
          import('./mobile-stand.component').then((m) => m.MobileStandComponent),
      },
      {
        path: 'stats',
        canActivate: [roleGuard('admin')],
        loadComponent: () =>
          import('./mobile-stats.component').then((m) => m.MobileStatsComponent),
      },
    ],
  },
  {
    path: 'score/:id',
    canActivate: [authGuard, roleGuard('maintainer', 'admin')],
    loadComponent: () =>
      import('./mobile-score.component').then((m) => m.MobileScoreComponent),
  },
];
```

- [ ] **Step 3: Commit**

```bash
git add src/app/app.routes.ts src/app/mobile/mobile.routes.ts
git commit -m "feat: add route guards — viewer/maintainer/admin protection on all routes"
```

---

## Task 13: Update app.component (nav + logout) and spec

**Files:**
- Modify: `frontend/src/app/app.component.ts`
- Modify: `frontend/src/app/app.component.html`
- Modify: `frontend/src/app/app.component.spec.ts`

- [ ] **Step 1: Update app.component.ts — inject AuthService**

Replace the entire `app.component.ts`:

```typescript
import { Component, inject, OnInit, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterOutlet, RouterLink, RouterLinkActive, Router, NavigationEnd } from '@angular/router';
import { CommonModule, AsyncPipe } from '@angular/common';
import { Title } from '@angular/platform-browser';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule } from '@angular/material/form-field';
import { filter } from 'rxjs';
import { SeasonService } from './services/season.service';
import { ConfigService } from './services/config.service';
import { AuthService } from './services/auth.service';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  imports: [
    RouterOutlet,
    RouterLink,
    RouterLinkActive,
    CommonModule,
    AsyncPipe,
    MatToolbarModule,
    MatButtonModule,
    MatIconModule,
    MatSelectModule,
    MatFormFieldModule,
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  protected seasonService = inject(SeasonService);
  protected configService = inject(ConfigService);
  protected authService = inject(AuthService);
  protected version = environment.version;
  private router = inject(Router);
  private titleService = inject(Title);
  private destroyRef = inject(DestroyRef);

  isMobile = false;

  ngOnInit(): void {
    this.seasonService.load();
    this.configService.load();
    this.configService.appTitle$
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe((title) => this.titleService.setTitle(title));
    this.router.events
      .pipe(
        filter((e) => e instanceof NavigationEnd),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe(() => {
        this.isMobile = this.router.url.startsWith('/m');
      });
  }
}
```

- [ ] **Step 2: Replace app.component.html with role-aware nav**

```html
@if (!isMobile) {
  <mat-toolbar color="primary" class="app-toolbar"
    [style.background-color]="(configService.primaryColor$ | async) || null">
    <img src="assets/logo.svg" alt="dart" class="toolbar-logo" />
    <a class="toolbar-title" routerLink="/">{{ configService.clubName$ | async }}</a>
    <span class="toolbar-spacer"></span>
    @if ((seasonService.seasons$ | async)?.length) {
      <mat-form-field appearance="outline" subscriptSizing="dynamic" class="toolbar-season-select">
        <mat-label>Seizoen</mat-label>
        <mat-select [value]="seasonService.selectedId$ | async" (selectionChange)="seasonService.select($event.value)">
          @for (s of seasonService.seasons$ | async; track s) {
            <mat-option [value]="s.id">
              {{ s.season || s.competitionName }}
            </mat-option>
          }
        </mat-select>
      </mat-form-field>
    }
    <nav class="toolbar-nav">
      <a mat-button routerLink="/" routerLinkActive="nav-active" [routerLinkActiveOptions]="{ exact: true }">
        <mat-icon>event_note</mat-icon> Schema
      </a>
      @if (authService.role() === 'admin') {
        <a mat-button routerLink="/spelers" routerLinkActive="nav-active">
          <mat-icon>people</mat-icon> Spelers
        </a>
        <a mat-button routerLink="/standings" routerLinkActive="nav-active">
          <mat-icon>leaderboard</mat-icon> Klassement
        </a>
        <a mat-button routerLink="/info" routerLinkActive="nav-active">
          <mat-icon>insights</mat-icon> Info
        </a>
        <a mat-button routerLink="/beheer" routerLinkActive="nav-active">
          <mat-icon>settings</mat-icon> Beheer
        </a>
        <a mat-button routerLink="/gebruikers" routerLinkActive="nav-active">
          <mat-icon>manage_accounts</mat-icon> Gebruikers
        </a>
      }
      <button mat-icon-button (click)="authService.logout()" title="Uitloggen">
        <mat-icon>logout</mat-icon>
      </button>
    </nav>
  </mat-toolbar>
}

<main [class.app-content]="!isMobile">
  <router-outlet />
</main>
```

- [ ] **Step 3: Update app.component.spec.ts to mock APP_INITIALIZER**

Replace the entire `app.component.spec.ts`:

```typescript
import { TestBed } from '@angular/core/testing';
import { AppComponent } from './app.component';
import { provideRouter } from '@angular/router';
import { provideHttpClient } from '@angular/common/http';
import { APP_INITIALIZER } from '@angular/core';

describe('AppComponent', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent],
      providers: [
        provideRouter([]),
        provideHttpClient(),
        // Skip the real init() call in unit tests
        { provide: APP_INITIALIZER, useValue: () => () => Promise.resolve(), multi: true },
      ],
    }).compileComponents();
  });

  it('should create the app', () => {
    const fixture = TestBed.createComponent(AppComponent);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
  });
});
```

- [ ] **Step 4: Commit**

```bash
git add src/app/app.component.ts src/app/app.component.html src/app/app.component.spec.ts
git commit -m "feat: role-aware navigation with logout button in app toolbar"
```

---

## Task 14: Gebruikers component (user management)

**Files:**
- Create: `frontend/src/app/components/gebruikers/gebruikers.component.ts`

- [ ] **Step 1: Create frontend/src/app/components/gebruikers/gebruikers.component.ts**

```typescript
import { Component, inject, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { AuthService } from '../../services/auth.service';

interface User {
  id: string;
  username: string;
  role: string;
  createdAt: string;
}

@Component({
  selector: 'app-gebruikers',
  imports: [
    CommonModule,
    FormsModule,
    MatCardModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
  ],
  styles: [`
    .page { padding: 24px; }
    h2 { margin: 0 0 20px 0; }
    .add-form {
      display: flex;
      gap: 12px;
      align-items: flex-end;
      flex-wrap: wrap;
      margin-bottom: 24px;
      padding: 16px;
      background: #f5f5f5;
      border-radius: 8px;
    }
    .add-form mat-form-field { min-width: 180px; }
    .error { color: #c62828; font-size: 13px; margin-top: 8px; }
    .role-chip {
      display: inline-block;
      padding: 2px 10px;
      border-radius: 12px;
      font-size: 12px;
      font-weight: 600;
    }
    .role-admin { background: #e8eaf6; color: #3949ab; }
    .role-maintainer { background: #e8f5e9; color: #2e7d32; }
    .role-viewer { background: #fafafa; color: #616161; border: 1px solid #e0e0e0; }
  `],
  template: `
    <div class="page">
      <h2>Gebruikersbeheer</h2>

      <!-- Add user form -->
      <div class="add-form">
        <mat-form-field subscriptSizing="dynamic">
          <mat-label>Gebruikersnaam</mat-label>
          <input matInput [(ngModel)]="newUsername" />
        </mat-form-field>
        <mat-form-field subscriptSizing="dynamic">
          <mat-label>Wachtwoord</mat-label>
          <input matInput type="password" [(ngModel)]="newPassword" />
        </mat-form-field>
        <mat-form-field subscriptSizing="dynamic" style="min-width:140px">
          <mat-label>Rol</mat-label>
          <mat-select [(ngModel)]="newRole">
            <mat-option value="viewer">Viewer</mat-option>
            <mat-option value="maintainer">Maintainer</mat-option>
            <mat-option value="admin">Admin</mat-option>
          </mat-select>
        </mat-form-field>
        <button mat-raised-button color="primary" (click)="addUser()" [disabled]="!newUsername || !newPassword || !newRole">
          <mat-icon>person_add</mat-icon> Toevoegen
        </button>
        @if (addError()) {
          <span class="error">{{ addError() }}</span>
        }
      </div>

      <!-- User table -->
      <mat-card>
        <mat-card-content>
          <table mat-table [dataSource]="users()" style="width:100%">
            <ng-container matColumnDef="username">
              <th mat-header-cell *matHeaderCellDef>Gebruikersnaam</th>
              <td mat-cell *matCellDef="let u"><strong>{{ u.username }}</strong></td>
            </ng-container>
            <ng-container matColumnDef="role">
              <th mat-header-cell *matHeaderCellDef style="width:140px">Rol</th>
              <td mat-cell *matCellDef="let u">
                <mat-select [value]="u.role" (selectionChange)="updateRole(u, $event.value)"
                            style="font-size:13px">
                  <mat-option value="viewer">Viewer</mat-option>
                  <mat-option value="maintainer">Maintainer</mat-option>
                  <mat-option value="admin">Admin</mat-option>
                </mat-select>
              </td>
            </ng-container>
            <ng-container matColumnDef="createdAt">
              <th mat-header-cell *matHeaderCellDef style="width:160px">Aangemaakt</th>
              <td mat-cell *matCellDef="let u" style="font-size:12px;color:#757575">
                {{ u.createdAt | date:'d MMM yyyy' }}
              </td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef style="width:200px;text-align:right">Acties</th>
              <td mat-cell *matCellDef="let u" style="text-align:right">
                <button mat-stroked-button style="margin-right:8px;font-size:12px"
                        (click)="promptResetPassword(u)">
                  <mat-icon style="font-size:16px;height:16px;width:16px">lock_reset</mat-icon>
                  Wachtwoord
                </button>
                <button mat-icon-button color="warn"
                        [disabled]="u.id === currentUserId()"
                        [title]="u.id === currentUserId() ? 'Kan eigen account niet verwijderen' : 'Verwijderen'"
                        (click)="deleteUser(u)">
                  <mat-icon>delete</mat-icon>
                </button>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="cols"></tr>
            <tr mat-row *matRowDef="let row; columns: cols"></tr>
          </table>
          @if (users().length === 0) {
            <p style="color:#9e9e9e;text-align:center;padding:24px 0;margin:0">Geen gebruikers gevonden.</p>
          }
        </mat-card-content>
      </mat-card>
    </div>
  `,
})
export class GebruikersComponent implements OnInit {
  private http = inject(HttpClient);
  private authService = inject(AuthService);

  cols = ['username', 'role', 'createdAt', 'actions'];
  users = signal<User[]>([]);
  addError = signal('');

  newUsername = '';
  newPassword = '';
  newRole = 'viewer';

  currentUserId = signal('');

  ngOnInit(): void {
    this.loadUsers();
    // Decode current user ID from JWT for self-delete guard
    const token = localStorage.getItem('dart_token');
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        this.currentUserId.set(payload.sub ?? '');
      } catch {
        // local-network user — no token, use sentinel
        this.currentUserId.set('local-network');
      }
    } else {
      this.currentUserId.set('local-network');
    }
  }

  private loadUsers(): void {
    this.http.get<User[]>('/api/users').subscribe({
      next: (users) => this.users.set(users),
      error: () => this.users.set([]),
    });
  }

  addUser(): void {
    this.addError.set('');
    this.http.post<User>('/api/users', {
      username: this.newUsername,
      password: this.newPassword,
      role: this.newRole,
    }).subscribe({
      next: (user) => {
        this.users.update((list) => [...list, user]);
        this.newUsername = '';
        this.newPassword = '';
        this.newRole = 'viewer';
      },
      error: (err) => {
        this.addError.set(err.status === 409 ? 'Gebruikersnaam bestaat al.' : 'Aanmaken mislukt.');
      },
    });
  }

  updateRole(user: User, role: string): void {
    this.http.put(`/api/users/${user.id}`, { role }).subscribe({
      next: () => this.users.update((list) => list.map((u) => u.id === user.id ? { ...u, role } : u)),
      error: () => alert('Rol wijzigen mislukt.'),
    });
  }

  promptResetPassword(user: User): void {
    const pw = prompt(`Nieuw wachtwoord voor ${user.username}:`);
    if (!pw) return;
    this.http.put(`/api/users/${user.id}`, { password: pw }).subscribe({
      next: () => alert('Wachtwoord gewijzigd.'),
      error: () => alert('Wachtwoord wijzigen mislukt.'),
    });
  }

  deleteUser(user: User): void {
    if (!confirm(`Gebruiker "${user.username}" verwijderen?`)) return;
    this.http.delete(`/api/users/${user.id}`).subscribe({
      next: () => this.users.update((list) => list.filter((u) => u.id !== user.id)),
      error: () => alert('Verwijderen mislukt.'),
    });
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add src/app/components/gebruikers/gebruikers.component.ts
git commit -m "feat: add GebruikersComponent for admin user management"
```

---

## Task 15: Build + verify

**Files:** none (verification only)

- [ ] **Step 1: Build Angular frontend**

```bash
cd frontend
npm run build 2>&1 | tail -20
```

Expected: `✓ Building...` with no errors. Output goes to `web/dist/`.

- [ ] **Step 2: Run Angular unit tests**

```bash
npm test -- --watch=false --browsers=ChromeHeadless 2>&1 | tail -20
```

Expected: `AppComponent > should create the app` PASSES, no failures.

- [ ] **Step 3: Build Go binary**

```bash
cd ..
go build ./cmd/server/
```

Expected: `server` binary created, no errors.

- [ ] **Step 4: Run all Go tests**

```bash
go test ./...
```

Expected: all tests PASS including new auth + middleware tests.

- [ ] **Step 5: Smoke test the seed command**

```bash
./server seed-admin testadmin testpass123
```

Expected: `Admin user "testadmin" created successfully.`

- [ ] **Step 6: Smoke test API**

```bash
# Should get 401 without auth
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/schedules

# Should get token
curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testadmin","password":"testpass123"}'
```

Expected: first command returns `401`, second returns JSON with `token`, `username`, `role`.

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "chore: verify authentication feature complete — all tests pass"
git push
```

---

## Usage After Deployment

1. Set `JWT_SECRET` environment variable to a strong random string.
2. Run `./server seed-admin <username> <password>` to create the first admin.
3. Log in at the `/login` page, or access from `192.168.x.x` for automatic admin access.
4. Go to **Gebruikers** to add viewer/maintainer/admin accounts.
