package handler

import (
	"errors"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
	"multiroom/auth-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type UsuarioAdminHandler struct {
	usuarioAdminService port.UsuarioAdminService
}

func (u UsuarioAdminHandler) ObtenerUsuarioAdminById(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	usuario, err := u.usuarioAdminService.ObtenerUsuarioAdminById(c.UserContext(), &usuarioId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(usuario)
}

func (u UsuarioAdminHandler) RegistrarUsuarioAdmin(c *fiber.Ctx) error {
	var request domain.UsuarioAdminRequest
	if err := c.BodyParser(&request); err != nil {
		log.Println("Error al escanear body:", err.Error())
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	usuarioId, err := u.usuarioAdminService.RegistrarUsuarioAdmin(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessageData(domain.UsuarioAdminId{Id: *usuarioId}, "Usuario registrado correctamente"))

}

func (u UsuarioAdminHandler) ModificarUsuarioAdminById(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}
	var request domain.UsuarioAdminRequest
	if err := c.BodyParser(&request); err != nil {
		log.Println("Error al escanear body:", err.Error())
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err = u.usuarioAdminService.ModificarUsuarioAdminById(c.UserContext(), &usuarioId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessage("Usuario modificado correctamente"))
}

func (u UsuarioAdminHandler) ListarUsuariosAdmin(c *fiber.Ctx) error {
	list, err := u.usuarioAdminService.ListarUsuariosAdmin(c.UserContext(), c.Queries())
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

func NewUsuarioAdminHandler(usuarioAdminService port.UsuarioAdminService) *UsuarioAdminHandler {
	return &UsuarioAdminHandler{usuarioAdminService: usuarioAdminService}
}

var _ port.UsuarioAdminHandler = (*UsuarioAdminHandler)(nil)
