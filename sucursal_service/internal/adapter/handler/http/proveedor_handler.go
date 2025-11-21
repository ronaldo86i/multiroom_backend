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

type ProveedorHandler struct {
	proveedorService port.ProveedorService
}

func (p ProveedorHandler) RegistrarProveedor(c *fiber.Ctx) error {
	var request domain.ProveedorRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	proveedorId, err := p.proveedorService.RegistrarProveedor(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.ProveedorId{Id: *proveedorId}, "Proveedor registrado correctamente"))
}

func (p ProveedorHandler) ModificarProveedor(c *fiber.Ctx) error {
	var request domain.ProveedorRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	proveedorId, err := c.ParamsInt("proveedorId", 0)
	if err != nil || proveedorId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del proveedor debe ser un número válido mayor a 0"))
	}
	err = p.proveedorService.ModificarProveedor(c.UserContext(), &proveedorId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Proveedor modificado correctamente"))
}

func (p ProveedorHandler) ListarProveedores(c *fiber.Ctx) error {
	list, err := p.proveedorService.ListarProveedores(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(list)
}

func (p ProveedorHandler) ObtenerProveedorById(c *fiber.Ctx) error {
	proveedorId, err := c.ParamsInt("proveedorId", 0)
	if err != nil || proveedorId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del proveedor debe ser un número válido mayor a 0"))
	}
	proveedor, err := p.proveedorService.ObtenerProveedorById(c.UserContext(), &proveedorId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(proveedor)
}

func NewProveedorHandler(proveedorService port.ProveedorService) *ProveedorHandler {
	return &ProveedorHandler{proveedorService: proveedorService}
}

var _ port.ProveedorHandler = (*ProveedorHandler)(nil)
