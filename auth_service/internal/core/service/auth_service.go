package service

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
	"multiroom/auth-service/internal/core/util"
	"time"
)

type AuthService struct {
	usuarioRepository port.UsuarioRepository
}

func (a AuthService) VerificarUsuario(ctx context.Context, token string) (*domain.Usuario, error) {
	log.Println(token)
	claimsAccessToken, err := util.Token.VerifyToken(token, "access-token-app")
	if err != nil {
		return nil, err
	}
	// Extraer userId
	userIdFloat, ok := claimsAccessToken["userId"].(float64)
	if !ok {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	userId := int(userIdFloat)
	user, err := a.usuarioRepository.ObtenerUsuarioById(ctx, &userId)
	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	if user.Estado == "Inactivo" {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	return user, nil
}

func (a AuthService) Login(ctx context.Context, request *domain.LoginRequest) (*domain.TokenResponse, error) {
	user, err := a.usuarioRepository.ObtenerUsuarioByUsername(ctx, &request.Username)
	if err != nil {
		return nil, err
	}
	log.Println("Usuario:", user.Username)
	// Verificamos la contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}

	// Generamos el token si la contraseña es correcta
	expAccess := time.Now().UTC().Add(7 * 24 * time.Hour)
	accessToken, err := util.Token.CreateToken(jwt.MapClaims{
		"userId":     user.Id,
		"expiration": expAccess.Unix(),
		"type":       "access-token-app",
	})

	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}

	return &domain.TokenResponse{
		Token:     accessToken,
		ExpiresIn: expAccess.Unix(),
	}, nil
}

func NewAuthService(usuarioRepository port.UsuarioRepository) *AuthService {
	return &AuthService{usuarioRepository: usuarioRepository}
}

var _ port.AuthService = (*AuthService)(nil)
