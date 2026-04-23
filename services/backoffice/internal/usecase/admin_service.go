package usecase

import (
	"context"
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
}

func NewAdminService(serviceName string, clock ports.Clock) *AdminService {
	return &AdminService{
		serviceName: serviceName,
		clock:       clock,
	}
}

func (s *AdminService) GetStatus(_ context.Context) domain.AdminStatus {
	return domain.AdminStatus{
		Service:   s.serviceName,
		Status:    "ok",
		Timestamp: s.clock.Now().UTC(),
	}
}
