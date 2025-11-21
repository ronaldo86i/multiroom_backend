package port

import (
	"context"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type PaisRepository interface {
	RegistrarPais(ctx context.Context, request *domain.PaisRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarPais(ctx context.Context, id *int, request *domain.PaisRequest, fileHeader *multipart.FileHeader) error
	ObtenerPaisById(ctx context.Context, id *int) (*domain.PaisDetail, error)
	ObtenerListaPaises(ctx context.Context, filtros map[string]string) (*[]domain.PaisInfo, error)
	HabilitarPaisById(ctx context.Context, id *int) error
	DeshabilitarPaisById(ctx context.Context, id *int) error
}

type PaisService interface {
	RegistrarPais(ctx context.Context, request *domain.PaisRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarPais(ctx context.Context, id *int, request *domain.PaisRequest, fileHeader *multipart.FileHeader) error
	ObtenerPaisById(ctx context.Context, id *int) (*domain.PaisDetail, error)
	ObtenerListaPaises(ctx context.Context, filtros map[string]string) (*[]domain.PaisInfo, error)
	HabilitarPaisById(ctx context.Context, id *int) error
	DeshabilitarPaisById(ctx context.Context, id *int) error
}

type PaisHandler interface {
	RegistrarPais(c *fiber.Ctx) error
	ModificarPais(c *fiber.Ctx) error
	ObtenerPaisById(c *fiber.Ctx) error
	ObtenerListaPaises(c *fiber.Ctx) error
	HabilitarPaisById(c *fiber.Ctx) error
	DeshabilitarPaisById(c *fiber.Ctx) error
}
