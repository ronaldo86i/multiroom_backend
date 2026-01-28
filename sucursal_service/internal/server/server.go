package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"multiroom/sucursal-service/internal/server/setup"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	httpPort      = "8083"
	websocketPort = "8084"
)

// getCORSOrigins obtiene los orígenes permitidos desde variables de entorno
func getCORSOrigins() string {
	// Si necesitas credentials, define orígenes específicos
	if needsCredentials() {
		origins := os.Getenv("CORS_ORIGINS")
		if origins == "" {
			// Orígenes por defecto para desarrollo
			origins = "*"
		}
		return origins
	}
	// Si no necesitas credentials, puedes usar wildcard
	return "*"
}

// needsCredentials determina si la aplicación necesita enviar credentials
func needsCredentials() bool {
	// Cambiar según las necesidades de tu aplicación
	return os.Getenv("CORS_ALLOW_CREDENTIALS") == "true"
}

// createCORSConfig crea la configuración de CORS basada en el entorno
func createCORSConfig() cors.Config {
	origins := getCORSOrigins()
	allowCredentials := needsCredentials()

	// Si usamos wildcard, no podemos permitir credentials
	if origins == "*" {
		allowCredentials = false
	}

	return cors.Config{
		AllowOrigins: origins,
		AllowHeaders: strings.Join([]string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAuthorization,
			fiber.HeaderXDownloadOptions,
			fiber.HeaderReferrerPolicy,
			fiber.HeaderAccept,
			fiber.HeaderAcceptLanguage,
			fiber.HeaderConnection,
			fiber.HeaderUserAgent,
		}, ","),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowCredentials: allowCredentials,
		MaxAge:           300, // 5 minutos
	}
}

type Server struct {
	handlers setup.Handler
	httpApp  *fiber.App
	wsApp    *fiber.App
	wg       sync.WaitGroup
	shutdown chan os.Signal
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewServer(handlers setup.Handler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		handlers: handlers,
		shutdown: make(chan os.Signal, 1),
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Server) createHTTPApp() *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit:             20 << 23, // 160MB
		ReadTimeout:           60 * time.Second,
		WriteTimeout:          60 * time.Second,
		IdleTimeout:           120 * time.Second,
		DisableStartupMessage: true,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Prefork:               false,
		AppName:               "Multiroom Backend sucursal",
		// Configuración adicional para mejor manejo de errores
		ServerHeader:            "Multiroom-HTTP",
		StrictRouting:           false,
		CaseSensitive:           false,
		EnableTrustedProxyCheck: false,
		ProxyHeader:             fiber.HeaderXForwardedFor,
		// Configuración de red TCP
		Network: "tcp",
		// Deshabilitar keep-alive si causa problemas
		DisableKeepalive: false,
		// Configuración de compresión
		CompressedFileSuffix: ".gz",
		// Buffer pool para mejorar performance
		EnableIPValidation: false,
		// Error handler personalizado
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			// Log del error para debugging
			slog.Error("HTTP Error",
				"error", err.Error(),
				"method", c.Method(),
				"path", c.Path(),
				"ip", c.IP(),
				"code", code,
			)

			// Respuesta de error personalizada
			return c.Status(code).JSON(fiber.Map{
				"message": err.Error(),
			})
		},
	})

	// Middleware de recuperación de panic
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// CORS configuración segura
	app.Use(cors.New(createCORSConfig()))

	// Logger mejorado
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${ip} - ${method} ${path} - ${status} - ${latency} - ${error}\n",
		TimeFormat: "2006/01/02 15:04:05",
		TimeZone:   "Local",
		Output:     os.Stdout,
	}))

	// Middleware de health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	return app
}

