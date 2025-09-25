package port

import (
	"context"
	"multiroom/dispositivo-service/internal/core/domain"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type DispositivoRepository interface {
	RegistrarDispositivo(ctx context.Context, request *domain.DispositivoRequest) error
	ObtenerListaDispositivos(ctx context.Context, filtros map[string]string) (*[]domain.DispositivoInfo, error)
	ObtenerListaDispositivosByUsuarioId(ctx context.Context, usuarioId *int) (*[]domain.DispositivoInfo, error)
	DeshabilitarDispositivo(ctx context.Context, id *int) error
	HabilitarDispositivo(ctx context.Context, id *int) error
	ObtenerDispositivoById(ctx context.Context, id *int) (*domain.DispositivoInfo, error)
	ObtenerDispositivoByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.DispositivoInfo, error)
	EliminarDispositivoById(ctx context.Context, id *int) error
	ActualizarDispositivoEnLinea(ctx context.Context, id *int, enLinea *bool) error
}

type DispositivoService interface {
	RegistrarDispositivo(ctx context.Context, request *domain.DispositivoRequest) error
	ObtenerListaDispositivos(ctx context.Context, filtros map[string]string) (*[]domain.DispositivoInfo, error)
	ObtenerListaDispositivosByUsuarioId(ctx context.Context, usuarioId *int) (*[]domain.DispositivoInfo, error)
	DeshabilitarDispositivo(ctx context.Context, id *int) error
	HabilitarDispositivo(ctx context.Context, id *int) error
	ObtenerDispositivoById(ctx context.Context, id *int) (*domain.DispositivoInfo, error)
	ObtenerDispositivoByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.DispositivoInfo, error)
	EliminarDispositivoById(ctx context.Context, id *int) error
	ActualizarDispositivoEnLinea(ctx context.Context, id *int, enLinea *bool) error
}

type DispositivoHandler interface {
	RegistrarDispositivo(c *fiber.Ctx) error
	ObtenerListaDispositivos(c *fiber.Ctx) error
	ObtenerListaDispositivosByUsuarioId(c *fiber.Ctx) error
	DeshabilitarDispositivo(c *fiber.Ctx) error
	HabilitarDispositivo(c *fiber.Ctx) error
	ObtenerDispositivoByDispositivoId(c *fiber.Ctx) error
	EliminarDispositivoById(c *fiber.Ctx) error
}

type DispositivoHandlerWS interface {
	NotificarDispositivoHabilitar(c *websocket.Conn)
}
