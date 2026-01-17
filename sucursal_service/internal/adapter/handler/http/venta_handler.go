package http

import (
	"errors"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type VentaHandler struct {
	ventaService port.VentaService
}

func (v VentaHandler) ListarProductosVentas(c *fiber.Ctx) error {
	list, err := v.ventaService.ListarProductosVentas(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(list)
}

func (v VentaHandler) RegistrarVenta(c *fiber.Ctx) error {
	var request domain.VentaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	ventaId, err := v.ventaService.RegistrarVenta(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessageData(domain.VentaId{Id: *ventaId}, "Venta registrada correctamente"))
}

func (v VentaHandler) AnularVentaById(c *fiber.Ctx) error {
	ventaId, err := c.ParamsInt("ventaId", 0)
	if err != nil || ventaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la venta debe ser un número válido mayor a 0"))
	}
	err = v.ventaService.AnularVentaById(c.UserContext(), &ventaId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Venta anulada correctamente"))
}

func (v VentaHandler) RegistrarPagoVenta(c *fiber.Ctx) error {
	ventaId, err := c.ParamsInt("ventaId", 0)
	if err != nil || ventaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la venta debe ser un número válido mayor a 0"))
	}
	var request domain.RegistrarPagosRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	_, err = v.ventaService.RegistrarPagoVenta(c.UserContext(), &ventaId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Pagos registrados correctamente"))
}

func (v VentaHandler) ObtenerVenta(c *fiber.Ctx) error {
	ventaId, err := c.ParamsInt("ventaId", 0)
	if err != nil || ventaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la venta debe ser un número válido mayor a 0"))
	}
	venta, err := v.ventaService.ObtenerVenta(c.UserContext(), &ventaId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(venta)
}

func (v VentaHandler) ListarVentas(c *fiber.Ctx) error {
	list, err := v.ventaService.ListarVentas(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(list)
}

func NewVentaHandler(ventaService port.VentaService) *VentaHandler {
	return &VentaHandler{ventaService: ventaService}
}

var _ port.VentaHandler = (*VentaHandler)(nil)
