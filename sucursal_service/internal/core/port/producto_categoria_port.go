package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type ProductoCategoriaRepository interface {
	RegistrarCategoria(ctx context.Context, request *domain.ProductoCategoriaRequest) (*int, error)
	ModificarCategoriaById(ctx context.Context, id *int, request *domain.ProductoCategoriaRequest) error
	ListarCategorias(ctx context.Context, filtros map[string]string) (*[]domain.ProductoCategoria, error)
	ObtenerCategoriaById(ctx context.Context, id *int) (*domain.ProductoCategoria, error)
}

type ProductoCategoriaService interface {
	RegistrarCategoria(ctx context.Context, request *domain.ProductoCategoriaRequest) (*int, error)
	ModificarCategoriaById(ctx context.Context, id *int, request *domain.ProductoCategoriaRequest) error
	ListarCategorias(ctx context.Context, filtros map[string]string) (*[]domain.ProductoCategoria, error)
	ObtenerCategoriaById(ctx context.Context, id *int) (*domain.ProductoCategoria, error)
}

type ProductoCategoriaHandler interface {
	RegistrarCategoria(c *fiber.Ctx) error
	ModificarCategoriaById(c *fiber.Ctx) error
	ListarCategorias(c *fiber.Ctx) error
	ObtenerCategoriaById(c *fiber.Ctx) error
}
