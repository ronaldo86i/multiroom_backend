package server

import (
	"github.com/gofiber/contrib/websocket"
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
	v1 := api.Group("/v1")            // Versión 1 de la API
	s.endPointsApiUsuarioSucursal(v1) // Inicializar endpoints de administrador
	// path: /api/v1/paises
	v1Paises := v1.Group("/paises")
	v1Paises.Use(middleware.HostnameMiddleware)
	v1Paises.Get("", middleware.VerifyUsuarioAdmin("ADMIN", "EMPLEADO"), s.handlers.Pais.ObtenerListaPaises)
	v1Paises.Get("/:paisId", middleware.VerifyUsuarioAdmin("ADMIN", "EMPLEADO"), s.handlers.Pais.ObtenerPaisById)
	v1Paises.Post("", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Pais.RegistrarPais)
	v1Paises.Put("/:paisId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Pais.ModificarPais)
	v1Paises.Patch("/:paisId/deshabilitar", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Pais.DeshabilitarPaisById)
	v1Paises.Patch("/:paisId/habilitar", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Pais.HabilitarPaisById)

	//  path: /api/v1/sucursales
	v1Sucursales := v1.Group("/sucursales")
	v1Sucursales.Use(middleware.HostnameMiddleware)
	v1Sucursales.Get("", middleware.VerifyUsuarioAdmin("ADMIN", "EMPLEADO"), s.handlers.Sucursal.ObtenerListaSucursales)
	v1Sucursales.Get("/:sucursalId", middleware.VerifyUsuarioAdmin("ADMIN", "EMPLEADO"), s.handlers.Sucursal.ObtenerSucursalById)
	v1Sucursales.Post("", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sucursal.RegistrarSucursal)
	v1Sucursales.Put("/:sucursalId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sucursal.ModificarSucursal)
	v1Sucursales.Patch("/:sucursalId/habilitar", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sucursal.HabilitarSucursal)
	v1Sucursales.Patch("/:sucursalId/deshabilitar", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sucursal.DeshabilitarSucursal)

	// path: /api/v1/salas
	v1Salas := v1.Group("/salas")
	v1Salas.Use(middleware.HostnameMiddleware)
	v1Salas.Get("", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.ObtenerListaSalas)
	v1Salas.Get("/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.ObtenerSalaById)
	v1Salas.Post("", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.RegistrarSala)
	v1Salas.Put("/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.ModificarSala)
	v1Salas.Patch("/:salaId/habilitar", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.HabilitarSala)
	v1Salas.Patch("/:salaId/deshabilitar", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.DeshabilitarSala)

	v1AccionesSalas := v1.Group("/acciones/salas")
	v1AccionesSalas.Use(middleware.HostnameMiddleware)
	v1AccionesSalas.Post("", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.AsignarTiempoUsoSala)
	v1AccionesSalas.Patch("/cancelar/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.CancelarSala)
	v1AccionesSalas.Patch("/pausar/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.PausarTiempoUsoSala)
	v1AccionesSalas.Patch("/reanudar/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.ReanudarTiempoUsoSala)
	v1AccionesSalas.Patch("/incrementar/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), s.handlers.Sala.IncrementarTiempoUsoSala)
}

func (s *Server) endPointsApiUsuarioSucursal(api fiber.Router) {
	v1UsuarioSucursal := api.Group("/usuario-sucursal")
	v1UsuarioSucursal.Use(middleware.HostnameMiddleware)

	v1Salas := v1UsuarioSucursal.Group("/salas")
	v1Salas.Get("", middleware.VerifyUsuarioSucursal, s.handlers.Sala.ObtenerListaSalas)
	v1Salas.Get("/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.ObtenerSalaById)

	v1AccionesSalas := v1UsuarioSucursal.Group("/acciones/salas")
	v1AccionesSalas.Use(middleware.HostnameMiddleware)
	v1AccionesSalas.Post("", middleware.VerifyUsuarioSucursal, s.handlers.Sala.AsignarTiempoUsoSala)
	v1AccionesSalas.Patch("/cancelar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.CancelarSala)
	v1AccionesSalas.Patch("/pausar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.PausarTiempoUsoSala)
	v1AccionesSalas.Patch("/reanudar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.ReanudarTiempoUsoSala)
	v1AccionesSalas.Patch("/incrementar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.IncrementarTiempoUsoSala)
}

func (s *Server) initEndPointsWS(app *fiber.App) {

	ws := app.Group("/ws") // Crear un grupo para las rutas de la API
	// Inicializar los endpoints de la API
	v1 := ws.Group("/v1")
	s.endPointsWS(v1)
	s.endPointsWSUsuarioSucursal(v1)

}

func (s *Server) endPointsWS(api fiber.Router) {
	// path: /ws/v1/salas
	v1Salas := api.Group("/salas")
	v1Salas.Get("", middleware.VerifyUsuarioAdmin("ADMIN"), websocket.New(s.handlers.SalaWS.UsoSalas))
	v1Salas.Get("/:salaId", middleware.VerifyUsuarioAdmin("ADMIN"), websocket.New(s.handlers.SalaWS.UsoSala))
}

func (s *Server) endPointsWSUsuarioSucursal(api fiber.Router) {
	// path: /ws/v1/usuario-sucursal
	v1UsuarioSucursal := api.Group("/usuario-sucursal")

	// path: /ws/v1/usuario-sucursal/salas
	v1Salas := v1UsuarioSucursal.Group("/salas")
	v1Salas.Get("", middleware.VerifyUsuarioSucursal, websocket.New(s.handlers.SalaWS.UsoSalasBySucursalId))
}
