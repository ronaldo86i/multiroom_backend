package setup

import (
	"log"
	httpHandler "multiroom/dispositivo-service/internal/adapter/handler/http"
	wsHandler "multiroom/dispositivo-service/internal/adapter/handler/websocket"
	"multiroom/dispositivo-service/internal/adapter/repository"
	"multiroom/dispositivo-service/internal/core/port"
	"multiroom/dispositivo-service/internal/core/service"
	"multiroom/dispositivo-service/internal/postgresql"
	"os"

	"github.com/joho/godotenv"

	"sync"
)

type Repository struct {
	Dispositivo port.DispositivoRepository
	Cliente     port.ClienteRepository
	Sala        port.SalaRepository
}

type Service struct {
	Dispositivo port.DispositivoService
	Cliente     port.ClienteService
	Sala        port.SalaService
	RabbitMQ    port.RabbitMQService
}

type Handler struct {
	Dispositivo   port.DispositivoHandler
	Cliente       port.ClienteHandler
	DispositivoWS port.DispositivoHandlerWS
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
		repositories.Dispositivo = repository.NewDispositivoRepository(pool)
		repositories.Cliente = repository.NewClienteRepository(pool)
		repositories.Sala = repository.NewSalaRepository(pool)

		// Services
		services.Dispositivo = service.NewDispositivoService(repositories.Dispositivo)
		services.Cliente = service.NewClienteService(repositories.Cliente)
		services.Sala = service.NewSalaRepository(repositories.Sala)
		services.RabbitMQ = service.NewRabbitMQService(os.Getenv("RABBITMQ_URL"))

		// Handlers
		handlers.Dispositivo = httpHandler.NewDispositivoHandler(services.Dispositivo, services.RabbitMQ)
		handlers.Cliente = httpHandler.NewClienteHandler(services.Cliente)
		handlers.DispositivoWS = wsHandler.NewDispositivoHandlerWS(services.Dispositivo, services.Sala, services.RabbitMQ)
		instance = d
	})
}
