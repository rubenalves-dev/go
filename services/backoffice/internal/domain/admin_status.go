package domain

import "time"

type AdminStatus struct {
	Service      string    `json:"service"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
	TotalAdmins  int       `json:"total_admins"`
	ActiveAdmins int       `json:"active_admins"`
}
