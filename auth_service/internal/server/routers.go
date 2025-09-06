package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"multiroom/auth-service/internal/server/middleware"
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
	v1Usuarios.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.Usuario.ObtenerListaUsuarios)
	v1Usuarios.Get("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.Usuario.ObtenerUsuarioById)
	v1Usuarios.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.Usuario.RegistrarUsuario)
	v1Usuarios.Patch("/:usuarioId/deshabilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.Usuario.DeshabilitarUsuario)
	v1Usuarios.Patch("/:usuarioId/habilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.Usuario.HabilitarUsuario)

	// path: /api/v1/usuariosSucursal
	v1UsuariosSucursal := v1.Group("/usuariosSucursal")
	v1UsuariosSucursal.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.UsuarioSucursal.RegistrarUsuarioSucursal)
	v1UsuariosSucursal.Put("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.UsuarioSucursal.ModificarUsuarioSucursal)
	v1UsuariosSucursal.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.UsuarioSucursal.ObtenerListaUsuariosSucursal)
	v1UsuariosSucursal.Get("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.UsuarioSucursal.ObtenerUsuarioSucursalById)
	v1UsuariosSucursal.Patch("/:usuarioId/deshabilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.UsuarioSucursal.DeshabilitarUsuarioSucursal)
	v1UsuariosSucursal.Patch("/:usuarioId/habilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyRolesMiddleware("ADMIN"), s.handlers.UsuarioSucursal.HabilitarUsuarioSucursal)

	// path: /api/v1/auth
	v1Auth := v1.Group("/auth")
	v1Auth.Post("/login", s.handlers.Auth.Login)
	v1Auth.Post("/admin/login", s.handlers.Auth.LoginAdmin)
	v1Auth.Post("/sucursal/login", s.handlers.Auth.LoginSucursal)

	v1Auth.Get("/verify", s.handlers.Auth.VerificarUsuario)
	v1Auth.Get("/admin/verify", s.handlers.Auth.VerificarUsuarioAdmin)
	v1Auth.Get("/sucursal/verify", s.handlers.Auth.VerificarUsuarioSucursal)
}
