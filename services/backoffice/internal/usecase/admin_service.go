package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"raiiaa.dev/services/backoffice/internal/domain"
	"raiiaa.dev/services/backoffice/internal/ports"
)

var (
	ErrInvalidAdminID    = errors.New("invalid admin id")
	ErrInvalidAdminEmail = errors.New("invalid admin email")
	ErrInvalidAdminRole  = errors.New("invalid admin role")
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

func (s *AdminService) GetAdminByID(ctx context.Context, id string) (domain.AdminUser, error) {
	adminID := strings.TrimSpace(id)
	if adminID == "" {
		return domain.AdminUser{}, ErrInvalidAdminID
	}

	admin, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("get admin by id: %w", err)
	}
	return admin, nil
}

func (s *AdminService) CreateAdmin(ctx context.Context, email, role string, isActive bool) (domain.AdminUser, error) {
	adminEmail := normalizeEmail(email)
	if !isValidEmail(adminEmail) {
		return domain.AdminUser{}, ErrInvalidAdminEmail
	}

	adminRole, err := parseAdminRole(role)
	if err != nil {
		return domain.AdminUser{}, err
	}

	admin, err := s.repo.CreateAdmin(ctx, adminEmail, adminRole, isActive)
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("create admin: %w", err)
	}
	return admin, nil
}

func (s *AdminService) UpdateAdmin(ctx context.Context, id, email, role string, isActive bool) (domain.AdminUser, error) {
	adminID := strings.TrimSpace(id)
	if adminID == "" {
		return domain.AdminUser{}, ErrInvalidAdminID
	}

	adminEmail := normalizeEmail(email)
	if !isValidEmail(adminEmail) {
		return domain.AdminUser{}, ErrInvalidAdminEmail
	}

	adminRole, err := parseAdminRole(role)
	if err != nil {
		return domain.AdminUser{}, err
	}

	admin, err := s.repo.UpdateAdmin(ctx, adminID, adminEmail, adminRole, isActive)
	if err != nil {
		return domain.AdminUser{}, fmt.Errorf("update admin: %w", err)
	}
	return admin, nil
}

func (s *AdminService) DeleteAdmin(ctx context.Context, id string) error {
	adminID := strings.TrimSpace(id)
	if adminID == "" {
		return ErrInvalidAdminID
	}

	if err := s.repo.DeleteAdmin(ctx, adminID); err != nil {
		return fmt.Errorf("delete admin: %w", err)
	}
	return nil
}

func parseAdminRole(raw string) (domain.AdminRole, error) {
	switch domain.AdminRole(strings.TrimSpace(strings.ToLower(raw))) {
	case domain.AdminRoleOwner:
		return domain.AdminRoleOwner, nil
	case domain.AdminRoleEditor:
		return domain.AdminRoleEditor, nil
	case domain.AdminRoleViewer:
		return domain.AdminRoleViewer, nil
	default:
		return "", ErrInvalidAdminRole
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@")
}