func (s *Server) createWebSocketApp() *fiber.App {
	wsApp := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Multiroom Backend WS sucursal",
		ReadTimeout:           0,
		WriteTimeout:          0,
		IdleTimeout:           0,
		ServerHeader:          "Multiroom-WS",
		// Error handler para WebSocket
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			slog.Error("WebSocket Error",
				"error", err.Error(),
				"path", c.Path(),
				"ip", c.IP(),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware de recuperación para WebSocket
	wsApp.Use(recover.New())

	return wsApp
}

func (s *Server) startHTTPServer() {
	s.httpApp = s.createHTTPApp()
	s.initEndPointsHTTP(s.httpApp)

	serverAddr := fmt.Sprintf(":%s", httpPort)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		slog.Info("🚀 Servidor HTTP iniciado", "url", "http://localhost"+serverAddr)

		if err := s.httpApp.Listen(serverAddr); err != nil {
			slog.Error("Error en servidor HTTP", "error", err)
			s.cancel() // Cancelar contexto en caso de error
		}
	}()
}

func (s *Server) startWebSocketServer() {
	s.wsApp = s.createWebSocketApp()
	s.initEndPointsWS(s.wsApp)

	wsAddr := fmt.Sprintf(":%s", websocketPort)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		slog.Info("🌐 Servidor WebSocket iniciado", "url", "ws://localhost"+wsAddr+"/ws")

		if err := s.wsApp.Listen(wsAddr); err != nil {
			slog.Error("Error en servidor WebSocket", "error", err)
			s.cancel() // Cancelar contexto en caso de error
		}
	}()
}

func (s *Server) gracefulShutdown() {
	// Configurar señales del sistema
	signal.Notify(s.shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-s.shutdown
		slog.Info("🛑 Iniciando apagado graceful...")

		// Cancelar contexto
		s.cancel()

		// Crear contexto con timeout para shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Apagar servidores
		if s.httpApp != nil {
			slog.Info("Cerrando servidor HTTP...")
			if err := s.httpApp.ShutdownWithContext(shutdownCtx); err != nil {
				slog.Error("Error cerrando servidor HTTP", "error", err)
			}
		}

		if s.wsApp != nil {
			slog.Info("Cerrando servidor WebSocket...")
			if err := s.wsApp.ShutdownWithContext(shutdownCtx); err != nil {
				slog.Error("Error cerrando servidor WebSocket", "error", err)
			}
		}

		slog.Info("✅ Servidores cerrados correctamente")
	}()
}

func (s *Server) Initialize() {
	// Configurar puertos desde variables de entorno
	if port := os.Getenv("HTTP_PORT_3"); port != "" {
		httpPort = port
	}
	if wsPort := os.Getenv("WS_PORT_2"); wsPort != "" {
		websocketPort = wsPort
	}

	// Configurar graceful shutdown
	s.gracefulShutdown()

	// Iniciar servidores
	s.startHTTPServer()
	s.startWebSocketServer()

	slog.Info("✅ Todos los servidores iniciados correctamente")

	// Esperar hasta que se cancele el contexto o se reciba señal
	select {
	case <-s.ctx.Done():
		slog.Info("Contexto cancelado, cerrando servidores...")
	case <-s.shutdown:
		slog.Info("Señal de cierre recibida...")
	}

	// Esperar a que terminen todas las goroutines
	s.wg.Wait()
	slog.Info("🔚 Aplicación cerrada completamente")
}

// Métodos auxiliares para agregar middlewares específicos si son necesarios
func (s *Server) AddHTTPMiddleware(middleware ...fiber.Handler) {
	if s.httpApp != nil {
		for _, m := range middleware {
			s.httpApp.Use(m)
		}
	}
}

func (s *Server) AddWSMiddleware(middleware ...fiber.Handler) {
	if s.wsApp != nil {
		for _, m := range middleware {
			s.wsApp.Use(m)
		}
	}
}

// GetServerStats obtiene stats del servidor
func (s *Server) GetServerStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if s.httpApp != nil {
		// Aquí puedes agregar estadísticas del servidor HTTP si Fiber las proporciona
		stats["http_server"] = "running"
	}

	if s.wsApp != nil {
		stats["ws_server"] = "running"
	}

	stats["uptime"] = time.Now().Unix()
	return stats
}
