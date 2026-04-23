package ports

import (
	"context"

	"raiiaa.dev/services/backoffice/internal/domain"
)

type AdminRepository interface {
	ListAdmins(ctx context.Context) ([]domain.AdminUser, error)
}
