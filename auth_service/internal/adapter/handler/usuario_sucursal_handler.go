package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
	"multiroom/auth-service/internal/core/util"
	"net/http"
)

type UsuarioSucursalHandler struct {
	usuarioSucursalService port.UsuarioSucursalService
}

func (u UsuarioSucursalHandler) RegistrarUsuarioSucursal(c *fiber.Ctx) error {
	var usuarioRequest domain.UsuarioSucursalRequest
	if err := c.BodyParser(&usuarioRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	id, err := u.usuarioSucursalService.RegistrarUsuarioSucursal(c.UserContext(), &usuarioRequest)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.UsuarioResponse{Id: *id}, "Usuario registrado correctamente"))
}

func (u UsuarioSucursalHandler) ModificarUsuarioSucursal(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	var usuarioRequest domain.UsuarioSucursalRequest
	if err := c.BodyParser(&usuarioRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err = u.usuarioSucursalService.ModificarUsuarioSucursal(c.UserContext(), &usuarioId, &usuarioRequest)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Usuario modificado correctamente"))
}

func (u UsuarioSucursalHandler) ObtenerListaUsuariosSucursal(c *fiber.Ctx) error {
	list, err := u.usuarioSucursalService.ObtenerListaUsuariosSucursal(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(list)
}

func (u UsuarioSucursalHandler) ObtenerUsuarioSucursalById(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	usuario, err := u.usuarioSucursalService.ObtenerUsuarioSucursalById(c.UserContext(), &usuarioId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(usuario)
}

func (u UsuarioSucursalHandler) HabilitarUsuarioSucursal(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	err = u.usuarioSucursalService.HabilitarUsuarioSucursal(c.UserContext(), &usuarioId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Usuario habilitado correctamente"))
}

func (u UsuarioSucursalHandler) DeshabilitarUsuarioSucursal(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	err = u.usuarioSucursalService.DeshabilitarUsuarioSucursal(c.UserContext(), &usuarioId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Usuario deshabilitado correctamente"))
}

func NewUsuarioSucursalHandler(usuarioSucursalService port.UsuarioSucursalService) *UsuarioSucursalHandler {
	return &UsuarioSucursalHandler{usuarioSucursalService: usuarioSucursalService}
}

var _ port.UsuarioSucursalHandler = (*UsuarioSucursalHandler)(nil)
