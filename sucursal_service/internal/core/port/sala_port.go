package port

import (
	"context"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"multiroom/sucursal-service/internal/core/domain"
)

type SalaRepository interface {
	RegistrarSala(ctx context.Context, request *domain.SalaRequest) (*int, error)
	ModificarSala(ctx context.Context, id *int, request *domain.SalaRequest) error
	ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error)
	ObtenerListaSalas(ctx context.Context, filtros map[string]string) (*[]domain.SalaInfo, error)
	HabilitarSala(ctx context.Context, id *int) error
	DeshabilitarSala(ctx context.Context, id *int) error
	IncrementarTiempoUsoSala(ctx context.Context, salaId *int, request *domain.UsoSalaRequest) error
	CancelarSala(ctx context.Context, salaId *int) error
	AsignarTiempoUsoSala(ctx context.Context, request *domain.UsoSalaRequest) error
	PausarTiempoUsoSala(ctx context.Context, salaId *int) error
	ReanudarTiempoUsoSala(ctx context.Context, salaId *int) error
	ActualizarUsoSalas(ctx context.Context) (*[]int, error)
	ObtenerListaSalasDetailByIds(ctx context.Context, ids []int) (*[]domain.SalaDetail, error)
}

type SalaService interface {
	RegistrarSala(ctx context.Context, request *domain.SalaRequest) (*int, error)
	ModificarSala(ctx context.Context, id *int, request *domain.SalaRequest) error
	ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error)
	ObtenerListaSalas(ctx context.Context, filtros map[string]string) (*[]domain.SalaInfo, error)
	HabilitarSala(ctx context.Context, id *int) error
	DeshabilitarSala(ctx context.Context, id *int) error
	IncrementarTiempoUsoSala(ctx context.Context, salaId *int, request *domain.UsoSalaRequest) error
	CancelarSala(ctx context.Context, salaId *int) error
	AsignarTiempoUsoSala(ctx context.Context, request *domain.UsoSalaRequest) error
	PausarTiempoUsoSala(ctx context.Context, salaId *int) error
	ReanudarTiempoUsoSala(ctx context.Context, salaId *int) error
	ActualizarUsoSalas(ctx context.Context) (*[]int, error)
	ObtenerListaSalasDetailByIds(ctx context.Context, ids []int) (*[]domain.SalaDetail, error)
}

type SalaHandler interface {
	RegistrarSala(c *fiber.Ctx) error
	ModificarSala(c *fiber.Ctx) error
	ObtenerSalaById(c *fiber.Ctx) error
	ObtenerListaSalas(c *fiber.Ctx) error
	HabilitarSala(c *fiber.Ctx) error
	DeshabilitarSala(c *fiber.Ctx) error
	IncrementarTiempoUsoSala(c *fiber.Ctx) error
	CancelarSala(c *fiber.Ctx) error
	AsignarTiempoUsoSala(c *fiber.Ctx) error
	PausarTiempoUsoSala(c *fiber.Ctx) error
	ReanudarTiempoUsoSala(c *fiber.Ctx) error
}

type SalaHandlerWS interface {
	UsoSalas(c *websocket.Conn)
	UsoSala(c *websocket.Conn)
	UsoSalasBySucursalId(c *websocket.Conn)
}
