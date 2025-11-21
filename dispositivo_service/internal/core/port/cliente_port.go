package port

import (
	"context"
	"multiroom/dispositivo-service/internal/core/domain"

	"github.com/gofiber/fiber/v2"
)

type ClienteRepository interface {
	RegistrarCliente(ctx context.Context, request *domain.ClienteRequest) (*int64, error)
	ModificarCliente(ctx context.Context, id *int, request *domain.ClienteRequest) error
	ObtenerListaClientes(ctx context.Context, filtros map[string]string) (*[]domain.ClienteInfo, error)
	ObtenerClienteDetailById(ctx context.Context, id *int) (*domain.ClienteDetail, error)
	HabilitarCliente(ctx context.Context, id *int) error
	DeshabilitarCliente(ctx context.Context, id *int) error
	EliminarClienteById(ctx context.Context, id *int) error
}

type ClienteService interface {
	RegistrarCliente(ctx context.Context, request *domain.ClienteRequest) (*int64, error)
	ModificarCliente(ctx context.Context, id *int, request *domain.ClienteRequest) error
	ObtenerListaClientes(ctx context.Context, filtros map[string]string) (*[]domain.ClienteInfo, error)
	ObtenerClienteDetailById(ctx context.Context, id *int) (*domain.ClienteDetail, error)
	HabilitarCliente(ctx context.Context, id *int) error
	DeshabilitarCliente(ctx context.Context, id *int) error
	EliminarClienteById(ctx context.Context, id *int) error
}

type ClienteHandler interface {
	RegistrarCliente(c *fiber.Ctx) error
	ModificarCliente(c *fiber.Ctx) error
	ObtenerListaClientes(c *fiber.Ctx) error
	ObtenerClienteDetailById(c *fiber.Ctx) error
	HabilitarCliente(c *fiber.Ctx) error
	DeshabilitarCliente(c *fiber.Ctx) error
	EliminarClienteById(c *fiber.Ctx) error
}
