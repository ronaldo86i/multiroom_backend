package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/util"
	"net/http"
	"os"
	"strings"
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

func VerifyUser(c *fiber.Ctx) error {
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
	url := fmt.Sprintf("http://%s:%s/api/v1/auth/verify", service1, httpPort1)
	log.Println(url)
	agent := fiber.Get(url)
	agent.Set("Authorization", authHeader)

	// Hacer la solicitud y obtener respuesta
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		log.Println("Errores en la solicitud:", errs)
		return c.Status(http.StatusInternalServerError).JSON(util.NewMessage("Error al verificar token"))
	}
	log.Println(statusCode, string(body))
	if statusCode != http.StatusOK {
		return c.Status(statusCode).SendString(string(body))
	}
	var user domain.MessageData[domain.Usuario]
	err := json.Unmarshal(body, &user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(util.NewMessage("Error el usuario"))
	}
	// Guardar en contexto y locals
	ctx := context.WithValue(c.UserContext(), util.ContextUserIdKey, user.Data.Id)
	c.SetUserContext(ctx)

	c.Locals(util.ContextUserIdKey, user.Data.Id)
	return c.Next()
}
