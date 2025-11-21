package http

import (
	"errors"
	"log"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"

	"github.com/gofiber/fiber/v2"
)

type MetodoPagoHandler struct {
	metodoPagoService port.MetodoPagoService
}

func (m MetodoPagoHandler) ListarMetodosPago(c *fiber.Ctx) error {
	metodosPago, err := m.metodoPagoService.ListarMetodosPago(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&metodosPago)
}

func NewMetodoPagoHandler(metodoPagoService port.MetodoPagoService) *MetodoPagoHandler {
	return &MetodoPagoHandler{metodoPagoService: metodoPagoService}
}

var _ port.MetodoPagoHandler = (*MetodoPagoHandler)(nil)
