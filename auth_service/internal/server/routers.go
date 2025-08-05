package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"time"
)

func rateLimiter(max int, expiration, delay time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,        // Intentos máximos
		Expiration: expiration, // Tiempo de expiración
		LimitReached: func(c *fiber.Ctx) error {
			time.Sleep(delay)
			return c.Next()
		},
	})

}

func (s *Server) initEndPoints(app *fiber.App) {

	api := app.Group("/api") // Crear un grupo para las rutas de la API
	// Inicializar los endpoints de la API
	s.endPointsAPI(api)
}

func (s *Server) endPointsAPI(api fiber.Router) {

	v1 := api.Group("/v1") // Versión 1 de la API
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server running")
	})

	// path: /api/v1/usuarios
	v1Usuarios := v1.Group("/usuarios")
	v1Usuarios.Get("", s.handlers.Usuario.ObtenerListaUsuarios)
	v1Usuarios.Get("/:usuarioId", s.handlers.Usuario.ObtenerUsuarioById)
	v1Usuarios.Post("", s.handlers.Usuario.RegistrarUsuario)
	v1Usuarios.Patch("/:usuarioId/deshabilitar", s.handlers.Usuario.DeshabilitarUsuario)
	v1Usuarios.Patch("/:usuarioId/habilitar", s.handlers.Usuario.HabilitarUsuario)

	// path: /api/v1/auth
	v1Auth := v1.Group("/auth")
	v1Auth.Post("/login", s.handlers.Auth.Login)
	v1Auth.Get("/verify", s.handlers.Auth.VerificarUsuario)
}
