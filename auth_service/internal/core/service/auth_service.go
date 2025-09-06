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
	usuarioRepository         port.UsuarioRepository
	usuarioAdminRepository    port.UsuarioAdminRepository
	usuarioSucursalRepository port.UsuarioSucursalRepository
}

func (a AuthService) LoginSucursal(ctx context.Context, request *domain.LoginSucursalRequest) (*domain.TokenResponse[domain.UsuarioSucursal], error) {
	user, err := a.usuarioSucursalRepository.ObtenerUsuarioSucursalByUsername(ctx, &request.Username)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	log.Println("Usuario:", user.Username)
	if request.SucursalId != user.Sucursal.Id {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	if user.Estado == "Inactivo" {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	// Verificamos la contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}

	// Generamos el token si la contraseña es correcta
	expAccess := time.Now().UTC().Add(7 * 24 * time.Hour)
	accessToken, err := util.Token.CreateToken(jwt.MapClaims{
		"userId":     user.Id,
		"expiration": expAccess.Unix(),
		"type":       "access-token-sucursal",
	})

	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}

	return &domain.TokenResponse[domain.UsuarioSucursal]{
		Token:     accessToken,
		ExpiresIn: expAccess.Unix(),
		Data:      *user,
	}, nil
}

func (a AuthService) VerificarUsuarioSucursal(ctx context.Context, token string) (*domain.UsuarioSucursal, error) {
	claimsAccessToken, err := util.Token.VerifyToken(token, "access-token-sucursal")
	if err != nil {
		return nil, err
	}
	// Extraer userId
	userIdFloat, ok := claimsAccessToken["userId"].(float64)
	if !ok {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	userId := int(userIdFloat)
	user, err := a.usuarioSucursalRepository.ObtenerUsuarioSucursalById(ctx, &userId)
	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	if user.Estado == "Inactivo" {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	return user, nil
}

func (a AuthService) LoginAdmin(ctx context.Context, request *domain.LoginAdminRequest) (*domain.TokenResponse[domain.UsuarioAdmin], error) {
	user, err := a.usuarioAdminRepository.ObtenerUsuarioAdminByUsername(ctx, &request.Username)
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
		"type":       "access-token-admin",
	})

	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}

	return &domain.TokenResponse[domain.UsuarioAdmin]{
		Token:     accessToken,
		ExpiresIn: expAccess.Unix(),
		Data:      *user,
	}, nil
}

func (a AuthService) VerificarUsuarioAdmin(ctx context.Context, token string) (*domain.UsuarioAdmin, error) {
	claimsAccessToken, err := util.Token.VerifyToken(token, "access-token-admin")
	if err != nil {
		return nil, err
	}
	// Extraer userId
	userIdFloat, ok := claimsAccessToken["userId"].(float64)
	if !ok {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	userId := int(userIdFloat)
	user, err := a.usuarioAdminRepository.ObtenerUsuarioAdminById(ctx, &userId)
	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	if user.Estado == "Inactivo" {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}
	return user, nil
}

func (a AuthService) VerificarUsuario(ctx context.Context, token string) (*domain.Usuario, error) {
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

func (a AuthService) Login(ctx context.Context, request *domain.LoginRequest) (*domain.TokenResponse[domain.Usuario], error) {
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
	expAccess := time.Now().UTC().Add(365 * 24 * time.Hour)
	accessToken, err := util.Token.CreateToken(jwt.MapClaims{
		"userId":     user.Id,
		"expiration": expAccess.Unix(),
		"type":       "access-token-app",
	})

	if err != nil {
		return nil, datatype.NewStatusUnauthorizedError("Usuario no autorizado")
	}

	return &domain.TokenResponse[domain.Usuario]{
		Token:     accessToken,
		ExpiresIn: expAccess.Unix(),
		Data:      *user,
	}, nil
}

func NewAuthService(usuarioRepository port.UsuarioRepository, usuarioAdminRepository port.UsuarioAdminRepository, usuarioSucursalRepository port.UsuarioSucursalRepository) *AuthService {
	return &AuthService{usuarioRepository: usuarioRepository, usuarioAdminRepository: usuarioAdminRepository, usuarioSucursalRepository: usuarioSucursalRepository}
}

var _ port.AuthService = (*AuthService)(nil)
