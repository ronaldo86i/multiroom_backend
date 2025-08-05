package setup

import (
	"github.com/joho/godotenv"
	"log"
	httpHandler "multiroom/dispositivo-service/internal/adapter/handler/http"
	wsHandler "multiroom/dispositivo-service/internal/adapter/handler/websocket"
	"multiroom/dispositivo-service/internal/adapter/repository"
	"multiroom/dispositivo-service/internal/core/port"
	"multiroom/dispositivo-service/internal/core/service"
	"multiroom/dispositivo-service/internal/postgresql"
	"os"

	"sync"
)

type Repository struct {
	Dispositivo port.DispositivoRepository
}

type Service struct {
	Dispositivo port.DispositivoService
	RabbitMQ    port.RabbitMQService
}

type Handler struct {
	Dispositivo   port.DispositivoHandler
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

		// Services
		services.Dispositivo = service.NewDispositivoService(repositories.Dispositivo)
		services.RabbitMQ = service.NewRabbitMQService(os.Getenv("RABBITMQ_URL"))
		// Handlers
		handlers.Dispositivo = httpHandler.NewDispositivoHandler(services.Dispositivo, services.RabbitMQ)
		handlers.DispositivoWS = wsHandler.NewDispositivoHandlerWS(services.Dispositivo, services.RabbitMQ)
		instance = d
	})
}
