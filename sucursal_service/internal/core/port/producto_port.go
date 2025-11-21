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
	ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.Producto, error)
	ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error)
}

type ProductoService interface {
	RegistrarProducto(ctx context.Context, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarProductoById(ctx context.Context, productoId *int, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) error
	ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.Producto, error)
	ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error)
}

type ProductoHandler interface {
	RegistrarProducto(c *fiber.Ctx) error
	ModificarProductoById(c *fiber.Ctx) error
	ListarProductos(c *fiber.Ctx) error
	ObtenerProductoById(c *fiber.Ctx) error
}
