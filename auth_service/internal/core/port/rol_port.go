package port

import (
	"context"
	"multiroom/auth-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type RolRepository interface {
	RegistrarRol(ctx context.Context, request *domain.RolRequest) (*int, error)
	ModificarRolById(ctx context.Context, id *int, request *domain.RolRequest) error
	ListarRoles(ctx context.Context, filtros map[string]string) (*[]domain.RolInfo, error)
	ObtenerRolById(ctx context.Context, id *int) (*domain.Rol, error)
}

type RolService interface {
	RegistrarRol(ctx context.Context, request *domain.RolRequest) (*int, error)
	ModificarRolById(ctx context.Context, id *int, request *domain.RolRequest) error
	ListarRoles(ctx context.Context, filtros map[string]string) (*[]domain.RolInfo, error)
	ObtenerRolById(ctx context.Context, id *int) (*domain.Rol, error)
}

type RolHandler interface {
	RegistrarRol(c *fiber.Ctx) error
	ModificarRolById(c *fiber.Ctx) error
	ListarRoles(c *fiber.Ctx) error
	ObtenerRolById(c *fiber.Ctx) error
}
