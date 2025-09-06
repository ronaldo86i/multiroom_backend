package setup

import (
	"github.com/joho/godotenv"
	"log"
	"multiroom/auth-service/internal/adapter/handler"
	"multiroom/auth-service/internal/adapter/repository"
	"multiroom/auth-service/internal/core/port"
	"multiroom/auth-service/internal/core/service"
	"multiroom/auth-service/internal/postgresql"
	"sync"
)

type Repository struct {
	Usuario         port.UsuarioRepository
	UsuarioAdmin    port.UsuarioAdminRepository
	UsuarioSucursal port.UsuarioSucursalRepository
}

type Service struct {
	Auth            port.AuthService
	Usuario         port.UsuarioService
	UsuarioAdmin    port.UsuarioAdminService
	UsuarioSucursal port.UsuarioSucursalService
}

type Handler struct {
	Auth            port.AuthHandler
	Usuario         port.UsuarioHandler
	UsuarioSucursal port.UsuarioSucursalHandler
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

		// Services
		services.Auth = service.NewAuthService(repositories.Usuario, repositories.UsuarioAdmin, repositories.UsuarioSucursal)
		services.Usuario = service.NewUsuarioService(repositories.Usuario)
		services.UsuarioAdmin = service.NewUsuarioAdminService(repositories.UsuarioAdmin)
		services.UsuarioSucursal = service.NewUsuarioSucursalService(repositories.UsuarioSucursal)
		// Handlers
		handlers.Auth = handler.NewAuthHandler(services.Auth)
		handlers.Usuario = handler.NewUsuarioHandler(services.Usuario)
		handlers.UsuarioSucursal = handler.NewUsuarioSucursalHandler(services.UsuarioSucursal)

		instance = d
	})
}
