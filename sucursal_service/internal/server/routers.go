package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"multiroom/sucursal-service/internal/core/util"
	"multiroom/sucursal-service/internal/server/middleware"
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
	app.Static("/", "./public", fiber.Static{
		ModifyResponse: util.EnviarArchivo,
	})
	api := app.Group("/api") // Crear un grupo para las rutas de la API
	// Inicializar los endpoints de la API
	s.endPointsAPI(api)
}

func (s *Server) endPointsAPI(api fiber.Router) {

	v1 := api.Group("/v1") // Versión 1 de la API
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server running")
	})

	// path: /api/v1/paises
	v1Paises := v1.Group("/paises")
	v1Paises.Use(middleware.HostnameMiddleware)
	v1Paises.Get("", s.handlers.Pais.ObtenerListaPaises)
	v1Paises.Get("/:paisId", s.handlers.Pais.ObtenerPaisById)
	v1Paises.Post("", s.handlers.Pais.RegistrarPais)
	v1Paises.Put("/:paisId", s.handlers.Pais.ModificarPais)
	v1Paises.Patch("/:paisId/deshabilitar", s.handlers.Pais.DeshabilitarPaisById)
	v1Paises.Patch("/:paisId/habilitar", s.handlers.Pais.HabilitarPaisById)

}
