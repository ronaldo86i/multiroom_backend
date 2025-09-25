package http

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"
)

type AppVersionHandler struct {
	appVersionService port.AppVersionService
}

func (a AppVersionHandler) ObtenerListaVersiones(c *fiber.Ctx) error {

	lista, err := a.appVersionService.ObtenerListaVersiones(c.UserContext(), c.Queries())
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(lista)
}

func (a AppVersionHandler) ObtenerUltimaVersion(c *fiber.Ctx) error {
	q := domain.AppLastVersionQuery{
		Tipo:         domain.TipoApp(c.Query("tipo")),
		Arquitectura: domain.ArquitecturaApp(c.Query("arquitectura")),
		Os:           domain.OsApp(c.Query("os")),
	}

	appVersion, err := a.appVersionService.ObtenerUltimaVersion(c.UserContext(), &q)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	return c.JSON(appVersion)
}

func (a AppVersionHandler) RegistrarApp(c *fiber.Ctx) error {

	var request domain.AppVersionRequest
	if err := json.Unmarshal([]byte(c.FormValue("body")), &request); err != nil {
		log.Println("Error al deserializar body:", err)
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("Error al leer el formulario"))
	}
	appId, err := a.appVersionService.RegistrarApp(c.UserContext(), &request, fileHeader)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessageData(domain.AppVersionId{Id: *appId}, "App registrado correctamente"))
}

func (a AppVersionHandler) ModificarVersion(c *fiber.Ctx) error {
	appVersionId, err := c.ParamsInt("appVersionId", 0)
	if err != nil || appVersionId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}
	var request domain.AppVersionRequest
	if err := json.Unmarshal([]byte(c.FormValue("body")), &request); err != nil {
		log.Println("Error al deserializar body:", err)
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	fileHeader, err := c.FormFile("image")
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("Error al leer el formulario"))
	}
	err = a.appVersionService.ModificarVersion(c.UserContext(), &appVersionId, &request, fileHeader)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	return c.JSON(util.NewMessage("App modificada correctamente"))
}

func (a AppVersionHandler) ObtenerVersion(c *fiber.Ctx) error {
	appVersionId, err := c.ParamsInt("appVersionId", 0)
	if err != nil || appVersionId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del país debe ser un número válido mayor a 0"))
	}
	appVersion, err := a.appVersionService.ObtenerVersion(c.UserContext(), &appVersionId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	return c.JSON(appVersion)
}

func NewAppVersionHandler(appVersionService port.AppVersionService) *AppVersionHandler {
	return &AppVersionHandler{appVersionService: appVersionService}
}

var _ port.AppVersionHandler = (*AppVersionHandler)(nil)
