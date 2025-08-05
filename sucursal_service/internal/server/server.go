package server

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"multiroom/sucursal-service/internal/server/setup"

	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

var (
	httpPort = "8080"
)

type Server struct {
	handlers setup.Handler
}

func NewServer(handlers setup.Handler) *Server {
	return &Server{handlers}
}

func (s *Server) startHTTPServer() {
	app := fiber.New(fiber.Config{
		BodyLimit:             20 << 23,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           30 * time.Second,
		DisableStartupMessage: true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Prefork:               false,
		AppName:               "Multiroom Backend-Servicio de sucursal",
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
		AllowMethods: "GET, POST, PUT, PATCH, DELETE",
	}))

	app.Use(logger.New(logger.Config{
		Format: "${ip} - ${method} ${path} - ${status} - ${latency}\n",
	}))

	s.initEndPointsHTTP(app)

	serverAddr := fmt.Sprintf(":%s", httpPort)
	slog.Info("🚀 Servidor HTTP iniciado", "url", "http://localhost"+serverAddr)

	if err := app.Listen(serverAddr); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}

func (s *Server) Initialize() {
	if port := os.Getenv("HTTP_PORT_3"); port != "" {
		httpPort = port
	}

	s.startHTTPServer()
}
