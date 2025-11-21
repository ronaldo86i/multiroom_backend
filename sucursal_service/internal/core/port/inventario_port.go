package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type InventarioRepository interface {
	ListarInventario(ctx context.Context, filtros map[string]string) (*[]domain.Inventario, error)
	RegistrarAjusteConDetalle(ctx context.Context, request *domain.AjusteInventarioRequest) (*int, error)
	RegistrarTransferencia(ctx context.Context, request *domain.TransferenciaRequest) (*int, error)
	ListarTransferencias(ctx context.Context, filtros map[string]string) (*[]domain.TransferenciaInventarioInfo, error)
	ListarAjustes(ctx context.Context, filtros map[string]string) (*[]domain.AjusteInventarioInfo, error)
	ObtenerAjusteById(ctx context.Context, id *int) (*domain.AjusteInventario, error)
	ObtenerTransferenciaById(ctx context.Context, id *int) (*domain.TransferenciaInventario, error)
}

type InventarioService interface {
	ListarInventario(ctx context.Context, filtros map[string]string) (*[]domain.Inventario, error)
	RegistrarAjusteConDetalle(ctx context.Context, request *domain.AjusteInventarioRequest) (*int, error)
	RegistrarTransferencia(ctx context.Context, request *domain.TransferenciaRequest) (*int, error)
	ListarTransferencias(ctx context.Context, filtros map[string]string) (*[]domain.TransferenciaInventarioInfo, error)
	ListarAjustes(ctx context.Context, filtros map[string]string) (*[]domain.AjusteInventarioInfo, error)
	ObtenerAjusteById(ctx context.Context, id *int) (*domain.AjusteInventario, error)
	ObtenerTransferenciaById(ctx context.Context, id *int) (*domain.TransferenciaInventario, error)
}

type InventarioHandler interface {
	ListarInventario(c *fiber.Ctx) error
	RegistrarAjusteConDetalle(c *fiber.Ctx) error
	RegistrarTransferencia(c *fiber.Ctx) error
	ListarTransferencias(c *fiber.Ctx) error
	ListarAjustes(c *fiber.Ctx) error
	ObtenerAjusteById(c *fiber.Ctx) error
	ObtenerTransferenciaById(c *fiber.Ctx) error
}
