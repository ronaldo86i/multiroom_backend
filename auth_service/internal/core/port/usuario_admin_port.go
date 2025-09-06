package port

import (
	"context"
	"multiroom/auth-service/internal/core/domain"
)

type UsuarioAdminRepository interface {
	ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error)
	ObtenerUsuarioAdminById(ctx context.Context, id *int) (*domain.UsuarioAdmin, error)
}

type UsuarioAdminService interface {
	ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error)
	ObtenerUsuarioAdminById(ctx context.Context, id *int) (*domain.UsuarioAdmin, error)
}
