package middleware

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"multiroom/sucursal-service/internal/core/util"
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
