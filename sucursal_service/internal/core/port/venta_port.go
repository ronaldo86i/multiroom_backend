package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type VentaRepository interface {
	RegistrarVenta(ctx context.Context, request *domain.VentaRequest) (*int, error)
	AnularVentaById(ctx context.Context, id *int) error
	RegistrarPagoVenta(ctx context.Context, ventaId *int, request *domain.RegistrarPagosRequest) (*[]int, error)
	ObtenerVenta(ctx context.Context, id *int) (*domain.Venta, error)
	ListarVentas(ctx context.Context, filtros map[string]string) (*[]domain.VentaInfo, error)
	ListarProductosVentas(ctx context.Context, filtros map[string]string) (*[]domain.ProductoVentaStat, error)
}

type VentaService interface {
	RegistrarVenta(ctx context.Context, request *domain.VentaRequest) (*int, error)
	AnularVentaById(ctx context.Context, id *int) error
	RegistrarPagoVenta(ctx context.Context, ventaId *int, request *domain.RegistrarPagosRequest) (*[]int, error)
	ObtenerVenta(ctx context.Context, id *int) (*domain.Venta, error)
	ListarVentas(ctx context.Context, filtros map[string]string) (*[]domain.VentaInfo, error)
	ListarProductosVentas(ctx context.Context, filtros map[string]string) (*[]domain.ProductoVentaStat, error)
}

type VentaHandler interface {
	RegistrarVenta(c *fiber.Ctx) error
	AnularVentaById(c *fiber.Ctx) error
	RegistrarPagoVenta(c *fiber.Ctx) error
	ObtenerVenta(c *fiber.Ctx) error
	ListarVentas(c *fiber.Ctx) error
	ListarProductosVentas(c *fiber.Ctx) error
}
