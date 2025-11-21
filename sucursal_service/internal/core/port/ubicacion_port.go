package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type UbicacionRepository interface {
	RegistrarUbicacion(ctx context.Context, request *domain.UbicacionRequest) (*int, error)
	ModificarUbicacionById(ctx context.Context, id *int, request *domain.UbicacionRequest) error
	ListarUbicaciones(ctx context.Context, filtros map[string]string) (*[]domain.Ubicacion, error)
	ObtenerUbicacionById(ctx context.Context, id *int) (*domain.Ubicacion, error)
	HabilitarUbicacion(ctx context.Context, id *int) error
	DeshabilitarUbicacion(ctx context.Context, id *int) error
}

type UbicacionService interface {
	RegistrarUbicacion(ctx context.Context, request *domain.UbicacionRequest) (*int, error)
	ModificarUbicacionById(ctx context.Context, id *int, request *domain.UbicacionRequest) error
	ListarUbicaciones(ctx context.Context, filtros map[string]string) (*[]domain.Ubicacion, error)
	ObtenerUbicacionById(ctx context.Context, id *int) (*domain.Ubicacion, error)
	HabilitarUbicacion(ctx context.Context, id *int) error
	DeshabilitarUbicacion(ctx context.Context, id *int) error
}

type UbicacionHandler interface {
	RegistrarUbicacion(c *fiber.Ctx) error
	ModificarUbicacionById(c *fiber.Ctx) error
	ListarUbicaciones(c *fiber.Ctx) error
	ObtenerUbicacionById(c *fiber.Ctx) error
	HabilitarUbicacion(c *fiber.Ctx) error
	DeshabilitarUbicacion(c *fiber.Ctx) error
}
