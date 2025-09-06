package port

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"multiroom/auth-service/internal/core/domain"
)

type UsuarioSucursalRepository interface {
	RegistrarUsuarioSucursal(ctx context.Context, request *domain.UsuarioSucursalRequest) (*int, error)
	ModificarUsuarioSucursal(ctx context.Context, id *int, request *domain.UsuarioSucursalRequest) error
	ObtenerListaUsuariosSucursal(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioSucursalInfo, error)
	ObtenerUsuarioSucursalById(ctx context.Context, id *int) (*domain.UsuarioSucursal, error)
	ObtenerUsuarioSucursalByUsername(ctx context.Context, username *string) (*domain.UsuarioSucursal, error)
	HabilitarUsuarioSucursal(ctx context.Context, id *int) error
	DeshabilitarUsuarioSucursal(ctx context.Context, id *int) error
}

type UsuarioSucursalService interface {
	RegistrarUsuarioSucursal(ctx context.Context, request *domain.UsuarioSucursalRequest) (*int, error)
	ModificarUsuarioSucursal(ctx context.Context, id *int, request *domain.UsuarioSucursalRequest) error
	ObtenerListaUsuariosSucursal(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioSucursalInfo, error)
	ObtenerUsuarioSucursalById(ctx context.Context, id *int) (*domain.UsuarioSucursal, error)
	ObtenerUsuarioSucursalByUsername(ctx context.Context, username *string) (*domain.UsuarioSucursal, error)
	HabilitarUsuarioSucursal(ctx context.Context, id *int) error
	DeshabilitarUsuarioSucursal(ctx context.Context, id *int) error
}

type UsuarioSucursalHandler interface {
	RegistrarUsuarioSucursal(c *fiber.Ctx) error
	ModificarUsuarioSucursal(c *fiber.Ctx) error
	ObtenerListaUsuariosSucursal(c *fiber.Ctx) error
	ObtenerUsuarioSucursalById(c *fiber.Ctx) error
	HabilitarUsuarioSucursal(c *fiber.Ctx) error
	DeshabilitarUsuarioSucursal(c *fiber.Ctx) error
}
