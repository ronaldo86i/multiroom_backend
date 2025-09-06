package port

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"multiroom/sucursal-service/internal/core/domain"
)

type SucursalRepository interface {
	RegistrarSucursal(ctx context.Context, request *domain.SucursalRequest) (*int, error)
	ModificarSucursal(ctx context.Context, id *int, request *domain.SucursalRequest) error
	ObtenerSucursalById(ctx context.Context, id *int) (*domain.SucursalDetail, error)
	ObtenerListaSucursales(ctx context.Context, filtros map[string]string) (*[]domain.SucursalInfo, error)
	HabilitarSucursal(ctx context.Context, id *int) error
	DeshabilitarSucursal(ctx context.Context, id *int) error
}

type SucursalService interface {
	RegistrarSucursal(ctx context.Context, request *domain.SucursalRequest) (*int, error)
	ModificarSucursal(ctx context.Context, id *int, request *domain.SucursalRequest) error
	ObtenerSucursalById(ctx context.Context, id *int) (*domain.SucursalDetail, error)
	ObtenerListaSucursales(ctx context.Context, filtros map[string]string) (*[]domain.SucursalInfo, error)
	HabilitarSucursal(ctx context.Context, id *int) error
	DeshabilitarSucursal(ctx context.Context, id *int) error
}

type SucursalHandler interface {
	RegistrarSucursal(c *fiber.Ctx) error
	ModificarSucursal(c *fiber.Ctx) error
	ObtenerSucursalById(c *fiber.Ctx) error
	ObtenerListaSucursales(c *fiber.Ctx) error
	HabilitarSucursal(c *fiber.Ctx) error
	DeshabilitarSucursal(c *fiber.Ctx) error
}
