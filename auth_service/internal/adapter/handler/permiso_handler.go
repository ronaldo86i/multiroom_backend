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

type PermisoHandler struct {
	permisoService port.PermisoService
}

func (p PermisoHandler) ListarPermisos(c *fiber.Ctx) error {
	list, err := p.permisoService.ListarPermisos(c.UserContext(), c.Queries())
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

func (p PermisoHandler) ObtenerPermisoById(c *fiber.Ctx) error {
	permisoId, err := c.ParamsInt("permisoId", 0)
	if err != nil || permisoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del permiso debe ser un número válido mayor a 0"))
	}
	permiso, err := p.permisoService.ObtenerPermisoById(c.UserContext(), &permisoId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(permiso)
}

func (p PermisoHandler) RegistrarPermiso(c *fiber.Ctx) error {
	var request domain.PermisoRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	permisoId, err := p.permisoService.RegistrarPermiso(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessageData(domain.PermisoId{Id: *permisoId}, "Permiso registrado correctamente"))
}

func (p PermisoHandler) ModificarPermisoById(c *fiber.Ctx) error {
	permisoId, err := c.ParamsInt("permisoId", 0)
	if err != nil || permisoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del permiso debe ser un número válido mayor a 0"))
	}
	var request domain.PermisoRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err = p.permisoService.ModificarPermisoById(c.UserContext(), &permisoId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Permiso modificado correctamente"))
}

func NewPermisoHandler(permisoService port.PermisoService) *PermisoHandler {
	return &PermisoHandler{permisoService: permisoService}
}

var _ port.PermisoHandler = (*PermisoHandler)(nil)
