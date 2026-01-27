package http

import (
	"errors"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ReporteHandler struct {
	reporteService port.ReporteService
}

func (r ReporteHandler) ReportePDFProductosVendidos(c *fiber.Ctx) error {
	doc, err := r.reporteService.ReportePDFProductosVendidos(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}

	c.Response().Header.Set("Content-Type", "application/pdf")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="reporte-productos-vendidos-%s.pdf"`, time.Now().Format("2006-01-02 03-04-05")))
	c.Response().Header.Set("Content-Transfer-Encoding", "binary")

	return c.Send(doc.GetBytes())
}

func (r ReporteHandler) ReportePDFVentas(c *fiber.Ctx) error {
	doc, err := r.reporteService.ReportePDFVentas(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}

	c.Response().Header.Set("Content-Type", "application/pdf")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="reporte-ventas-%s.pdf"`, time.Now().String()))
	c.Response().Header.Set("Content-Transfer-Encoding", "binary")

	return c.Send(doc.GetBytes())
}

func (r ReporteHandler) ComprobantePDFVentaById(c *fiber.Ctx) error {
	ventaId, err := c.ParamsInt("ventaId", 0)
	if err != nil || ventaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la venta debe ser un número válido mayor a 0"))
	}
	doc, err := r.reporteService.ComprobantePDFVentaById(c.UserContext(), &ventaId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}

	c.Response().Header.Set("Content-Type", "application/pdf")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="comprobante-venta-%d.pdf"`, ventaId))
	c.Response().Header.Set("Content-Transfer-Encoding", "binary")

	return c.Send(doc.GetBytes())
}

func NewReporteHandler(reporteService port.ReporteService) *ReporteHandler {
	return &ReporteHandler{reporteService: reporteService}
}

var _ port.ReporteHandler = (*ReporteHandler)(nil)
