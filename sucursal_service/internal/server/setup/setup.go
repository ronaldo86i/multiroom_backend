package setup

import (
	"log"
	httpHandler "multiroom/sucursal-service/internal/adapter/handler/http"
	"multiroom/sucursal-service/internal/adapter/repository"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/service"

	"github.com/joho/godotenv"

	"os"

	wsHandler "multiroom/sucursal-service/internal/adapter/handler/websocket"
	"multiroom/sucursal-service/internal/postgresql"
	"sync"
)

type Repository struct {
	Pais              port.PaisRepository
	Sucursal          port.SucursalRepository
	Sala              port.SalaRepository
	AppVersion        port.AppVersionRepository
	Proveedor         port.ProveedorRepository
	Producto          port.ProductoRepository
	Ubicacion         port.UbicacionRepository
	Compra            port.CompraRepository
	Inventario        port.InventarioRepository
	Venta             port.VentaRepository
	MetodoPago        port.MetodoPagoRepository
	ProductoCategoria port.ProductoCategoriaRepository
}

type Service struct {
	Pais              port.PaisService
	Sucursal          port.SucursalService
	Sala              port.SalaService
	RabbitMQ          port.RabbitMQService
	AppVersion        port.AppVersionService
	Proveedor         port.ProveedorService
	Producto          port.ProductoService
	Ubicacion         port.UbicacionService
	Compra            port.CompraService
	Inventario        port.InventarioService
	Venta             port.VentaService
	MetodoPago        port.MetodoPagoService
	ProductoCategoria port.ProductoCategoriaService
	Reporte           port.ReporteService
}

type Handler struct {
	Pais              port.PaisHandler
	Sucursal          port.SucursalHandler
	Sala              port.SalaHandler
	SalaWS            port.SalaHandlerWS
	AppVersion        port.AppVersionHandler
	Proveedor         port.ProveedorHandler
	Producto          port.ProductoHandler
	Ubicacion         port.UbicacionHandler
	Compra            port.CompraHandler
	Inventario        port.InventarioHandler
	Venta             port.VentaHandler
	MetodoPago        port.MetodoPagoHandler
	ProductoCategoria port.ProductoCategoriaHandler
	Reporte           port.ReporteHandler
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
		repositories.Proveedor = repository.NewProveedorRepository(pool)
		repositories.Producto = repository.NewProductoRepository(pool)
		repositories.Ubicacion = repository.NewUbicacionRepository(pool)
		repositories.Compra = repository.NewCompraRepository(pool)
		repositories.Inventario = repository.NewInventarioRepository(pool)
		repositories.Venta = repository.NewVentaRepository(pool)
		repositories.MetodoPago = repository.NewMetodoPagoRepository(pool)
		repositories.ProductoCategoria = repository.NewProductoCategoriaRepository(pool)
		// Services
		services.RabbitMQ = service.NewRabbitMQService(os.Getenv("RABBITMQ_URL"))
		services.Pais = service.NewPaisService(repositories.Pais)
		services.Sucursal = service.NewSucursalService(repositories.Sucursal)
		services.Sala = service.NewSalaService(repositories.Sala)
		services.AppVersion = service.NewAppVersionService(repositories.AppVersion)
		services.Proveedor = service.NewProveedorService(repositories.Proveedor)
		services.Producto = service.NewProductoService(repositories.Producto)
		services.Ubicacion = service.NewUbicacionService(repositories.Ubicacion)
		services.Compra = service.NewCompraService(repositories.Compra)
		services.Inventario = service.NewInventarioService(repositories.Inventario)
		services.Venta = service.NewVentaService(repositories.Venta)
		services.MetodoPago = service.NewMetodoPagoService(repositories.MetodoPago)
		services.ProductoCategoria = service.NewProductoCategoriaService(repositories.ProductoCategoria)
		services.Reporte = service.NewReporteService(repositories.Venta, repositories.Sucursal, repositories.Producto)
		// Handlers
		handlers.Pais = httpHandler.NewPaisHandler(services.Pais)
		handlers.Sucursal = httpHandler.NewSucursalHandler(services.Sucursal)
		handlers.SalaWS = wsHandler.NewSalaHandlerWS(services.Sala, services.RabbitMQ)
		handlers.Sala = httpHandler.NewSalaHandler(services.Sala, services.RabbitMQ)
		handlers.AppVersion = httpHandler.NewAppVersionHandler(services.AppVersion)
		handlers.Proveedor = httpHandler.NewProveedorHandler(services.Proveedor)
		handlers.Producto = httpHandler.NewProductoHandler(services.Producto)
		handlers.Ubicacion = httpHandler.NewUbicacionHandler(services.Ubicacion)
		handlers.Compra = httpHandler.NewCompraHandler(services.Compra)
		handlers.Inventario = httpHandler.NewInventarioHandler(services.Inventario)
		handlers.Venta = httpHandler.NewVentaHandler(services.Venta)
		handlers.MetodoPago = httpHandler.NewMetodoPagoHandler(services.MetodoPago)
		handlers.ProductoCategoria = httpHandler.NewProductoCategoriaHandler(services.ProductoCategoria)
		handlers.Reporte = httpHandler.NewReporteHandler(services.Reporte)
		instance = d
	})
}
