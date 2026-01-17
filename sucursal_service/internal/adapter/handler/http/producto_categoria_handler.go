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

type ProductoCategoriaHandler struct {
	productoCategoriaService port.ProductoCategoriaService
}

func (p ProductoCategoriaHandler) RegistrarCategoria(c *fiber.Ctx) error {
	var request domain.ProductoCategoriaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	id, err := p.productoCategoriaService.RegistrarCategoria(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.ProductoCategoriaId{Id: *id}, "Categoría de productos registrado correctamente"))
}

func (p ProductoCategoriaHandler) ModificarCategoriaById(c *fiber.Ctx) error {
	categoriaId, err := c.ParamsInt("categoriaId", 0)
	if err != nil || categoriaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de categoría debe ser un número válido mayor a 0"))
	}
	var request domain.ProductoCategoriaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err = p.productoCategoriaService.ModificarCategoriaById(c.UserContext(), &categoriaId, &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("Categoría de productos modificado correctamente"))
}

func (p ProductoCategoriaHandler) ListarCategorias(c *fiber.Ctx) error {
	list, err := p.productoCategoriaService.ListarCategorias(c.UserContext(), c.Queries())
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

func (p ProductoCategoriaHandler) ObtenerCategoriaById(c *fiber.Ctx) error {
	categoriaId, err := c.ParamsInt("categoriaId", 0)
	if err != nil || categoriaId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' de categoría debe ser un número válido mayor a 0"))
	}
	categoria, err := p.productoCategoriaService.ObtenerCategoriaById(c.UserContext(), &categoriaId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(categoria)
}

func NewProductoCategoriaHandler(productoCategoriaService port.ProductoCategoriaService) *ProductoCategoriaHandler {
	return &ProductoCategoriaHandler{productoCategoriaService: productoCategoriaService}
}

var _ port.ProductoCategoriaHandler = (*ProductoCategoriaHandler)(nil)
