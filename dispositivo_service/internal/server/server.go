package server

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
	"log/slog"
	"multiroom/dispositivo-service/internal/server/setup"
	"os"
	"strings"
	"time"
)

var (
	httpPort      = "8080"
	websocketPort = "8081" // Puerto para WebSocket
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
		AppName:               "Multiroom Backend",
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

	go func() {
		if err := app.Listen(serverAddr); err != nil {
			log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
		}
	}()
}

func (s *Server) startWebSocketServer() {
	wsApp := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Multiroom WS",
	})

	s.initEndPointsWS(wsApp)

	wsAddr := fmt.Sprintf(":%s", websocketPort)
	slog.Info("🌐 Servidor WebSocket iniciado", "url", "ws://localhost"+wsAddr+"/ws")

	go func() {
		if err := wsApp.Listen(wsAddr); err != nil {
			log.Fatalf("Error al iniciar WebSocket server: %v", err)
		}
	}()
}

func (s *Server) Initialize() {
	if port := os.Getenv("HTTP_PORT_2"); port != "" {
		httpPort = port
	}
	if wsPort := os.Getenv("WS_PORT_1"); wsPort != "" {
		websocketPort = wsPort
	}

	s.startHTTPServer()
	s.startWebSocketServer()
	select {}
}
