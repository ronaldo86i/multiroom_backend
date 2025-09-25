package http

import (
	"errors"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/domain/datatype"
	"multiroom/dispositivo-service/internal/core/port"
	"multiroom/dispositivo-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type DispositivoHandler struct {
	dispositivoService port.DispositivoService
	rabbitService      port.RabbitMQService
}

func (d DispositivoHandler) EliminarDispositivoById(c *fiber.Ctx) error {
	dispositivoId, err := c.ParamsInt("dispositivoId", 0)
	if err != nil || dispositivoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del dispositivo debe ser un número válido mayor a 0"))
	}
	err = d.dispositivoService.EliminarDispositivoById(c.UserContext(), &dispositivoId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Dispositivo eliminado correctamente"))
}

func (d DispositivoHandler) ObtenerDispositivoByDispositivoId(c *fiber.Ctx) error {
	dispositivoId := c.Params("dispositivoId")

	list, err := d.dispositivoService.ObtenerDispositivoByDispositivoId(c.UserContext(), &dispositivoId)
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

func (d DispositivoHandler) DeshabilitarDispositivo(c *fiber.Ctx) error {
	dispositivoId, err := c.ParamsInt("dispositivoId", 0)
	if err != nil || dispositivoId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del dispositivo debe ser un número válido mayor a 0"))
	}
	err = d.dispositivoService.DeshabilitarDispositivo(c.UserContext(), &dispositivoId)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	//dispositivo, err := d.dispositivoService.ObtenerDispositivoById(c.UserContext(), &dispositivoId)
	//if err != nil {
	//	log.Print(err.Error())
	//	var errorResponse *datatype.ErrorResponse
	//	if errors.As(err, &errorResponse) {
	//		return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
	//	}
	//	return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	//}
	//queue := fmt.Sprintf("usuario_%d_dispositivo_%d", dispositivo.Usuario.Id, dispositivo.Id)
	//err = d.rabbitService.Publish(queue, &dispositivo, 1)
	//if err != nil {
	//	log.Print("Error al publicar estado del dispositivo al usuario:", err.Error())
	//}
	return c.Status(http.StatusOK).JSON(util.NewMessage("Dispositivo deshabilitado correctamente"))
}

func (d DispositivoHandler) HabilitarDispositivo(c *fiber.Ctx) error {
	dispositivoId, err := c.ParamsInt("dispositivoId", 0)
	if err != nil || dispositivoId <= 0 {
		return c.Status(http.StatusBadRequest).
			JSON(util.NewMessage("El 'id' del dispositivo debe ser un número válido mayor a 0"))
	}

	ctx := c.UserContext()

	// Habilitar el dispositivo
	if err := d.dispositivoService.HabilitarDispositivo(ctx, &dispositivoId); err != nil {
		log.Println("❌ Error habilitando dispositivo:", err)
		return util.HandleServiceError(err)
	}

	////// Obtener el dispositivo actualizado
	//dispositivo, err := d.dispositivoService.ObtenerDispositivoById(ctx, &dispositivoId)
	//if err != nil {
	//	log.Println("❌ Error obteniendo dispositivo:", err)
	//	return util.HandleServiceError(err)
	//}

	////Publicar mensaje a RabbitMQ (de forma asíncrona opcionalmente)
	//queue := fmt.Sprintf("dispositivo_usuario_%d", dispositivo.Usuario.Id)
	//if err := d.rabbitService.Publish(queue, &dispositivo); err != nil {
	//	log.Printf("⚠️ Error publicando a RabbitMQ [%s]: %v", queue, err)
	//}

	return c.Status(http.StatusOK).JSON(util.NewMessage("Dispositivo habilitado correctamente"))
}

func (d DispositivoHandler) ObtenerListaDispositivosByUsuarioId(c *fiber.Ctx) error {
	usuarioId, err := c.ParamsInt("usuarioId", 0)
	if err != nil || usuarioId <= 0 {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("El 'id' del usuario debe ser un número válido mayor a 0"))
	}

	list, err := d.dispositivoService.ObtenerListaDispositivosByUsuarioId(c.UserContext(), &usuarioId)
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

func (d DispositivoHandler) RegistrarDispositivo(c *fiber.Ctx) error {
	var request domain.DispositivoRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	err := d.dispositivoService.RegistrarDispositivo(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}

	return c.Status(http.StatusCreated).JSON(util.NewMessage("Dispositivo registrado correctamente"))
}

func (d DispositivoHandler) ObtenerListaDispositivos(c *fiber.Ctx) error {
	filtros := c.Queries()
	list, err := d.dispositivoService.ObtenerListaDispositivos(c.UserContext(), filtros)
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

func NewDispositivoHandler(dispositivoService port.DispositivoService, rabbitService port.RabbitMQService) *DispositivoHandler {
	return &DispositivoHandler{dispositivoService: dispositivoService, rabbitService: rabbitService}
}

var _ port.DispositivoHandler = (*DispositivoHandler)(nil)
