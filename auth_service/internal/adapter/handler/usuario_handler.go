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

type UsuarioHandler struct {
	usuarioService port.UsuarioService
}

func (u UsuarioHandler) DeshabilitarUsuario(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	err = u.usuarioService.DeshabilitarUsuario(c.UserContext(), &usuarioId)
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

func (u UsuarioHandler) HabilitarUsuario(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	err = u.usuarioService.HabilitarUsuario(c.UserContext(), &usuarioId)
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

func (u UsuarioHandler) RegistrarUsuario(c *fiber.Ctx) error {
	var usuarioRequest domain.UsuarioRequest
	if err := c.BodyParser(&usuarioRequest); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	id, err := u.usuarioService.RegistrarUsuario(c.UserContext(), &usuarioRequest)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.UsuarioResponse{Id: *id}, "Usuario creado correctamente"))
}

func (u UsuarioHandler) ObtenerUsuarioById(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	usuario, err := u.usuarioService.ObtenerUsuarioById(c.UserContext(), &usuarioId)
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

func (u UsuarioHandler) ObtenerListaUsuarios(c *fiber.Ctx) error {
	list, err := u.usuarioService.ObtenerListaUsuarios(c.UserContext())
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

func NewUsuarioHandler(usuarioService port.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{usuarioService: usuarioService}
}

var _ port.UsuarioHandler = (*UsuarioHandler)(nil)
