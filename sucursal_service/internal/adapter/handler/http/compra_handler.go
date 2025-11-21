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

type CompraHandler struct {
	compraService port.CompraService
}

func (c2 CompraHandler) ListarCompras(c *fiber.Ctx) error {
	compras, err := c2.compraService.ListarCompras(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(compras)
}

func (c2 CompraHandler) ObtenerCompraById(c *fiber.Ctx) error {
	compraId, err := c.ParamsInt("compraId", 0)
	if err != nil || compraId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la compra debe ser un número válido mayor a 0"))
	}
	compra, err := c2.compraService.ObtenerCompraById(c.UserContext(), &compraId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(compra)
}

func (c2 CompraHandler) RegistrarOrdenCompra(c *fiber.Ctx) error {

	var request domain.CompraRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	compraId, err := c2.compraService.RegistrarOrdenCompra(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessageData(domain.CompraId{Id: *compraId}, "Orden de compra registrada correctamente"))
}

func (c2 CompraHandler) ModificarOrdenCompra(c *fiber.Ctx) error {
	compraId, err := c.ParamsInt("compraId", 0)
	if err != nil || compraId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la compra debe ser un número válido mayor a 0"))
	}
	var request domain.CompraRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	err = c2.compraService.ModificarOrdenCompra(c.UserContext(), &compraId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Orden de compra modificada correctamente"))
}

func (c2 CompraHandler) ConfirmarRecepcionCompra(c *fiber.Ctx) error {
	compraId, err := c.ParamsInt("compraId", 0)
	if err != nil || compraId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la compra debe ser un número válido mayor a 0"))
	}
	err = c2.compraService.ConfirmarRecepcionCompra(c.UserContext(), &compraId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Compra recepcionada correctamente"))
}

func NewCompraHandler(compraService port.CompraService) *CompraHandler {
	return &CompraHandler{compraService: compraService}
}

var _ port.CompraHandler = (*CompraHandler)(nil)
