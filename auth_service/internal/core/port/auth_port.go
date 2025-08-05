package port

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"multiroom/auth-service/internal/core/domain"
)

type AuthService interface {
	Login(ctx context.Context, request *domain.LoginRequest) (*domain.TokenResponse, error)
	VerificarUsuario(ctx context.Context, token string) (*domain.Usuario, error)
}

type AuthHandler interface {
	Login(c *fiber.Ctx) error
	VerificarUsuario(c *fiber.Ctx) error
}
