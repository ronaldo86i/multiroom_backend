package http

import (
	"errors"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SalaHandler struct {
	salaService     port.SalaService
	rabbitMQService port.RabbitMQService
}

func (s SalaHandler) ObtenerListaUsoSalas(c *fiber.Ctx) error {
	salas, err := s.salaService.ObtenerListaUsoSalas(c.UserContext(), c.Queries())
	if err != nil {
		return handleError(err)
	}
	return c.JSON(salas)
}

func (s SalaHandler) EliminarSalaById(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)
	if err := s.salaService.HabilitarSala(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}

	err := s.salaService.EliminarSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}
	publishSalaAsync(s.rabbitMQService, *sala, salaId)
	return c.JSON(util.NewMessage("Sala eliminada correctamente"))
}

// --- Helper para manejar errores ---
func handleError(err error) error {
	var errorResponse *datatype.ErrorResponse
	if errors.As(err, &errorResponse) {
		return fiber.NewError(errorResponse.Code, errorResponse.Message)
	}
	return fiber.NewError(http.StatusInternalServerError, "Error interno del servidor")
}

// --- Helper para publicar en RabbitMQ async ---
func publishSalaAsync(rabbitMQ port.RabbitMQService, sala domain.SalaDetail, salaId int) {
	go func(s domain.SalaDetail, id int) {

		if err := rabbitMQ.Publish(fmt.Sprintf("salas_%d", salaId), s, amqp.Table{
			// Máximo de mensajes
			amqp.QueueMaxLenArg: int32(1),

			// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {
			log.Print("Error al publicar sala_"+fmt.Sprint(id)+":", err)
		}

		if err := rabbitMQ.Publish("salas", s, amqp.Table{
			// Máximo de mensajes
			amqp.QueueMaxLenArg: int32(1),

			// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {
			log.Print("Error al publicar canal general salas:", err)
		}

		if err := rabbitMQ.Publish(fmt.Sprintf("dispositivo_%d_usuario_%d", s.Dispositivo.Id, s.Dispositivo.Usuario.Id), s, amqp.Table{
			// Máximo de mensajes
			amqp.QueueMaxLenArg: int32(1),

			// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {
			log.Print("Error al publicar en ", fmt.Sprintf("dispositivo_%d_usuario_%d", s.Dispositivo.Id, s.Dispositivo.Usuario.Id), ":", err)
		}

		if err := rabbitMQ.Publish(fmt.Sprintf("sucursal_%d_salas", s.Sucursal.Id), s, amqp.Table{
			// Máximo de mensajes
			amqp.QueueMaxLenArg: int32(1),

			// Política de descarte ("drop-head" elimina el más antiguo, "reject-publish" rechaza mensajes nuevos)
			amqp.QueueOverflowArg: amqp.QueueOverflowDropHead,
		}); err != nil {
			log.Print("Error al publicar en ", fmt.Sprintf("sucursal_%d_salas", s.Sucursal.Id), ":", err)
		}
	}(sala, salaId)
}

// --- Endpoints ---

func (s SalaHandler) IncrementarTiempoUsoSala(c *fiber.Ctx) error {
	var request domain.UsoSalaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida"))
	}

	salaId, _ := c.ParamsInt("salaId", 0)

	// Obtener sala
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	// Validar rol y sucursal
	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}
		if sucursalId != sala.Sucursal.Id {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("La sala no pertenece a este usuario sucursal"))
		}
	case "":
		// rol nil o no definido → se permite continuar
	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	// Incrementar tiempo
	if err := s.salaService.IncrementarTiempoUsoSala(c.UserContext(), &salaId, &request); err != nil {
		return handleError(err)
	}

	// Volver a obtener sala y publicar
	if sala, err = s.salaService.ObtenerSalaById(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}
	publishSalaAsync(s.rabbitMQService, *sala, salaId)

	return c.JSON(util.NewMessage("Se ha modificado el tiempo de uso correctamente"))
}

func (s SalaHandler) CancelarSala(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)

	// Obtener sala
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	// Validar rol y sucursal
	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}
		if sucursalId != sala.Sucursal.Id {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("La sala no pertenece a este usuario sucursal"))
		}
	case "":
		// rol nil o no definido → se permite continuar
	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	// Cancelar sala
	if err := s.salaService.CancelarSala(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}

	// Volver a obtener sala y publicar
	if sala, err = s.salaService.ObtenerSalaById(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}
	publishSalaAsync(s.rabbitMQService, *sala, salaId)

	return c.JSON(util.NewMessage("Se ha cancelado el tiempo de uso correctamente"))
}

func (s SalaHandler) AsignarTiempoUsoSala(c *fiber.Ctx) error {
	var request domain.UsoSalaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida"))
	}

	salaId := request.SalaId

	// Obtener sala
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	// Validar rol y sucursal
	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}
		if sucursalId != sala.Sucursal.Id {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("La sala no pertenece a este usuario sucursal"))
		}
	case "":
		// rol nil o no definido → se permite continuar
	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	// Asignar tiempo
	usoId, err := s.salaService.AsignarTiempoUsoSala(c.UserContext(), &request)
	if err != nil {
		return handleError(err)
	}

	// Volver a obtener sala y publicar
	if sala, err = s.salaService.ObtenerSalaById(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}
	publishSalaAsync(s.rabbitMQService, *sala, salaId)

	return c.JSON(util.NewMessageData(domain.UsoSalaId{Id: *usoId}, "Se ha asignado tiempo de uso correctamente"))
}

