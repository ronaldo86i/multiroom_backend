package postgresql

import (
	"context"
	"fmt"
	"log/slog"
	"multiroom/sucursal-service/internal/postgresql/slog_logger"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

type DB struct {
	Pool *pgxpool.Pool
}

var (
	dbInstance DB
	once       sync.Once
	initErr    error
	logger     *slog.Logger
)

// GetDB retorna la instancia única del pool de conexión
func GetDB() *pgxpool.Pool {
	return dbInstance.Pool
}

// Connection inicializa la conexión si no se ha hecho antes
func Connection() error {
	once.Do(func() {
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")
		timezone := os.Getenv("DB_TIMEZONE")
		sslMode := "disable"

		if host == "" || user == "" || password == "" || dbname == "" || port == "" || timezone == "" {
			initErr = fmt.Errorf("una o más variables de entorno están vacías para inicializar la base de datos")
			return
		}

		connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&timezone=%s",
			user, password, host, port, dbname, sslMode, timezone)

		config, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			initErr = fmt.Errorf("error al parsear la cadena de conexión: %w", err)
			return
		}

		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelError,
			}),
		)

		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   slog_logger.NewLogger(logger),
			LogLevel: tracelog.LogLevelError,
		}

		config.MaxConns = 30
		config.MinConns = 5
		config.MaxConnLifetime = time.Hour
		config.HealthCheckPeriod = time.Minute

		ctx := context.Background()
		pool, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			initErr = fmt.Errorf("error al conectar con la base de datos: %w", err)
			return
		}

		if err := pool.Ping(ctx); err != nil {
			initErr = fmt.Errorf("no se pudo hacer ping a la base de datos: %w", err)
			return
		}

		dbInstance.Pool = pool

		logger.Info("Inicialización exitosa de la base de datos",
			slog.String("host", host),
			slog.String("usuario", user),
			slog.String("base de datos", dbname),
		)
		err = dbInstance.Migration()
		if err != nil {
			initErr = fmt.Errorf("no se pudo realizar la migración: %w", err)
			return
		}
	})

	return initErr
}
