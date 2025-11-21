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

type UbicacionHandler struct {
	ubicacionService port.UbicacionService
}

func (u UbicacionHandler) HabilitarUbicacion(c *fiber.Ctx) error {
	ubicacionId, err := c.ParamsInt("ubicacionId", 0)
	if err != nil || ubicacionId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la ubicación debe ser un número válido mayor a 0"))
	}
	err = u.ubicacionService.HabilitarUbicacion(c.UserContext(), &ubicacionId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Ubicación de sucursal habilitada correctamente"))
}

func (u UbicacionHandler) DeshabilitarUbicacion(c *fiber.Ctx) error {
	ubicacionId, err := c.ParamsInt("ubicacionId", 0)
	if err != nil || ubicacionId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la ubicación debe ser un número válido mayor a 0"))
	}
	err = u.ubicacionService.DeshabilitarUbicacion(c.UserContext(), &ubicacionId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Ubicación de sucursal deshabilitada correctamente"))

}

func (u UbicacionHandler) RegistrarUbicacion(c *fiber.Ctx) error {
	var request domain.UbicacionRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	ubicacionId, err := u.ubicacionService.RegistrarUbicacion(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.UbicacionId{Id: *ubicacionId}, "Ubicación de sucursal registrado correctamente"))
}

func (u UbicacionHandler) ModificarUbicacionById(c *fiber.Ctx) error {
	var request domain.UbicacionRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	ubicacionId, err := c.ParamsInt("ubicacionId", 0)
	if err != nil || ubicacionId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la ubicación debe ser un número válido mayor a 0"))
	}

	err = u.ubicacionService.ModificarUbicacionById(c.UserContext(), &ubicacionId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Ubicación de sucursal modificado correctamente"))
}

func (u UbicacionHandler) ListarUbicaciones(c *fiber.Ctx) error {
	ubicaciones, err := u.ubicacionService.ListarUbicaciones(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&ubicaciones)
}

func (u UbicacionHandler) ObtenerUbicacionById(c *fiber.Ctx) error {
	ubicacionId, err := c.ParamsInt("ubicacionId", 0)
	if err != nil || ubicacionId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de la ubicación debe ser un número válido mayor a 0"))
	}
	ubicacion, err := u.ubicacionService.ObtenerUbicacionById(c.UserContext(), &ubicacionId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&ubicacion)
}

func NewUbicacionHandler(ubicacionService port.UbicacionService) *UbicacionHandler {
	return &UbicacionHandler{ubicacionService: ubicacionService}
}

var _ port.UbicacionHandler = (*UbicacionHandler)(nil)
