package ports

import (
	"context"
	"errors"

	"raiiaa.dev/services/backoffice/internal/domain"
)

var (
	ErrAdminNotFound     = errors.New("admin not found")
	ErrAdminAlreadyExist = errors.New("admin already exists")
)

type AdminRepository interface {
	CreateAdmin(ctx context.Context, email string, role domain.AdminRole, isActive bool) (domain.AdminUser, error)
	GetAdminByID(ctx context.Context, id string) (domain.AdminUser, error)
	ListAdmins(ctx context.Context) ([]domain.AdminUser, error)
	UpdateAdmin(ctx context.Context, id, email string, role domain.AdminRole, isActive bool) (domain.AdminUser, error)
	DeleteAdmin(ctx context.Context, id string) error
}
