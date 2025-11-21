package http

import (
	"encoding/json"
	"errors"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type PaisHandler struct {
	paisService port.PaisService
}

func (p PaisHandler) HabilitarPaisById(c *fiber.Ctx) error {
	paisId, err := c.ParamsInt("paisId", 0)
	if err != nil || paisId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}
	err = p.paisService.HabilitarPaisById(c.UserContext(), &paisId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	return c.JSON(util.NewMessage("País habilitado correctamente"))
}

func (p PaisHandler) DeshabilitarPaisById(c *fiber.Ctx) error {
	paisId, err := c.ParamsInt("paisId", 0)
	if err != nil || paisId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}

	err = p.paisService.DeshabilitarPaisById(c.UserContext(), &paisId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	return c.JSON(util.NewMessage("País deshabilitado correctamente"))
}

func (p PaisHandler) RegistrarPais(c *fiber.Ctx) error {
	var paisRequest domain.PaisRequest
	if err := json.Unmarshal([]byte(c.FormValue("body")), &paisRequest); err != nil {
		log.Println("Error al deserializar body:", err)
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("Error al leer el formulario"))
	}

	paisId, err := p.paisService.RegistrarPais(c.UserContext(), &paisRequest, fileHeader)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.PaisId{Id: *paisId}, "País registrado correctamente"))
}

func (p PaisHandler) ModificarPais(c *fiber.Ctx) error {
	paisId, err := c.ParamsInt("paisId", 0)
	if err != nil || paisId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}
	var paisRequest domain.PaisRequest
	if err := json.Unmarshal([]byte(c.FormValue("body")), &paisRequest); err != nil {
		log.Println("Error al deserializar body:", err)
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("Error al leer el formulario"))
	}

	err = p.paisService.ModificarPais(c.UserContext(), &paisId, &paisRequest, fileHeader)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("País modificado correctamente"))
}

func (p PaisHandler) ObtenerPaisById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	paisId, err := c.ParamsInt("paisId", 0)
	if err != nil || paisId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}
	pais, err := p.paisService.ObtenerPaisById(ctx, &paisId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&pais)
}

func (p PaisHandler) ObtenerListaPaises(c *fiber.Ctx) error {
	paises, err := p.paisService.ObtenerListaPaises(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(&paises)
}

func NewPaisHandler(paisService port.PaisService) *PaisHandler {
	return &PaisHandler{paisService: paisService}
}

var _ port.PaisHandler = (*PaisHandler)(nil)
