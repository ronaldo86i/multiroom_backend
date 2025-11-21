package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type ProveedorRepository interface {
	RegistrarProveedor(ctx context.Context, request *domain.ProveedorRequest) (*int, error)
	ModificarProveedor(ctx context.Context, id *int, request *domain.ProveedorRequest) error
	ListarProveedores(ctx context.Context, filtros map[string]string) (*[]domain.Proveedor, error)
	ObtenerProveedorById(ctx context.Context, id *int) (*domain.Proveedor, error)
}

type ProveedorService interface {
	RegistrarProveedor(ctx context.Context, request *domain.ProveedorRequest) (*int, error)
	ModificarProveedor(ctx context.Context, id *int, request *domain.ProveedorRequest) error
	ListarProveedores(ctx context.Context, filtros map[string]string) (*[]domain.Proveedor, error)
	ObtenerProveedorById(ctx context.Context, id *int) (*domain.Proveedor, error)
}

type ProveedorHandler interface {
	RegistrarProveedor(c *fiber.Ctx) error
	ModificarProveedor(c *fiber.Ctx) error
	ListarProveedores(c *fiber.Ctx) error
	ObtenerProveedorById(c *fiber.Ctx) error
}
