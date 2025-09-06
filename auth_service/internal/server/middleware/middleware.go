package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/util"
	"multiroom/auth-service/internal/server/setup"
	"net/http"
)

// HostnameMiddleware guarda y registra el hostname completo de la petición
//func HostnameMiddleware(c *fiber.Ctx) error {
//	fullHostname := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())
//	log.Printf("Petición recibida desde host: %s", fullHostname)
//	// Guardar fullHostname en context
//	ctx := context.WithValue(c.UserContext(), util.ContextFullHostnameKey, fullHostname)
//	c.SetUserContext(ctx)
//	return c.Next()
//}

//func VerifyUser(c *fiber.Ctx) error {
//	tokenString, err := util.Token.GetToken(c.Get("Authorization"))
//	if err != nil {
//		return err
//	}
//
//	claimsAccessToken, err := util.Token.VerifyToken(tokenString, "access-token-app")
//	if err != nil {
//		return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
//	}
//
//	// Extraer userId
//	userIdFloat, ok := claimsAccessToken["userId"].(float64)
//	if !ok {
//		return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
//	}
//	userId := int(userIdFloat)
//	// Guardar en el contexto
//	ctx := context.WithValue(c.UserContext(), util.ContextUserIdKey, userId)
//	c.SetUserContext(ctx)
//
//	// Guardar en local
//	c.Locals(util.ContextUserIdKey, userId)
//
//	return c.Next()
//}

func VerifyUsuarioAdmin(c *fiber.Ctx) error {
	tokenString, err := util.Token.GetToken(c.Get("Authorization"))
	if err != nil {
		log.Println("Error al obtener token", err)
		return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
	}

	claimsAccessToken, err := util.Token.VerifyToken(tokenString, "access-token-admin")
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
		userId, ok := c.UserContext().Value(util.ContextUserIdKey).(int)
		if !ok {
			log.Println("Sin userId en Context:", ok)
			return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
		}
		user, err := setup.GetDependencies().Repository.UsuarioAdmin.ObtenerUsuarioAdminById(c.UserContext(), &userId)
		if err != nil {
			log.Print("Error al obtener usuario:", err)
			var errorResponse *datatype.ErrorResponse
			if errors.As(err, &errorResponse) {
				return c.Status(errorResponse.Code).JSON(util.NewMessage(errorResponse.Message))
			}
			return c.Status(http.StatusInternalServerError).JSON(util.NewMessage("Error interno en el servidor"))
		}

		// 🔹 Imprimir usuario en JSON (solo para debug)
		if b, err := json.MarshalIndent(user, "", "  "); err == nil {
			log.Printf("Usuario autenticado:\n%s", string(b))
		} else {
			log.Printf("Error al serializar usuario: %v", err)
		}

		//Verificar si tiene el rol
		roleMap := make(map[string]struct{}, len(user.Roles))
		for _, r := range user.Roles {
			roleMap[r.Nombre] = struct{}{}
		}
		for _, rol := range roles {
			if _, ok := roleMap[rol]; ok {
				return c.Next()
			}
		}

		return c.Status(http.StatusForbidden).JSON(util.NewMessage("Usuario no permitido"))
	}
}
