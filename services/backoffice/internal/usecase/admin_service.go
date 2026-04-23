package usecase

import (
	"context"
	"fmt"
	"time"

	"raiiaa.dev/services/backoffice/internal/domain"
	"raiiaa.dev/services/backoffice/internal/ports"
)

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}

type AdminService struct {
	serviceName string
	clock       ports.Clock
	repo        ports.AdminRepository
}

func NewAdminService(serviceName string, clock ports.Clock, repo ports.AdminRepository) *AdminService {
	return &AdminService{
		serviceName: serviceName,
		clock:       clock,
		repo:        repo,
	}
}

func (s *AdminService) GetStatus(ctx context.Context) (domain.AdminStatus, error) {
	admins, err := s.repo.ListAdmins(ctx)
	if err != nil {
		return domain.AdminStatus{}, fmt.Errorf("list admins: %w", err)
	}

	activeAdmins := 0
	for _, admin := range admins {
		if admin.IsActive {
			activeAdmins++
		}
	}

	return domain.AdminStatus{
		Service:      s.serviceName,
		Status:       "ok",
		Timestamp:    s.clock.Now().UTC(),
		TotalAdmins:  len(admins),
		ActiveAdmins: activeAdmins,
	}, nil
}

func (s *AdminService) ListAdmins(ctx context.Context) ([]domain.AdminUser, error) {
	admins, err := s.repo.ListAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	return admins, nil
}
