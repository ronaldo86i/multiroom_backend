package middleware

import (
	"context"
	"fmt"
	"log"
	"multiroom/auth-service/internal/core/util"
	"multiroom/auth-service/internal/server/setup"

	"github.com/gofiber/fiber/v2"
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
	log.Println(claimsAccessToken)
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

func VerifyPermission(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Recuperar id de usuario
		val := c.Locals(util.ContextUserIdKey)
		if val == nil {
			log.Println("🔴 DEBUG: ContextUserIdKey es nil")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": true, "message": "Usuario no identificado"})
		}

		log.Printf("🔍 DEBUG ID TYPE: %T | VALUE: %v", val, val)

		var userId int
		switch v := val.(type) {
		case int:
			userId = v
		case float64:
			userId = int(v)
		default:
			log.Printf("🔴 DEBUG: Tipo de ID no soportado: %T", val)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "ID inválido"})
		}
		// Obtener dependencias
		deps := setup.GetDependencies()
		// Llamar a servicio para recuperar usuario
		user, err := deps.Service.UsuarioAdmin.ObtenerUsuarioAdminById(context.Background(), &userId)

		if err != nil {
			log.Printf("🔴 DEBUG ERROR SERVICE: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": true, "message": "Error servicio"})
		}

		if user == nil {
			log.Println("🔴 DEBUG: Usuario retornó nil")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": true, "message": "Usuario no encontrado"})
		}

		permisoEncontrado := false

		// Verificamos permiso con la lista de permisos del usuario
		for _, p := range user.Permisos {
			if p.Nombre == requiredPermission {
				log.Printf("✅ DEBUG: Permiso encontrado: %s", p.Nombre)
				permisoEncontrado = true
				break
			}
		}

		if permisoEncontrado {
			return c.Next()
		}

		// Permiso no encontrado
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "No tienes permisos suficientes (" + requiredPermission + ")",
		})
	}
}
