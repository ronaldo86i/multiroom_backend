package server

import (
	"multiroom/dispositivo-service/internal/server/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

//func rateLimiter(max int, expiration, delay time.Duration) fiber.Handler {
//	return limiter.New(limiter.Config{
//		Max:        max,        // Intentos máximos
//		Expiration: expiration, // Tiempo de expiración
//		LimitReached: func(c *fiber.Ctx) error {
//			time.Sleep(delay)
//			return c.Next()
//		},
//	})
//
//}

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
	v1Dispositivos.Get("", middleware.VerifyPermission("dispositivo:ver"), s.handlers.Dispositivo.ObtenerListaDispositivos)
	v1Dispositivos.Get("/byDispositivoId/:dispositivoId", middleware.VerifyUser, s.handlers.Dispositivo.ObtenerDispositivoByDispositivoId)
	v1Dispositivos.Post("", middleware.VerifyUser, s.handlers.Dispositivo.RegistrarDispositivo)
	v1Dispositivos.Patch("/:dispositivoId/habilitar", middleware.VerifyPermission("dispositivo:editar"), s.handlers.Dispositivo.HabilitarDispositivo)
	v1Dispositivos.Patch("/:dispositivoId/deshabilitar", middleware.VerifyPermission("dispositivo:editar"), s.handlers.Dispositivo.DeshabilitarDispositivo)
	v1Dispositivos.Delete("/:dispositivoId/eliminar", middleware.VerifyPermission("dispositivo:eliminar"), s.handlers.Dispositivo.EliminarDispositivoById)
	// path: /api/v1/usuarios
	v1Usuarios := v1.Group("/usuarios")
	v1Usuarios.Get("/:usuarioId/dispositivos", middleware.VerifyUser, s.handlers.Dispositivo.ObtenerListaDispositivosByUsuarioId)

	v1Clientes := v1.Group("/clientes")
	v1Clientes.Get("", middleware.VerifyPermission("cliente:ver"), s.handlers.Cliente.ObtenerListaClientes)
	v1Clientes.Get("/:clienteId", middleware.VerifyPermission("cliente:ver"), s.handlers.Cliente.ObtenerClienteDetailById)
	v1Clientes.Post("", middleware.VerifyPermission("cliente:crear"), s.handlers.Cliente.RegistrarCliente)
	v1Clientes.Put("/:clienteId", middleware.VerifyPermission("cliente:editar"), s.handlers.Cliente.ModificarCliente)
	v1Clientes.Patch("/:clienteId/habilitar", middleware.VerifyPermission("cliente:editar"), s.handlers.Cliente.HabilitarCliente)
	v1Clientes.Patch("/:clienteId/deshabilitar", middleware.VerifyPermission("cliente:editar"), s.handlers.Cliente.DeshabilitarCliente)
	v1Clientes.Delete("/:clienteId", middleware.VerifyPermission("cliente:eliminar"), s.handlers.Cliente.EliminarClienteById)
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