func (s SalaHandler) PausarTiempoUsoSala(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)

	// Obtener sala
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	// Validar rol y sucursal
	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}
		if sucursalId != sala.Sucursal.Id {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("La sala no pertenece a este usuario sucursal"))
		}
	case "":
		// rol nil o no definido → se permite continuar
	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	// Pausar sala
	if err := s.salaService.PausarTiempoUsoSala(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}

	// Volver a obtener sala y publicar
	if sala, err = s.salaService.ObtenerSalaById(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}
	publishSalaAsync(s.rabbitMQService, *sala, salaId)

	return c.JSON(util.NewMessage("Se ha pausado el tiempo de uso correctamente"))
}

func (s SalaHandler) ReanudarTiempoUsoSala(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)

	// Obtener sala
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	// Validar rol y sucursal
	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}
		if sucursalId != sala.Sucursal.Id {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("La sala no pertenece a este usuario sucursal"))
		}
	case "":
		// rol nil o no definido → se permite continuar
	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	// Reanudar sala
	if err := s.salaService.ReanudarTiempoUsoSala(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}

	// Volver a obtener sala y publicar
	if sala, err = s.salaService.ObtenerSalaById(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}
	publishSalaAsync(s.rabbitMQService, *sala, salaId)

	return c.JSON(util.NewMessage("Se ha reanudado el tiempo de uso correctamente"))
}

func (s SalaHandler) RegistrarSala(c *fiber.Ctx) error {
	var request domain.SalaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	salaId, err := s.salaService.RegistrarSala(c.UserContext(), &request)
	if err != nil {
		return handleError(err)
	}

	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), salaId)
	if err != nil {
		return handleError(err)
	}

	publishSalaAsync(s.rabbitMQService, *sala, *salaId)
	return c.Status(http.StatusCreated).JSON(util.NewMessageData(domain.SalaId{Id: *salaId}, "Sala registrada correctamente"))
}

func (s SalaHandler) ModificarSala(c *fiber.Ctx) error {
	var request domain.SalaRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	salaId, _ := c.ParamsInt("salaId", 0)

	if err := s.salaService.ModificarSala(c.UserContext(), &salaId, &request); err != nil {
		return handleError(err)
	}

	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	publishSalaAsync(s.rabbitMQService, *sala, salaId)
	return c.JSON(util.NewMessage("Sala modificada correctamente"))
}

func (s SalaHandler) ObtenerSalaById(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)

	// Obtener sala
	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	// Validar rol y sucursal
	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}
		if sucursalId != sala.Sucursal.Id {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("La sala no pertenece a este usuario sucursal"))
		}
	case "":
		// rol nil o no definido → se permite continuar
	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	return c.JSON(sala)
}

func (s SalaHandler) ObtenerListaSalas(c *fiber.Ctx) error {
	var lista []domain.SalaInfo

	rol, _ := c.Locals("rol").(string)
	switch rol {
	case "usuario-sucursal":
		sucursalId, ok := c.Locals("sucursalId").(int)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(util.NewMessage("sucursalId no válido o no presente"))
		}

		filtros := map[string]string{"sucursalId": fmt.Sprintf("%d", sucursalId)}
		salas, err := s.salaService.ObtenerListaSalas(c.UserContext(), filtros)
		if err != nil {
			return handleError(err)
		}
		lista = *salas

	case "":
		// rol nil o no definido → usar filtros de la query
		salas, err := s.salaService.ObtenerListaSalas(c.UserContext(), c.Queries())
		if err != nil {
			return handleError(err)
		}
		lista = *salas

	default:
		return c.Status(fiber.StatusForbidden).JSON(util.NewMessage("Rol no permitido"))
	}

	return c.JSON(lista)
}

func (s SalaHandler) HabilitarSala(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)
	if err := s.salaService.HabilitarSala(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}

	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	publishSalaAsync(s.rabbitMQService, *sala, salaId)
	return c.JSON(util.NewMessage("Sala habilitada correctamente"))
}

func (s SalaHandler) DeshabilitarSala(c *fiber.Ctx) error {
	salaId, _ := c.ParamsInt("salaId", 0)
	if err := s.salaService.DeshabilitarSala(c.UserContext(), &salaId); err != nil {
		return handleError(err)
	}

	sala, err := s.salaService.ObtenerSalaById(c.UserContext(), &salaId)
	if err != nil {
		return handleError(err)
	}

	publishSalaAsync(s.rabbitMQService, *sala, salaId)
	return c.JSON(util.NewMessage("Sala deshabilitada correctamente"))
}

// Constructor
func NewSalaHandler(salaService port.SalaService, rabbitMQService port.RabbitMQService) *SalaHandler {
	return &SalaHandler{salaService: salaService, rabbitMQService: rabbitMQService}
}

var _ port.SalaHandler = (*SalaHandler)(nil)
