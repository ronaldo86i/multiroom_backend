package port

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
)

type AppVersionRepository interface {
	ObtenerListaVersiones(ctx context.Context, filtros map[string]string) (*[]domain.AppVersion, error)
	ObtenerUltimaVersion(ctx context.Context, q *domain.AppLastVersionQuery) (*domain.AppVersion, error)
	RegistrarApp(ctx context.Context, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarVersion(ctx context.Context, id *int, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) error
	ObtenerVersion(ctx context.Context, id *int) (*domain.AppVersion, error)
}

type AppVersionService interface {
	ObtenerListaVersiones(ctx context.Context, filtros map[string]string) (*[]domain.AppVersion, error)
	ObtenerUltimaVersion(ctx context.Context, q *domain.AppLastVersionQuery) (*domain.AppVersion, error)
	RegistrarApp(ctx context.Context, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) (*int, error)
	ModificarVersion(ctx context.Context, id *int, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) error
	ObtenerVersion(ctx context.Context, id *int) (*domain.AppVersion, error)
}

type AppVersionHandler interface {
	ObtenerListaVersiones(c *fiber.Ctx) error
	ObtenerUltimaVersion(c *fiber.Ctx) error
	RegistrarApp(c *fiber.Ctx) error
	ModificarVersion(c *fiber.Ctx) error
	ObtenerVersion(c *fiber.Ctx) error
}
