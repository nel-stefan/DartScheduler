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
