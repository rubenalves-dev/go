package memory

import (
	"context"
	"errors"
	"strings"
	"time"

	"raiiaa.dev/services/backoffice/internal/domain"
)

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

func (r *AdminRepository) CreateAdmin(_ context.Context, email string, role domain.AdminRole, isActive bool) (domain.AdminUser, error) {
	for _, admin := range r.admins {
		if admin.Email == email {
			return domain.AdminUser{}, errors.New("admin already exists")
		}
	}

	admin := domain.AdminUser{
		ID:        "adm_mem_" + time.Now().UTC().Format("20060102150405.000000000"),
		Email:     email,
		Role:      role,
		IsActive:  isActive,
		CreatedAt: time.Now().UTC(),
	}
	r.admins = append(r.admins, admin)
	return admin, nil
}

func (r *AdminRepository) GetAdminByID(_ context.Context, id string) (domain.AdminUser, error) {
	for _, admin := range r.admins {
		if admin.ID == id {
			return admin, nil
		}
	}
	return domain.AdminUser{}, errors.New("admin not found")
}

func (r *AdminRepository) UpdateAdmin(_ context.Context, id, email string, role domain.AdminRole, isActive bool) (domain.AdminUser, error) {
	for i, admin := range r.admins {
		if admin.ID == id {
			for _, existing := range r.admins {
				if existing.ID != id && strings.EqualFold(existing.Email, email) {
					return domain.AdminUser{}, errors.New("admin already exists")
				}
			}
			admin.Email = email
			admin.Role = role
			admin.IsActive = isActive
			r.admins[i] = admin
			return admin, nil
		}
	}
	return domain.AdminUser{}, errors.New("admin not found")
}

func (r *AdminRepository) DeleteAdmin(_ context.Context, id string) error {
	for i, admin := range r.admins {
		if admin.ID == id {
			r.admins = append(r.admins[:i], r.admins[i+1:]...)
			return nil
		}
	}
	return errors.New("admin not found")
}
