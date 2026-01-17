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

type ProductoHandler struct {
	productoService port.ProductoService
}

func (p ProductoHandler) ObtenerProductoSucursalById(c *fiber.Ctx) error {
	productoSucursalId, err := c.ParamsInt("productoSucursalId", 0)
	if err != nil || productoSucursalId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' debe ser un número válido mayor a 0"))
	}
	productoSucursal, err := p.productoService.ObtenerProductoSucursalById(c.UserContext(), &productoSucursalId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(productoSucursal)
}

func (p ProductoHandler) ActualizarProductoSucursal(c *fiber.Ctx) error {
	productoSucursalId, err := c.ParamsInt("productoSucursalId", 0)
	if err != nil || productoSucursalId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' debe ser un número válido mayor a 0"))
	}
	var request domain.ProductoSucursalUpdateRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err = p.productoService.ActualizarProductoSucursal(c.UserContext(), &productoSucursalId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Producto de sucursal actualizado correctamente"))
}

func (p ProductoHandler) ListarProductosPorSucursal(c *fiber.Ctx) error {
	list, err := p.productoService.ListarProductosPorSucursal(c.UserContext(), c.Queries())
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

func (p ProductoHandler) HabilitarProductoById(c *fiber.Ctx) error {
	productoId, err := c.ParamsInt("productoId", 0)
	if err != nil || productoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del producto debe ser un número válido mayor a 0"))
	}

	err = p.productoService.HabilitarProductoById(c.UserContext(), &productoId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Producto habilitado correctamente"))
}

func (p ProductoHandler) DeshabilitarProductoById(c *fiber.Ctx) error {
	productoId, err := c.ParamsInt("productoId", 0)
	if err != nil || productoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del producto debe ser un número válido mayor a 0"))
	}

	err = p.productoService.DeshabilitarProductoById(c.UserContext(), &productoId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Producto deshabilitado correctamente"))
}

func (p ProductoHandler) ListarProductosMasVendidos(c *fiber.Ctx) error {
	list, err := p.productoService.ListarProductosMasVendidos(c.UserContext(), c.Queries())
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

func (p ProductoHandler) RegistrarProducto(c *fiber.Ctx) error {
	var request domain.ProductoRequest
	if err := json.Unmarshal([]byte(c.FormValue("body")), &request); err != nil {
		log.Println("Error al deserializar body:", err)
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("Error al leer el formulario"))
	}
	productoId, err := p.productoService.RegistrarProducto(c.UserContext(), &request, fileHeader)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.ProductoId{Id: *productoId}, "Producto registrado correctamente"))
}

func (p ProductoHandler) ModificarProductoById(c *fiber.Ctx) error {
	var request domain.ProductoRequest
	if err := json.Unmarshal([]byte(c.FormValue("body")), &request); err != nil {
		log.Println("Error al deserializar body:", err)
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	productoId, err := c.ParamsInt("productoId", 0)
	if err != nil || productoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del producto debe ser un número válido mayor a 0"))
	}
	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("Error al leer el formulario"))
	}
	err = p.productoService.ModificarProductoById(c.UserContext(), &productoId, &request, fileHeader)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Producto modificado correctamente"))
}

func (p ProductoHandler) ListarProductos(c *fiber.Ctx) error {
	list, err := p.productoService.ListarProductos(c.UserContext(), c.Queries())
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

func (p ProductoHandler) ObtenerProductoById(c *fiber.Ctx) error {
	productoId, err := c.ParamsInt("productoId", 0)
	if err != nil || productoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del producto debe ser un número válido mayor a 0"))
	}
	producto, err := p.productoService.ObtenerProductoById(c.UserContext(), &productoId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(producto)
}

func NewProductoHandler(productoService port.ProductoService) *ProductoHandler {
	return &ProductoHandler{productoService: productoService}
}

var _ port.ProductoHandler = (*ProductoHandler)(nil)
