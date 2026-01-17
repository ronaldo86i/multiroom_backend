package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/util"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var (
	service1  = "auth"
	httpPort1 = "8080"
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

func VerifyUsuarioSucursal(c *fiber.Ctx) error {
	service1 = os.Getenv("SERVICE_1")
	httpPort1 = os.Getenv("HTTP_PORT_1")
	const bearerPrefix = "Bearer "

	var accessToken string
	authHeader := c.Get("Authorization")

	if websocket.IsWebSocketUpgrade(c) {
		// Si es WebSocket: el token viene en query param
		accessToken = strings.TrimSpace(c.Query("token"))
	} else {
		// Si es HTTP normal: el token viene en header Authorization
		if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
			return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
		}
		accessToken = strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
	}

	if accessToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Token requerido"))
	}

	// Siempre reenviamos con Bearer al servicio de auth
	authHeader = bearerPrefix + accessToken

	// Crear agente HTTP con header Authorization
	url := fmt.Sprintf("http://%s:%s/api/v1/auth/sucursal/verify", service1, httpPort1)
	agent := fiber.Get(url)
	agent.Set("Authorization", authHeader)

	// Hacer la solicitud y obtener respuesta
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		log.Println("Errores en la solicitud:", errs)
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage("Error al verificar token"))
	}

	if statusCode != http.StatusOK {
		return c.Status(statusCode).SendString(string(body))
	}
	var user domain.MessageData[domain.UsuarioSucursal]
	err := json.Unmarshal(body, &user)
	if err != nil {
		log.Println("Error al obtener usuario", err)
		return c.Status(fiber.StatusInternalServerError).JSON(util.NewMessage("Error al obtener usuario"))
	}
	// Guardar en contexto y locals
	ctx := context.WithValue(c.UserContext(), util.ContextUserIdKey, user.Data.Id)
	c.SetUserContext(ctx)

	c.Locals(util.ContextUserIdKey, user.Data.Id)
	c.Locals("sucursalId", user.Data.Sucursal.Id)
	c.Locals("rol", "usuario-sucursal")
	return c.Next()
}

func VerifyPermission(requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Locals(util.ContextUserIdKey) != nil {
			return c.Next()
		}
		service1 = os.Getenv("SERVICE_1")
		httpPort1 = os.Getenv("HTTP_PORT_1")
		const bearerPrefix = "Bearer "
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, bearerPrefix) {
			return c.Status(fiber.StatusUnauthorized).JSON(util.NewMessage("Usuario no autorizado"))
		}
		accessToken := strings.TrimPrefix(authHeader, bearerPrefix)
		accessToken = strings.TrimSpace(accessToken)

		// Crear agente HTTP con header Authorization
		url := fmt.Sprintf("http://%s:%s/api/v1/auth/admin/verify", service1, httpPort1)
		agent := fiber.Get(url)
		agent.Set("Authorization", authHeader)

		// Hacer la solicitud y obtener respuesta
		statusCode, body, errs := agent.Bytes()
		if len(errs) > 0 {
			log.Println("Errores en la solicitud:", errs)
			return c.Status(http.StatusInternalServerError).JSON(util.NewMessage("Error al verificar token"))
		}
		if statusCode != http.StatusOK {
			return c.Status(statusCode).SendString(string(body))
		}
		var user domain.MessageData[domain.UsuarioAdmin]
		err := json.Unmarshal(body, &user)
		if err != nil {
			log.Println("Error al obtener usuario administrador", err)
			return c.Status(fiber.StatusInternalServerError).JSON(util.NewMessage("Error al obtener usuario"))
		}
		// Guardar en contexto y locals
		ctx := context.WithValue(c.UserContext(), util.ContextUserIdKey, user.Data.Id)
		c.SetUserContext(ctx)

		c.Locals(util.ContextUserIdKey, user.Data.Id)

		// Asumiendo que user.Permisos es un Slice de structs (lo más común en Go):
		permisoEncontrado := false
		// Nota: Ajusta 'p.Nombre' según cómo se llame el campo en tu struct domain.PermisoInfo
		for _, p := range user.Data.Permisos {
			if p.Nombre == requiredPermission {
				permisoEncontrado = true
				break
			}
		}

		if permisoEncontrado {
			return c.Next()
		}

		return c.Next()
	}
}
