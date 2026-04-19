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
