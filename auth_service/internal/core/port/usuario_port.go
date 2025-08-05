package port

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"multiroom/auth-service/internal/core/domain"
)

type UsuarioRepository interface {
	RegistrarUsuario(ctx context.Context, request *domain.UsuarioRequest) (*int, error)
	ObtenerUsuarioById(ctx context.Context, id *int) (*domain.Usuario, error)
	ObtenerListaUsuarios(ctx context.Context) (*[]domain.UsuarioInfo, error)
	DeshabilitarUsuario(ctx context.Context, id *int) error
	HabilitarUsuario(ctx context.Context, id *int) error
	ObtenerUsuarioByUsername(ctx context.Context, username *string) (*domain.Usuario, error)
}

type UsuarioService interface {
	RegistrarUsuario(ctx context.Context, request *domain.UsuarioRequest) (*int, error)
	ObtenerUsuarioById(ctx context.Context, id *int) (*domain.Usuario, error)
	ObtenerListaUsuarios(ctx context.Context) (*[]domain.UsuarioInfo, error)
	DeshabilitarUsuario(ctx context.Context, id *int) error
	HabilitarUsuario(ctx context.Context, id *int) error
}

type UsuarioHandler interface {
	RegistrarUsuario(c *fiber.Ctx) error
	ObtenerUsuarioById(c *fiber.Ctx) error
	ObtenerListaUsuarios(c *fiber.Ctx) error
	DeshabilitarUsuario(c *fiber.Ctx) error
	HabilitarUsuario(c *fiber.Ctx) error
}
