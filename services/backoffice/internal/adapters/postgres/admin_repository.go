package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"raiiaa.dev/services/backoffice/internal/domain"
	"raiiaa.dev/services/backoffice/internal/ports"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) CreateAdmin(ctx context.Context, email string, role domain.AdminRole, isActive bool) (domain.AdminUser, error) {
	query := `
		INSERT INTO admins (email, role, is_active)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, role, is_active, created_at
	`

	admin := domain.AdminUser{}
	err := r.pool.QueryRow(ctx, query, email, role, isActive).
		Scan(&admin.ID, &admin.Email, &admin.Role, &admin.IsActive, &admin.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.AdminUser{}, ports.ErrAdminAlreadyExist
		}
		return domain.AdminUser{}, fmt.Errorf("insert admin: %w", err)
	}

	return admin, nil
}

func (r *AdminRepository) GetAdminByID(ctx context.Context, id string) (domain.AdminUser, error) {
	query := `
		SELECT id::text, email, role, is_active, created_at
		FROM admins
		WHERE id = $1
	`

	admin := domain.AdminUser{}
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&admin.ID, &admin.Email, &admin.Role, &admin.IsActive, &admin.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AdminUser{}, ports.ErrAdminNotFound
		}
		return domain.AdminUser{}, fmt.Errorf("get admin by id: %w", err)
	}

	return admin, nil
}

func (r *AdminRepository) ListAdmins(ctx context.Context) ([]domain.AdminUser, error) {
	query := `
		SELECT id::text, email, role, is_active, created_at
		FROM admins
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	defer rows.Close()

	admins := make([]domain.AdminUser, 0)
	for rows.Next() {
		admin := domain.AdminUser{}
		if err := rows.Scan(&admin.ID, &admin.Email, &admin.Role, &admin.IsActive, &admin.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan admin: %w", err)
		}
		admins = append(admins, admin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admins: %w", err)
	}

	return admins, nil
}

func (r *AdminRepository) UpdateAdmin(ctx context.Context, id, email string, role domain.AdminRole, isActive bool) (domain.AdminUser, error) {
	query := `
		UPDATE admins
		SET email = $2, role = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id::text, email, role, is_active, created_at
	`

	admin := domain.AdminUser{}
	err := r.pool.QueryRow(ctx, query, id, email, role, isActive).
		Scan(&admin.ID, &admin.Email, &admin.Role, &admin.IsActive, &admin.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.AdminUser{}, ports.ErrAdminAlreadyExist
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AdminUser{}, ports.ErrAdminNotFound
		}
		return domain.AdminUser{}, fmt.Errorf("update admin: %w", err)
	}

	return admin, nil
}

func (r *AdminRepository) DeleteAdmin(ctx context.Context, id string) error {
	commandTag, err := r.pool.Exec(ctx, `DELETE FROM admins WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete admin: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ports.ErrAdminNotFound
	}
	return nil
}
