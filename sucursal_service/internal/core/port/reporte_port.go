package port

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/johnfercher/maroto/v2/pkg/core"
)

type ReporteService interface {
	ComprobantePDFVentaById(ctx context.Context, ventaId *int) (core.Document, error)
	ReportePDFVentas(ctx context.Context, filtros map[string]string) (core.Document, error)
	ReportePDFProductosVendidos(ctx context.Context, filtros map[string]string) (core.Document, error)
}

type ReporteHandler interface {
	ComprobantePDFVentaById(c *fiber.Ctx) error
	ReportePDFVentas(c *fiber.Ctx) error
	ReportePDFProductosVendidos(c *fiber.Ctx) error
}
