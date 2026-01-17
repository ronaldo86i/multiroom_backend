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

type RolHandler struct {
	rolService port.RolService
}

func (r RolHandler) RegistrarRol(c *fiber.Ctx) error {
	var request domain.RolRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	rolId, err := r.rolService.RegistrarRol(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessageData(domain.RolId{Id: *rolId}, "Rol registrado correctamente"))
}

func (r RolHandler) ModificarRolById(c *fiber.Ctx) error {
	rolId, err := c.ParamsInt("rolId", 0)
	if err != nil || rolId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del rol debe ser un número válido mayor a 0"))
	}
	var request domain.RolRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err = r.rolService.ModificarRolById(c.UserContext(), &rolId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Rol modificado correctamente"))
}

func (r RolHandler) ListarRoles(c *fiber.Ctx) error {
	list, err := r.rolService.ListarRoles(c.UserContext(), c.Queries())
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

func (r RolHandler) ObtenerRolById(c *fiber.Ctx) error {
	rolId, err := c.ParamsInt("rolId", 0)
	if err != nil || rolId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del rol debe ser un número válido mayor a 0"))
	}
	rol, err := r.rolService.ObtenerRolById(c.UserContext(), &rolId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(rol)
}

func NewRolHandler(rolService port.RolService) *RolHandler {
	return &RolHandler{rolService: rolService}
}

var _ port.RolHandler = (*RolHandler)(nil)
