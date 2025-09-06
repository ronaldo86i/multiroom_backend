package port

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"multiroom/auth-service/internal/core/domain"
)

type AuthService interface {
	/*Usuario de aplicación móvil*/

	Login(ctx context.Context, request *domain.LoginRequest) (*domain.TokenResponse[domain.Usuario], error)
	VerificarUsuario(ctx context.Context, token string) (*domain.Usuario, error)

	/*Usuario administrador*/

	LoginAdmin(ctx context.Context, request *domain.LoginAdminRequest) (*domain.TokenResponse[domain.UsuarioAdmin], error)
	VerificarUsuarioAdmin(ctx context.Context, token string) (*domain.UsuarioAdmin, error)

	/* Usuario sucursal*/

	LoginSucursal(ctx context.Context, request *domain.LoginSucursalRequest) (*domain.TokenResponse[domain.UsuarioSucursal], error)
	VerificarUsuarioSucursal(ctx context.Context, token string) (*domain.UsuarioSucursal, error)
}

type AuthHandler interface {
	Login(c *fiber.Ctx) error
	VerificarUsuario(c *fiber.Ctx) error
	LoginAdmin(c *fiber.Ctx) error
	VerificarUsuarioAdmin(c *fiber.Ctx) error
	LoginSucursal(c *fiber.Ctx) error
	VerificarUsuarioSucursal(c *fiber.Ctx) error
}
