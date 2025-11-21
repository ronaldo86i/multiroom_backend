package http

import (
	"errors"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/domain/datatype"
	"multiroom/dispositivo-service/internal/core/port"
	"multiroom/dispositivo-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type ClienteHandler struct {
	clienteService port.ClienteService
}

func (c2 ClienteHandler) EliminarClienteById(c *fiber.Ctx) error {
	clienteId, err := c.ParamsInt("clienteId", 0)
	if err != nil || clienteId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del cliente debe ser un número válido mayor a 0"))
	}
	err = c2.clienteService.EliminarClienteById(c.UserContext(), &clienteId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessage("Cliente eliminado correctamente"))
}

func (c2 ClienteHandler) RegistrarCliente(c *fiber.Ctx) error {
	var request domain.ClienteRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	clienteId, err := c2.clienteService.RegistrarCliente(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.ClienteIdResponse{Id: *clienteId}, "Cliente registrado correctamente"))
}

func (c2 ClienteHandler) ModificarCliente(c *fiber.Ctx) error {
	var request domain.ClienteRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	clienteId, err := c.ParamsInt("clienteId", 0)
	if err != nil || clienteId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del cliente debe ser un número válido mayor a 0"))
	}
	err = c2.clienteService.ModificarCliente(c.UserContext(), &clienteId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessage("Cliente modificado correctamente"))
}

func (c2 ClienteHandler) ObtenerListaClientes(c *fiber.Ctx) error {
	list, err := c2.clienteService.ObtenerListaClientes(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(list)
}

func (c2 ClienteHandler) ObtenerClienteDetailById(c *fiber.Ctx) error {
	clienteId, err := c.ParamsInt("clienteId", 0)
	if err != nil || clienteId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del cliente debe ser un número válido mayor a 0"))
	}
	cliente, err := c2.clienteService.ObtenerClienteDetailById(c.UserContext(), &clienteId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(cliente)
}

func (c2 ClienteHandler) HabilitarCliente(c *fiber.Ctx) error {
	clienteId, err := c.ParamsInt("clienteId", 0)
	if err != nil || clienteId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del cliente debe ser un número válido mayor a 0"))
	}
	err = c2.clienteService.HabilitarCliente(c.UserContext(), &clienteId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessage("Cliente actualizado correctamente"))
}

func (c2 ClienteHandler) DeshabilitarCliente(c *fiber.Ctx) error {
	clienteId, err := c.ParamsInt("clienteId", 0)
	if err != nil || clienteId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del cliente debe ser un número válido mayor a 0"))
	}
	err = c2.clienteService.DeshabilitarCliente(c.UserContext(), &clienteId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessage("Cliente actualizado correctamente"))
}

func NewClienteHandler(clienteService port.ClienteService) *ClienteHandler {
	return &ClienteHandler{clienteService: clienteService}
}

var _ port.ClienteHandler = (*ClienteHandler)(nil)
