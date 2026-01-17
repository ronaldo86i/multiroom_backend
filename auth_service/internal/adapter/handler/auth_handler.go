package handler

import (
	"errors"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
	"multiroom/auth-service/internal/core/util"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService port.AuthService
}

func (a AuthHandler) LoginSucursal(c *fiber.Ctx) error {
	var request domain.LoginSucursalRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}
	tokenResponse, err := a.authService.LoginSucursal(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(tokenResponse)
}

func (a AuthHandler) VerificarUsuarioSucursal(c *fiber.Ctx) error {
	tokenString, err := util.Token.GetToken(c.Get("Authorization"))
	user, err := a.authService.VerificarUsuarioSucursal(c.UserContext(), tokenString)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessageData(*user, "Usuario autenticado"))
}

func (a AuthHandler) LoginAdmin(c *fiber.Ctx) error {
	var request domain.LoginAdminRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	tokenResponse, err := a.authService.LoginAdmin(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(tokenResponse)
}

func (a AuthHandler) VerificarUsuarioAdmin(c *fiber.Ctx) error {
	tokenString, err := util.Token.GetToken(c.Get("Authorization"))
	user, err := a.authService.VerificarUsuarioAdmin(c.UserContext(), tokenString)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessageData(*user, "Usuario autenticado"))
}

func (a AuthHandler) VerificarUsuario(c *fiber.Ctx) error {
	tokenString, err := util.Token.GetToken(c.Get("Authorization"))
	user, err := a.authService.VerificarUsuario(c.UserContext(), tokenString)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.JSON(util.NewMessageData(*user, "Usuario autenticado"))
}

func (a AuthHandler) Login(c *fiber.Ctx) error {
	var request domain.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(util.NewMessage("Petición inválida: datos incompletos o incorrectos"))
	}

	tokenResponse, err := a.authService.Login(c.UserContext(), &request)
	if err != nil {
		log.Print(err.Error())
		var errorResponse *datatype.ErrorResponse
		if errors.As(err, &errorResponse) {
			return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
		}
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage(err.Error()))
	}
	return c.Status(http.StatusOK).JSON(tokenResponse)
}

func NewAuthHandler(authService port.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

var _ port.AuthHandler = (*AuthHandler)(nil)
