package port

import (
	"context"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type ProductoRepository interface {
	RegistrarProducto(ctx context.Context, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarProductoById(ctx context.Context, productoId *int, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) error
	ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.ProductoInfo, error)
	ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error)
	ListarProductosMasVendidos(ctx context.Context, filtros map[string]string) (*[]domain.ProductoStat, error)
	HabilitarProductoById(ctx context.Context, productoId *int) error
	DeshabilitarProductoById(ctx context.Context, productoId *int) error
	ListarProductosPorSucursal(ctx context.Context, filtros map[string]string) (*[]domain.ProductoSucursalInfo, error)
	ObtenerProductoSucursalById(ctx context.Context, id *int) (*domain.ProductoSucursalInfo, error)
	ActualizarProductoSucursal(ctx context.Context, id *int, req *domain.ProductoSucursalUpdateRequest) error
}

type ProductoService interface {
	RegistrarProducto(ctx context.Context, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarProductoById(ctx context.Context, productoId *int, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) error
	ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.ProductoInfo, error)
	ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error)
	ListarProductosMasVendidos(ctx context.Context, filtros map[string]string) (*[]domain.ProductoStat, error)
	HabilitarProductoById(ctx context.Context, productoId *int) error
	DeshabilitarProductoById(ctx context.Context, productoId *int) error
	ListarProductosPorSucursal(ctx context.Context, filtros map[string]string) (*[]domain.ProductoSucursalInfo, error)
	ObtenerProductoSucursalById(ctx context.Context, id *int) (*domain.ProductoSucursalInfo, error)
	ActualizarProductoSucursal(ctx context.Context, id *int, req *domain.ProductoSucursalUpdateRequest) error
}

type ProductoHandler interface {
	RegistrarProducto(c *fiber.Ctx) error
	ModificarProductoById(c *fiber.Ctx) error
	ListarProductos(c *fiber.Ctx) error
	ObtenerProductoById(c *fiber.Ctx) error
	ListarProductosMasVendidos(c *fiber.Ctx) error
	HabilitarProductoById(c *fiber.Ctx) error
	DeshabilitarProductoById(c *fiber.Ctx) error
	ListarProductosPorSucursal(c *fiber.Ctx) error
	ObtenerProductoSucursalById(c *fiber.Ctx) error
	ActualizarProductoSucursal(c *fiber.Ctx) error
}
