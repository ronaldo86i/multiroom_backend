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

type InventarioHandler struct {
	inventarioService port.InventarioService
}

func (i InventarioHandler) ObtenerTransferenciaById(c *fiber.Ctx) error {
	transferenciaId, err := c.ParamsInt("transferenciaId", 0)
	if err != nil || transferenciaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la transferencia debe ser un número válido mayor a 0"))
	}
	transferencia, err := i.inventarioService.ObtenerTransferenciaById(c.UserContext(), &transferenciaId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(transferencia)
}

func (i InventarioHandler) ObtenerAjusteById(c *fiber.Ctx) error {
	ajusteId, err := c.ParamsInt("ajusteId", 0)
	if err != nil || ajusteId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del ajuste de inventario debe ser un número válido mayor a 0"))
	}
	ajuste, err := i.inventarioService.ObtenerAjusteById(c.UserContext(), &ajusteId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(ajuste)
}

func (i InventarioHandler) ListarTransferencias(c *fiber.Ctx) error {
	list, err := i.inventarioService.ListarTransferencias(c.UserContext(), c.Queries())
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

func (i InventarioHandler) ListarAjustes(c *fiber.Ctx) error {
	list, err := i.inventarioService.ListarAjustes(c.UserContext(), c.Queries())
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

func (i InventarioHandler) RegistrarAjusteConDetalle(c *fiber.Ctx) error {
	var request domain.AjusteInventarioRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	id, err := i.inventarioService.RegistrarAjusteConDetalle(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessageData(domain.AjusteId{Id: *id}, "Ajuste registrado correctamente"))
}

func (i InventarioHandler) RegistrarTransferencia(c *fiber.Ctx) error {
	var request domain.TransferenciaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	id, err := i.inventarioService.RegistrarTransferencia(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessageData(domain.TransferenciaId{Id: *id}, "Transferencia registrada correctamente"))
}

func (i InventarioHandler) ListarInventario(c *fiber.Ctx) error {
	list, err := i.inventarioService.ListarInventario(c.UserContext(), c.Queries())
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

func NewInventarioHandler(inventarioService port.InventarioService) InventarioHandler {
	return InventarioHandler{inventarioService: inventarioService}
}

var _ port.InventarioHandler = (*InventarioHandler)(nil)
