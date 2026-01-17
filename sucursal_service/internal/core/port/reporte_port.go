package port

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/johnfercher/maroto/v2/pkg/core"
)

type ReporteService interface {
	ComprobantePDFVentaById(ctx context.Context, ventaId *int) (core.Document, error)
}

type ReporteHandler interface {
	ComprobantePDFVentaById(c *fiber.Ctx) error
}
