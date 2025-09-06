package http

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"
)

type SucursalHandler struct {
	sucursalService port.SucursalService
}

func (s SucursalHandler) RegistrarSucursal(c *fiber.Ctx) error {
	var request domain.SucursalRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	sucursalId, err := s.sucursalService.RegistrarSucursal(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.SucursalId{Id: *sucursalId}, "Sucursal registrado correctamente"))
}

func (s SucursalHandler) ModificarSucursal(c *fiber.Ctx) error {
	var request domain.SucursalRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	sucursalId, err := c.ParamsInt("sucursalId", 0)
	if err != nil || sucursalId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}
	err = s.sucursalService.ModificarSucursal(c.UserContext(), &sucursalId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Sucursal modificado correctamente"))
}

func (s SucursalHandler) ObtenerSucursalById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	sucursalId, err := c.ParamsInt("sucursalId", 0)
	if err != nil || sucursalId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de sucursal debe ser un número válido mayor a 0"))
	}
	sucursal, err := s.sucursalService.ObtenerSucursalById(ctx, &sucursalId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&sucursal)
}

func (s SucursalHandler) ObtenerListaSucursales(c *fiber.Ctx) error {
	sucursales, err := s.sucursalService.ObtenerListaSucursales(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&sucursales)
}

func (s SucursalHandler) HabilitarSucursal(c *fiber.Ctx) error {
	sucursalId, err := c.ParamsInt("sucursalId", 0)
	if err != nil || sucursalId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de sucursal debe ser un número válido mayor a 0"))
	}
	err = s.sucursalService.HabilitarSucursal(c.UserContext(), &sucursalId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Sucursal actualizado correctamente"))
}

func (s SucursalHandler) DeshabilitarSucursal(c *fiber.Ctx) error {
	sucursalId, err := c.ParamsInt("sucursalId", 0)
	if err != nil || sucursalId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de sucursal debe ser un número válido mayor a 0"))
	}
	err = s.sucursalService.DeshabilitarSucursal(c.UserContext(), &sucursalId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Sucursal actualizado correctamente"))
}

func NewSucursalHandler(sucursalService port.SucursalService) *SucursalHandler {
	return &SucursalHandler{sucursalService: sucursalService}
}

var _ port.SucursalHandler = (*SucursalHandler)(nil)
