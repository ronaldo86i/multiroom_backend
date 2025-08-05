package setup

import (
	"github.com/joho/godotenv"
	"log"
	httpHandler "multiroom/sucursal-service/internal/adapter/handler/http"
	"multiroom/sucursal-service/internal/adapter/repository"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/service"

	//wsHandler "multiroom/sucursal-service/internal/adapter/handler/websocket"
	"multiroom/sucursal-service/internal/postgresql"
	"sync"
)

type Repository struct {
	Pais port.PaisRepository
}

type Service struct {
	Pais port.PaisService
}

type Handler struct {
	Pais port.PaisHandler
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
		// Services
		services.Pais = service.NewPaisService(repositories.Pais)
		// Handlers
		handlers.Pais = httpHandler.NewPaisHandler(services.Pais)
		instance = d
	})
}
