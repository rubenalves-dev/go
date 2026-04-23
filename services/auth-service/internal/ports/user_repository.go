package ports

import (
	"context"
	"errors"

	"raiiaa.dev/services/auth-service/internal/domain"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserAlreadyExist = errors.New("user already exists")
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}
