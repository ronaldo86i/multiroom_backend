package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/util"
	"multiroom/auth-service/internal/server/setup"
	"net/http"
)

// HostnameMiddleware guarda y registra el hostname completo de la petición
func HostnameMiddleware(c *fiber.Ctx) error {
	fullHostname := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())
	log.Printf("Petición recibida desde host: %s", fullHostname)
	// Guardar fullHostname en context
	ctx := context.WithValue(c.UserContext(), util.ContextFullHostnameKey, fullHostname)
	c.SetUserContext(ctx)
	return c.Next()
}

func VerifyUser(c *fiber.Ctx) error {
	tokenString, err := util.Token.GetToken(c.Get("Authorization"))
	if err != nil {
		return err
	}

	claimsAccessToken, err := util.Token.VerifyToken(tokenString, "access-token-app")
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
	}

	// Extraer userId
	userIdFloat, ok := claimsAccessToken["userId"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
	}
	userId := int(userIdFloat)
	// Guardar en el contexto
	ctx := context.WithValue(c.UserContext(), util.ContextUserIdKey, userId)
	c.SetUserContext(ctx)

	// Guardar en local
	c.Locals(util.ContextUserIdKey, userId)

	return c.Next()
}

func VerifyRolesMiddleware(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.UserContext().Value(util.ContextUsernameKey).(string)
		user, err := setup.GetDependencies().Repository.Usuario.ObtenerUsuarioByUsername(c.UserContext(), &username)
		if err != nil {
			log.Print(err.Error())
			var errorResponse *datatype.ErrorResponse
			if errors.As(err, &errorResponse) {
				return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
			}
			return datatype.NewInternalServerErrorGeneric()
		}
		if user.Estado == "Inactivo" {
			return datatype.NewForbiddenError("Usuario no permitido")
		}
		// Verificar si tiene el rol
		//for _, rol := range roles {
		//	for _, userRole := range user.Roles {
		//		if rol == userRole.Nombre {
		//			return c.Next()
		//		}
		//	}
		//}

		return c.Status(http.StatusForbidden).JSON(util.NewMessage("Usuario no permitido"))
	}
}
