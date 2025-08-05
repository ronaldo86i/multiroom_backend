package server

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"multiroom/dispositivo-service/internal/server/middleware"
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

func (s *Server) initEndPointsHTTP(app *fiber.App) {

	api := app.Group("/api") // Crear un grupo para las rutas de la API
	// Inicializar los endpoints de la API
	s.endPointsAPI(api)
}

func (s *Server) endPointsAPI(api fiber.Router) {

	v1 := api.Group("/v1") // Versión 1 de la API
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server running")
	})

	// path: /api/v1/dispositivos
	v1Dispositivos := v1.Group("/dispositivos")
	v1Dispositivos.Get("", s.handlers.Dispositivo.ObtenerListaDispositivos)
	v1Dispositivos.Post("", s.handlers.Dispositivo.RegistrarDispositivo)
	v1Dispositivos.Patch("/:dispositivoId/deshabilitar", s.handlers.Dispositivo.DeshabilitarDispositivo)
	v1Dispositivos.Patch("/:dispositivoId/habilitar", s.handlers.Dispositivo.HabilitarDispositivo)

	// path: /api/v1/usuarios
	v1Usuarios := v1.Group("/usuarios")
	v1Usuarios.Get("/:usuarioId/dispositivos", middleware.VerifyUser, s.handlers.Dispositivo.ObtenerListaDispositivosByUsuarioId)
}

func (s *Server) initEndPointsWS(app *fiber.App) {

	ws := app.Group("/ws") // Crear un grupo para las rutas de la API
	// Inicializar los endpoints de la API
	s.endPointsWS(ws)
}

func (s *Server) endPointsWS(api fiber.Router) {
	v1 := api.Group("/v1")

	// path: /ws/v1/dispositivos
	v1Dispositivos := v1.Group("/dispositivos")
	v1Dispositivos.Get("/usuario/me", middleware.VerifyUser, websocket.New(s.handlers.DispositivoWS.NotificarDispositivoHabilitar))
}
