package port

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type MetodoPagoRepository interface {
	ListarMetodosPago(ctx context.Context, filtros map[string]string) (*[]domain.MetodoPago, error)
}

type MetodoPagoService interface {
	ListarMetodosPago(ctx context.Context, filtros map[string]string) (*[]domain.MetodoPago, error)
}

type MetodoPagoHandler interface {
	ListarMetodosPago(c *fiber.Ctx) error
}
