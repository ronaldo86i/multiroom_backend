package port

import (
	"context"
	"multiroom/auth-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type PermisoRepository interface {
	ListarPermisos(ctx context.Context, filtros map[string]string) (*[]domain.Permiso, error)
	ObtenerPermisoById(ctx context.Context, id *int) (*domain.Permiso, error)
	RegistrarPermiso(ctx context.Context, request *domain.PermisoRequest) (*int, error)
	ModificarPermisoById(ctx context.Context, id *int, request *domain.PermisoRequest) error
}

type PermisoService interface {
	ListarPermisos(ctx context.Context, filtros map[string]string) (*[]domain.Permiso, error)
	ObtenerPermisoById(ctx context.Context, id *int) (*domain.Permiso, error)
	RegistrarPermiso(ctx context.Context, request *domain.PermisoRequest) (*int, error)
	ModificarPermisoById(ctx context.Context, id *int, request *domain.PermisoRequest) error
}

type PermisoHandler interface {
	ListarPermisos(c *fiber.Ctx) error
	ObtenerPermisoById(c *fiber.Ctx) error
	RegistrarPermiso(c *fiber.Ctx) error
	ModificarPermisoById(c *fiber.Ctx) error
}
