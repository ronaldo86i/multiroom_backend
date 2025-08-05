package server

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
	"log/slog"
	"multiroom/auth-service/internal/server/setup"
	"os"
	"strings"
	"time"
)

var httpPort = "8080"

type Server struct {
	handlers setup.Handler
}

func NewServer(
	handlers setup.Handler,
) *Server {
	return &Server{
		handlers,
	}
}

func (s *Server) startServer() {
	app := fiber.New(fiber.Config{
		BodyLimit:             20 << 23,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           30 * time.Second,
		DisableStartupMessage: true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Prefork:               false,
		AppName:               "Miltiroom Backend",
		//ErrorHandler:          util.ErrorHandler,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: strings.Join([]string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAuthorization,
			fiber.HeaderXDownloadOptions,
			fiber.HeaderReferrerPolicy,
		}, ","),
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE",
		AllowCredentials: false,
	}))

	app.Use(logger.New(logger.Config{
		Format: "${ip} - ${method} ${path} - ${status} - ${latency}\n",
	}))

	s.initEndPoints(app)

	serverAddr := fmt.Sprintf(":%s", httpPort)
	slog.Info("🚀 Servidor iniciado", "url", "http://localhost"+serverAddr)

	if err := app.Listen(serverAddr); err != nil {
		log.Fatalf("Error al iniciar el servidor Fiber: %v", err)
	}
}

func (s *Server) Initialize() {
	portFromEnv := os.Getenv("HTTP_PORT_1")
	if portFromEnv != "" {
		httpPort = portFromEnv
	} else {
		slog.Info("Puerto HTTP no definido en .env, usando puerto por defecto", "port", httpPort)
	}
	s.startServer()
}
