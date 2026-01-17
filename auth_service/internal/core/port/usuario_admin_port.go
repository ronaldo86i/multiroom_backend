package port

import (
	"context"
	"multiroom/auth-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type UsuarioAdminRepository interface {
	ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error)
	ObtenerUsuarioAdminById(ctx context.Context, id *int) (*domain.UsuarioAdmin, error)
	RegistrarUsuarioAdmin(ctx context.Context, request *domain.UsuarioAdminRequest) (*int, error)
	ModificarUsuarioAdminById(ctx context.Context, id *int, request *domain.UsuarioAdminRequest) error
	ListarUsuariosAdmin(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioAdminInfo, error)
}

type UsuarioAdminService interface {
	ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error)
	ObtenerUsuarioAdminById(ctx context.Context, id *int) (*domain.UsuarioAdmin, error)
	RegistrarUsuarioAdmin(ctx context.Context, request *domain.UsuarioAdminRequest) (*int, error)
	ModificarUsuarioAdminById(ctx context.Context, id *int, request *domain.UsuarioAdminRequest) error
	ListarUsuariosAdmin(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioAdminInfo, error)
}

type UsuarioAdminHandler interface {
	ObtenerUsuarioAdminById(c *fiber.Ctx) error
	RegistrarUsuarioAdmin(c *fiber.Ctx) error
	ModificarUsuarioAdminById(c *fiber.Ctx) error
	ListarUsuariosAdmin(c *fiber.Ctx) error
}
