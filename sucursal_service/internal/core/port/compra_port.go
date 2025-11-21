package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type CompraRepository interface {
	ListarCompras(ctx context.Context, filtros map[string]string) (*[]domain.CompraInfo, error)
	ObtenerCompraById(ctx context.Context, id *int) (*domain.Compra, error)
	RegistrarOrdenCompra(ctx context.Context, request *domain.CompraRequest) (*int, error)
	ModificarOrdenCompra(ctx context.Context, id *int, request *domain.CompraRequest) error
	ConfirmarRecepcionCompra(ctx context.Context, id *int) error
}

type CompraService interface {
	ListarCompras(ctx context.Context, filtros map[string]string) (*[]domain.CompraInfo, error)
	ObtenerCompraById(ctx context.Context, id *int) (*domain.Compra, error)
	RegistrarOrdenCompra(ctx context.Context, request *domain.CompraRequest) (*int, error)
	ModificarOrdenCompra(ctx context.Context, id *int, request *domain.CompraRequest) error
	ConfirmarRecepcionCompra(ctx context.Context, id *int) error
}

type CompraHandler interface {
	ListarCompras(c *fiber.Ctx) error
	ObtenerCompraById(c *fiber.Ctx) error
	RegistrarOrdenCompra(c *fiber.Ctx) error
	ModificarOrdenCompra(c *fiber.Ctx) error
	ConfirmarRecepcionCompra(c *fiber.Ctx) error
}
