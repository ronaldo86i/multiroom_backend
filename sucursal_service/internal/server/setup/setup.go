package setup

import (
	"github.com/joho/godotenv"
	"log"
	httpHandler "multiroom/sucursal-service/internal/adapter/handler/http"
	"multiroom/sucursal-service/internal/adapter/repository"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/service"

	"os"

	wsHandler "multiroom/sucursal-service/internal/adapter/handler/websocket"
	"multiroom/sucursal-service/internal/postgresql"
	"sync"
)

type Repository struct {
	Pais       port.PaisRepository
	Sucursal   port.SucursalRepository
	Sala       port.SalaRepository
	AppVersion port.AppVersionRepository
}

type Service struct {
	Pais       port.PaisService
	Sucursal   port.SucursalService
	Sala       port.SalaService
	RabbitMQ   port.RabbitMQService
	AppVersion port.AppVersionService
}

type Handler struct {
	Pais       port.PaisHandler
	Sucursal   port.SucursalHandler
	Sala       port.SalaHandler
	SalaWS     port.SalaHandlerWS
	AppVersion port.AppVersionHandler
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
		repositories.Pais = repository.NewPaisRepository(pool)
		repositories.Sucursal = repository.NewSucursalRepository(pool)
		repositories.Sala = repository.NewSalaRepository(pool)
		repositories.AppVersion = repository.NewAppVersionRepository(pool)
		// Services
		services.RabbitMQ = service.NewRabbitMQService(os.Getenv("RABBITMQ_URL"))
		services.Pais = service.NewPaisService(repositories.Pais)
		services.Sucursal = service.NewSucursalService(repositories.Sucursal)
		services.Sala = service.NewSalaService(repositories.Sala)
		services.AppVersion = service.NewAppVersionService(repositories.AppVersion)
		// Handlers
		handlers.Pais = httpHandler.NewPaisHandler(services.Pais)
		handlers.Sucursal = httpHandler.NewSucursalHandler(services.Sucursal)
		handlers.SalaWS = wsHandler.NewSalaHandlerWS(services.Sala, services.RabbitMQ)
		handlers.Sala = httpHandler.NewSalaHandler(services.Sala, services.RabbitMQ)
		handlers.AppVersion = httpHandler.NewAppVersionHandler(services.AppVersion)

		instance = d
	})
}
