package domain

import "time"

type AdminRole string

const (
	AdminRoleOwner  AdminRole = "owner"
	AdminRoleEditor AdminRole = "editor"
	AdminRoleViewer AdminRole = "viewer"
)

type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      AdminRole `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
