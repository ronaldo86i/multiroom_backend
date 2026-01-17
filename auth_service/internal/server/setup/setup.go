package setup

import (
	"log"
	"multiroom/auth-service/internal/adapter/handler"
	"multiroom/auth-service/internal/adapter/repository"
	"multiroom/auth-service/internal/core/port"
	"multiroom/auth-service/internal/core/service"
	"multiroom/auth-service/internal/postgresql"
	"sync"

	"github.com/joho/godotenv"
)

type Repository struct {
	Usuario         port.UsuarioRepository
	UsuarioAdmin    port.UsuarioAdminRepository
	UsuarioSucursal port.UsuarioSucursalRepository
	Permiso         port.PermisoRepository
	Rol             port.RolRepository
}

type Service struct {
	Auth            port.AuthService
	Usuario         port.UsuarioService
	UsuarioAdmin    port.UsuarioAdminService
	UsuarioSucursal port.UsuarioSucursalService
	Permiso         port.PermisoService
	Rol             port.RolService
}

type Handler struct {
	Auth            port.AuthHandler
	Usuario         port.UsuarioHandler
	UsuarioAdmin    port.UsuarioAdminHandler
	UsuarioSucursal port.UsuarioSucursalHandler
	Permiso         port.PermisoHandler
	Rol             port.RolHandler
}

type Dependencies struct {
	Repository Repository
	Service    Service
	Handler    Handler
}

var (
	instance *Dependencies
	once     sync.Once
)

func GetDependencies() *Dependencies {
	return instance
}

func initEnv(filenames ...string) error {
	err := godotenv.Load(filenames...)
	if err != nil {
		return err
	}
	return nil
}

func initDB() error {
	err := postgresql.Connection()
	if err != nil {
		return err
	}
	return nil
}

func Init() {
	once.Do(func() {
		if err := initEnv(".env"); err != nil {
			log.Fatalf("Fallo al inicializar variables de entorno: %v", err)
		}

		if err := initDB(); err != nil {
			log.Fatalf("Fallo en conectar a la base de datos: %v", err)
		}
		var pool = postgresql.GetDB()
		d := &Dependencies{}
		repositories := &d.Repository
		services := &d.Service
		handlers := &d.Handler

		// Repositories
		repositories.Usuario = repository.NewUsuarioRepository(pool)
		repositories.UsuarioAdmin = repository.NewUsuarioAdminRepository(pool)
		repositories.UsuarioSucursal = repository.NewUsuarioSucursalRepository(pool)
		repositories.Permiso = repository.NewPermisoRepository(pool)
		repositories.Rol = repository.NewRolRepository(pool)
		// Services
		services.Auth = service.NewAuthService(repositories.Usuario, repositories.UsuarioAdmin, repositories.UsuarioSucursal)
		services.Usuario = service.NewUsuarioService(repositories.Usuario)
		services.UsuarioAdmin = service.NewUsuarioAdminService(repositories.UsuarioAdmin)
		services.UsuarioSucursal = service.NewUsuarioSucursalService(repositories.UsuarioSucursal)
		services.Permiso = service.NewPermisoService(repositories.Permiso)
		services.Rol = service.NewRolService(repositories.Rol)
		// Handlers
		handlers.Auth = handler.NewAuthHandler(services.Auth)
		handlers.Usuario = handler.NewUsuarioHandler(services.Usuario)
		handlers.UsuarioAdmin = handler.NewUsuarioAdminHandler(services.UsuarioAdmin)
		handlers.UsuarioSucursal = handler.NewUsuarioSucursalHandler(services.UsuarioSucursal)
		handlers.Permiso = handler.NewPermisoHandler(services.Permiso)
		handlers.Rol = handler.NewRolHandler(services.Rol)
		instance = d
	})
}
