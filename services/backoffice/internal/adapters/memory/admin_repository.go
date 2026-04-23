package memory

import (
	"context"
	"time"

	"raiiaa.dev/services/backoffice/internal/domain"
	"raiiaa.dev/services/backoffice/internal/ports"
)

var _ ports.AdminRepository = (*AdminRepository)(nil)

type AdminRepository struct {
	admins []domain.AdminUser
}

func NewAdminRepository() *AdminRepository {
	now := time.Now().UTC()
	return &AdminRepository{
		admins: []domain.AdminUser{
			{
				ID:        "adm_001",
				Email:     "owner@raiiaa.dev",
				Role:      domain.AdminRoleOwner,
				IsActive:  true,
				CreatedAt: now.AddDate(0, -8, 0),
			},
			{
				ID:        "adm_002",
				Email:     "ops@raiiaa.dev",
				Role:      domain.AdminRoleEditor,
				IsActive:  true,
				CreatedAt: now.AddDate(0, -4, 0),
			},
			{
				ID:        "adm_003",
				Email:     "auditor@raiiaa.dev",
				Role:      domain.AdminRoleViewer,
				IsActive:  false,
				CreatedAt: now.AddDate(0, -1, 0),
			},
		},
	}
}

func (r *AdminRepository) ListAdmins(_ context.Context) ([]domain.AdminUser, error) {
	admins := make([]domain.AdminUser, len(r.admins))
	copy(admins, r.admins)
	return admins, nil
}
