package server

import (
	"multiroom/auth-service/internal/server/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func rateLimiter(max int, expiration time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Demasiadas peticiones. Intente más tarde.",
			})
		},
	})
}

func (s *Server) initEndPoints(app *fiber.App) {
	api := app.Group("/api")
	s.endPointsAPI(api)
}

func (s *Server) endPointsAPI(api fiber.Router) {

	v1 := api.Group("/v1")

	v1.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server running - Permission System Active")
	})

	// =================================================================
	// AUTH (Público + Rate Limiter)
	// =================================================================
	v1Auth := v1.Group("/auth")
	v1Auth.Use(middleware.HostnameMiddleware)
	loginLimiter := rateLimiter(5, 1*time.Minute)

	v1Auth.Post("/login", loginLimiter, s.handlers.Auth.Login)
	v1Auth.Post("/admin/login", loginLimiter, s.handlers.Auth.LoginAdmin)
	v1Auth.Post("/sucursal/login", loginLimiter, s.handlers.Auth.LoginSucursal)

	// Verificaciones
	v1Auth.Get("/verify", middleware.VerifyUser, s.handlers.Auth.VerificarUsuario)
	v1Auth.Get("/admin/verify", middleware.VerifyUsuarioAdmin, s.handlers.Auth.VerificarUsuarioAdmin)

	// GESTIÓN DE ROLES (Recurso: 'rol')
	v1Roles := v1.Group("/roles")
	v1Roles.Use(middleware.HostnameMiddleware)
	v1Roles.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("rol:ver"), s.handlers.Rol.ListarRoles)
	v1Roles.Get("/:rolId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("rol:ver"), s.handlers.Rol.ObtenerRolById)
	v1Roles.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("rol:crear"), s.handlers.Rol.RegistrarRol)
	v1Roles.Put("/:rolId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("rol:editar"), s.handlers.Rol.ModificarRolById)

	// GESTIÓN DE PERMISOS (Recurso: 'permiso') - ¡Cuidado aquí, es meta-seguridad!
	v1Permisos := v1.Group("/permisos")
	v1Permisos.Use(middleware.HostnameMiddleware)
	v1Permisos.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("permiso:ver"), s.handlers.Permiso.ListarPermisos)
	v1Permisos.Get("/:permisoId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("permiso:ver"), s.handlers.Permiso.ObtenerPermisoById)
	// Solo quien tenga 'permiso:gestionar' puede crear o editar permisos del sistema
	v1Permisos.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("permiso:gestionar"), s.handlers.Permiso.RegistrarPermiso)
	v1Permisos.Put("/:permisoId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("permiso:gestionar"), s.handlers.Permiso.ModificarPermisoById)

	// 3. GESTIÓN DE ADMINISTRADORES (Recurso: 'usuario_admin')
	v1UsuariosAdmin := v1.Group("/usuariosAdmin")
	v1UsuariosAdmin.Use(middleware.HostnameMiddleware)
	v1UsuariosAdmin.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario_admin:ver"), s.handlers.UsuarioAdmin.ListarUsuariosAdmin)
	v1UsuariosAdmin.Get("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario_admin:ver"), s.handlers.UsuarioAdmin.ObtenerUsuarioAdminById)
	v1UsuariosAdmin.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario_admin:crear"), s.handlers.UsuarioAdmin.RegistrarUsuarioAdmin)
	v1UsuariosAdmin.Put("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario_admin:editar"), s.handlers.UsuarioAdmin.ModificarUsuarioAdminById)

	// GESTIÓN DE USUARIOS FINALES (Recurso: 'usuario')
	v1Usuarios := v1.Group("/usuarios")
	v1Usuarios.Use(middleware.HostnameMiddleware)
	v1Usuarios.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario:ver"), s.handlers.Usuario.ObtenerListaUsuarios)
	v1Usuarios.Get("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario:ver"), s.handlers.Usuario.ObtenerUsuarioById)
	v1Usuarios.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario:crear"), s.handlers.Usuario.RegistrarUsuario)
	// Acciones delicadas separadas
	v1Usuarios.Patch("/:usuarioId/deshabilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario:editar"), s.handlers.Usuario.DeshabilitarUsuario)
	v1Usuarios.Patch("/:usuarioId/habilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("usuario:editar"), s.handlers.Usuario.HabilitarUsuario)

	// GESTIÓN DE PERSONAL DE SUCURSAL (Recurso: 'personal_sucursal')
	v1UsuariosSucursal := v1.Group("/usuariosSucursal")
	v1UsuariosSucursal.Use(middleware.HostnameMiddleware)
	v1UsuariosSucursal.Get("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("personal_sucursal:ver"), s.handlers.UsuarioSucursal.ObtenerListaUsuariosSucursal)
	v1UsuariosSucursal.Get("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("personal_sucursal:ver"), s.handlers.UsuarioSucursal.ObtenerUsuarioSucursalById)
	v1UsuariosSucursal.Post("", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("personal_sucursal:crear"), s.handlers.UsuarioSucursal.RegistrarUsuarioSucursal)
	v1UsuariosSucursal.Put("/:usuarioId", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("personal_sucursal:editar"), s.handlers.UsuarioSucursal.ModificarUsuarioSucursal)
	v1UsuariosSucursal.Patch("/:usuarioId/deshabilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("personal_sucursal:editar"), s.handlers.UsuarioSucursal.DeshabilitarUsuarioSucursal)
	v1UsuariosSucursal.Patch("/:usuarioId/habilitar", middleware.VerifyUsuarioAdmin, middleware.VerifyPermission("personal_sucursal:editar"), s.handlers.UsuarioSucursal.HabilitarUsuarioSucursal)
}
