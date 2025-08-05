package port

import (
	"context"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"multiroom/dispositivo-service/internal/core/domain"
)

type DispositivoRepository interface {
	RegistrarDispositivo(ctx context.Context, request *domain.DispositivoRequest) error
	ObtenerListaDispositivos(ctx context.Context, filtros map[string]string) (*[]domain.DispositivoInfo, error)
	ObtenerListaDispositivosByUsuarioId(ctx context.Context, usuarioId *int) (*[]domain.DispositivoInfo, error)
	DeshabilitarDispositivo(ctx context.Context, id *int) error
	HabilitarDispositivo(ctx context.Context, id *int) error
	ObtenerDispositivoById(ctx context.Context, id *int) (*domain.DispositivoInfo, error)
}

type DispositivoService interface {
	RegistrarDispositivo(ctx context.Context, request *domain.DispositivoRequest) error
	ObtenerListaDispositivos(ctx context.Context, filtros map[string]string) (*[]domain.DispositivoInfo, error)
	ObtenerListaDispositivosByUsuarioId(ctx context.Context, usuarioId *int) (*[]domain.DispositivoInfo, error)
	DeshabilitarDispositivo(ctx context.Context, id *int) error
	HabilitarDispositivo(ctx context.Context, id *int) error
	ObtenerDispositivoById(ctx context.Context, id *int) (*domain.DispositivoInfo, error)
}

type DispositivoHandler interface {
	RegistrarDispositivo(c *fiber.Ctx) error
	ObtenerListaDispositivos(c *fiber.Ctx) error
	ObtenerListaDispositivosByUsuarioId(c *fiber.Ctx) error
	DeshabilitarDispositivo(c *fiber.Ctx) error
	HabilitarDispositivo(c *fiber.Ctx) error
}

type DispositivoHandlerWS interface {
	NotificarDispositivoHabilitar(c *websocket.Conn)
}
