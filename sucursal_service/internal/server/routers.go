package server

import (
	"multiroom/sucursal-service/internal/core/util"
	"multiroom/sucursal-service/internal/server/middleware"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

//func rateLimiter(max int, expiration, delay time.Duration) fiber.Handler {
//	return limiter.New(limiter.Config{
//		Max:        max,
//		Expiration: expiration,
//		LimitReached: func(c *fiber.Ctx) error {
//			time.Sleep(delay)
//			return c.Next()
//		},
//	})
//}

func (s *Server) initEndPointsHTTP(app *fiber.App) {
	app.Static("/", "./public", fiber.Static{
		ModifyResponse: util.EnviarArchivo,
	})
	api := app.Group("/api")
	s.endPointsAPI(api)
}

func (s *Server) endPointsAPI(api fiber.Router) {
	v1 := api.Group("/v1")
	s.endPointsApiUsuarioSucursal(v1)

	// ==========================================
	// PAISES (Recurso: pais)
	// ==========================================
	v1Paises := v1.Group("/paises")
	v1Paises.Use(middleware.HostnameMiddleware)
	v1Paises.Get("", middleware.VerifyPermission("pais:ver"), s.handlers.Pais.ObtenerListaPaises)
	v1Paises.Get("/:paisId", middleware.VerifyPermission("pais:ver"), s.handlers.Pais.ObtenerPaisById)
	v1Paises.Post("", middleware.VerifyPermission("pais:crear"), s.handlers.Pais.RegistrarPais)
	v1Paises.Put("/:paisId", middleware.VerifyPermission("pais:editar"), s.handlers.Pais.ModificarPais)
	v1Paises.Patch("/:paisId/deshabilitar", middleware.VerifyPermission("pais:editar"), s.handlers.Pais.DeshabilitarPaisById)
	v1Paises.Patch("/:paisId/habilitar", middleware.VerifyPermission("pais:editar"), s.handlers.Pais.HabilitarPaisById)

	// ==========================================
	// SUCURSALES (Recurso: sucursal)
	// ==========================================
	v1Sucursales := v1.Group("/sucursales")
	v1Sucursales.Use(middleware.HostnameMiddleware)
	v1Sucursales.Get("", middleware.VerifyPermission("sucursal:ver"), s.handlers.Sucursal.ObtenerListaSucursales)
	v1Sucursales.Get("/:sucursalId", middleware.VerifyPermission("sucursal:ver"), s.handlers.Sucursal.ObtenerSucursalById)
	v1Sucursales.Post("", middleware.VerifyPermission("sucursal:crear"), s.handlers.Sucursal.RegistrarSucursal)
	v1Sucursales.Put("/:sucursalId", middleware.VerifyPermission("sucursal:editar"), s.handlers.Sucursal.ModificarSucursal)
	v1Sucursales.Patch("/:sucursalId/habilitar", middleware.VerifyPermission("sucursal:editar"), s.handlers.Sucursal.HabilitarSucursal)
	v1Sucursales.Patch("/:sucursalId/deshabilitar", middleware.VerifyPermission("sucursal:editar"), s.handlers.Sucursal.DeshabilitarSucursal)

	// ==========================================
	// SALAS (Recurso: sala)
	// ==========================================
	v1Salas := v1.Group("/salas")
	v1Salas.Use(middleware.HostnameMiddleware)
	v1Salas.Get("", middleware.VerifyPermission("sala:ver"), s.handlers.Sala.ObtenerListaSalas)
	v1Salas.Get("/uso", middleware.VerifyPermission("sala:ver"), s.handlers.Sala.ObtenerListaUsoSalas)
	v1Salas.Get("/:salaId", middleware.VerifyPermission("sala:ver"), s.handlers.Sala.ObtenerSalaById)
	v1Salas.Post("", middleware.VerifyPermission("sala:crear"), s.handlers.Sala.RegistrarSala)
	v1Salas.Put("/:salaId", middleware.VerifyPermission("sala:editar"), s.handlers.Sala.ModificarSala)
	v1Salas.Patch("/:salaId/habilitar", middleware.VerifyPermission("sala:editar"), s.handlers.Sala.HabilitarSala)
	v1Salas.Patch("/:salaId/deshabilitar", middleware.VerifyPermission("sala:editar"), s.handlers.Sala.DeshabilitarSala)
	v1Salas.Delete("/:salaId/eliminar", middleware.VerifyPermission("sala:eliminar"), s.handlers.Sala.EliminarSalaById)

	// Acciones de Salas (Controlar tiempos) - Permiso específico
	v1AccionesSalas := v1.Group("/acciones/salas")
	v1AccionesSalas.Use(middleware.HostnameMiddleware)
	// 'sala:controlar' es para pausar, reanudar, asignar tiempo
	v1AccionesSalas.Post("", middleware.VerifyPermission("sala:controlar"), s.handlers.Sala.AsignarTiempoUsoSala)
	v1AccionesSalas.Patch("/cancelar/:salaId", middleware.VerifyPermission("sala:controlar"), s.handlers.Sala.CancelarSala)
	v1AccionesSalas.Patch("/pausar/:salaId", middleware.VerifyPermission("sala:controlar"), s.handlers.Sala.PausarTiempoUsoSala)
	v1AccionesSalas.Patch("/reanudar/:salaId", middleware.VerifyPermission("sala:controlar"), s.handlers.Sala.ReanudarTiempoUsoSala)
	v1AccionesSalas.Patch("/incrementar/:salaId", middleware.VerifyPermission("sala:controlar"), s.handlers.Sala.IncrementarTiempoUsoSala)

	// ==========================================
	// PROVEEDORES (Recurso: proveedor)
	// ==========================================
	v1Proveedores := v1.Group("/proveedores")
	v1Proveedores.Use(middleware.HostnameMiddleware)
	v1Proveedores.Get("", middleware.VerifyPermission("proveedor:ver"), s.handlers.Proveedor.ListarProveedores)
	v1Proveedores.Get("/:proveedorId", middleware.VerifyPermission("proveedor:ver"), s.handlers.Proveedor.ObtenerProveedorById)
	v1Proveedores.Post("", middleware.VerifyPermission("proveedor:crear"), s.handlers.Proveedor.RegistrarProveedor)
	v1Proveedores.Put("/:proveedorId", middleware.VerifyPermission("proveedor:editar"), s.handlers.Proveedor.ModificarProveedor)

	// ==========================================
	// PRODUCTOS (Recurso: producto)
	// ==========================================
	v1Productos := v1.Group("/productos")
	v1Productos.Use(middleware.HostnameMiddleware)
	v1Productos.Get("", middleware.VerifyPermission("producto:ver"), s.handlers.Producto.ListarProductos)
	v1Productos.Get("/stats/topProductos", middleware.VerifyPermission("producto:ver"), s.handlers.Producto.ListarProductosMasVendidos)
	v1Productos.Get("/sucursales", middleware.VerifyPermission("producto:ver"), s.handlers.Producto.ListarProductosPorSucursal)
	v1Productos.Get("/sucursales/:productoSucursalId", middleware.VerifyPermission("producto:ver"), s.handlers.Producto.ObtenerProductoSucursalById)
	v1Productos.Get("/:productoId", middleware.VerifyPermission("producto:ver"), s.handlers.Producto.ObtenerProductoById)

	v1Productos.Put("/sucursales/:productoSucursalId", middleware.VerifyPermission("producto:editar"), s.handlers.Producto.ActualizarProductoSucursal)
	v1Productos.Post("", middleware.VerifyPermission("producto:crear"), s.handlers.Producto.RegistrarProducto)
	v1Productos.Put("/:productoId", middleware.VerifyPermission("producto:editar"), s.handlers.Producto.ModificarProductoById)
	v1Productos.Patch("/:productoId/habilitar", middleware.VerifyPermission("producto:editar"), s.handlers.Producto.HabilitarProductoById)
	v1Productos.Patch("/:productoId/deshabilitar", middleware.VerifyPermission("producto:editar"), s.handlers.Producto.DeshabilitarProductoById)
	// Categorias
	v1ProductosCategorias := v1.Group("/productos-categorias")
	v1ProductosCategorias.Use(middleware.HostnameMiddleware)
	v1ProductosCategorias.Get("", middleware.VerifyPermission("categoria:ver"), s.handlers.ProductoCategoria.ListarCategorias)
	v1ProductosCategorias.Get("/:categoriaId", middleware.VerifyPermission("categoria:ver"), s.handlers.ProductoCategoria.ObtenerCategoriaById)
	v1ProductosCategorias.Post("", middleware.VerifyPermission("categoria:crear"), s.handlers.ProductoCategoria.RegistrarCategoria)
	v1ProductosCategorias.Put("/:categoriaId", middleware.VerifyPermission("categoria:editar"), s.handlers.ProductoCategoria.ModificarCategoriaById)

	// ==========================================
	// UBICACIONES (Recurso: ubicacion)
	// ==========================================
	v1Ubicaciones := v1.Group("/ubicaciones")
	v1Ubicaciones.Use(middleware.HostnameMiddleware)
	v1Ubicaciones.Get("", middleware.VerifyPermission("ubicacion:ver"), s.handlers.Ubicacion.ListarUbicaciones)
	v1Ubicaciones.Get("/:ubicacionId", middleware.VerifyPermission("ubicacion:ver"), s.handlers.Ubicacion.ObtenerUbicacionById)
	v1Ubicaciones.Post("", middleware.VerifyPermission("ubicacion:crear"), s.handlers.Ubicacion.RegistrarUbicacion)
	v1Ubicaciones.Put("/:ubicacionId", middleware.VerifyPermission("ubicacion:editar"), s.handlers.Ubicacion.ModificarUbicacionById)
	v1Ubicaciones.Patch("/:ubicacionId/habilitar", middleware.VerifyPermission("ubicacion:editar"), s.handlers.Ubicacion.HabilitarUbicacion)
	v1Ubicaciones.Patch("/:ubicacionId/deshabilitar", middleware.VerifyPermission("ubicacion:editar"), s.handlers.Ubicacion.DeshabilitarUbicacion)

	// ==========================================
	// COMPRAS (Recurso: compra)
	// ==========================================
	v1Compras := v1.Group("/compras")
	v1Compras.Use(middleware.HostnameMiddleware)
	v1Compras.Get("", middleware.VerifyPermission("compra:ver"), s.handlers.Compra.ListarCompras)
	v1Compras.Get("/:compraId", middleware.VerifyPermission("compra:ver"), s.handlers.Compra.ObtenerCompraById)
	v1Compras.Post("", middleware.VerifyPermission("compra:crear"), s.handlers.Compra.RegistrarOrdenCompra)
	v1Compras.Put("/:compraId", middleware.VerifyPermission("compra:editar"), s.handlers.Compra.ModificarOrdenCompra)
	// Recepcionar/Completar compra puede requerir un permiso especial
	v1Compras.Post("/:compraId/completar", middleware.VerifyPermission("compra:procesar"), s.handlers.Compra.ConfirmarRecepcionCompra)

	// ==========================================
	// INVENTARIO (Recurso: inventario / ajuste / transferencia)
	// ==========================================
	v1Inventario := v1.Group("/inventario")
	v1Inventario.Use(middleware.HostnameMiddleware)
	v1Inventario.Get("", middleware.VerifyPermission("inventario:ver"), s.handlers.Inventario.ListarInventario)
	v1Inventario.Get("/ajustes", middleware.VerifyPermission("ajuste_inventario:ver"), s.handlers.Inventario.ListarAjustes)
	v1Inventario.Get("/ajustes/:ajusteId", middleware.VerifyPermission("ajuste_inventario:ver"), s.handlers.Inventario.ObtenerAjusteById)
	v1Inventario.Post("/ajustes", middleware.VerifyPermission("ajuste_inventario:crear"), s.handlers.Inventario.RegistrarAjusteConDetalle)
	v1Inventario.Get("/transferencias", middleware.VerifyPermission("transferencia:ver"), s.handlers.Inventario.ListarTransferencias)
	v1Inventario.Get("/transferencias/:transferenciaId", middleware.VerifyPermission("transferencia:ver"), s.handlers.Inventario.ObtenerTransferenciaById)
	v1Inventario.Post("/transferencias", middleware.VerifyPermission("transferencia:crear"), s.handlers.Inventario.RegistrarTransferencia)

	// ==========================================
	// VENTAS (Recurso: venta)
	// ==========================================
	v1Ventas := v1.Group("/ventas")
	v1Ventas.Use(middleware.HostnameMiddleware)
	v1Ventas.Get("", middleware.VerifyPermission("venta:ver"), s.handlers.Venta.ListarVentas)
	v1Ventas.Get("/productos", middleware.VerifyPermission("venta:ver"), s.handlers.Venta.ListarProductosVentas)
	v1Ventas.Get("/:ventaId/comprobante", middleware.VerifyPermission("venta:ver"), s.handlers.Reporte.ComprobantePDFVentaById)
	v1Ventas.Get("/:ventaId", middleware.VerifyPermission("venta:ver"), s.handlers.Venta.ObtenerVenta)
	v1Ventas.Post("", middleware.VerifyPermission("venta:crear"), s.handlers.Venta.RegistrarVenta)
	v1Ventas.Post("/:ventaId/pagar", middleware.VerifyPermission("venta:cobrar"), s.handlers.Venta.RegistrarPagoVenta)
	v1Ventas.Post("/:ventaId/anular", middleware.VerifyPermission("venta:anular"), s.handlers.Venta.AnularVentaById)

	// Metodos de Pagos
	v1MetodosPagos := v1.Group("/metodos-pago")
	v1MetodosPagos.Get("", middleware.VerifyPermission("metodo_pago:ver"), s.handlers.MetodoPago.ListarMetodosPago)

	// ==========================================
	// APP VERSION (Recurso: app_version)
	// ==========================================
	v1AppVersion := v1.Group("/app-version")
	v1AppVersion.Use(middleware.HostnameMiddleware)
	v1AppVersion.Get("/ultimo", s.handlers.AppVersion.ObtenerUltimaVersion) // Puede ser pública o requerir auth básica
	v1AppVersion.Get("", middleware.VerifyPermission("app_version:ver"), s.handlers.AppVersion.ObtenerListaVersiones)
	v1AppVersion.Post("", middleware.VerifyPermission("app_version:crear"), s.handlers.AppVersion.RegistrarApp)
	v1AppVersion.Get("/:appVersionId", middleware.VerifyPermission("app_version:ver"), s.handlers.AppVersion.ObtenerVersion)
	v1AppVersion.Put("/:appVersionId", middleware.VerifyPermission("app_version:editar"), s.handlers.AppVersion.ModificarVersion)
}

func (s *Server) endPointsApiUsuarioSucursal(api fiber.Router) {
	// Mantenemos la lógica separada para usuarios de sucursal (POS, Terminales)
	v1UsuarioSucursal := api.Group("/usuario-sucursal")
	v1UsuarioSucursal.Use(middleware.HostnameMiddleware)

	v1Salas := v1UsuarioSucursal.Group("/salas")
	v1Salas.Get("", middleware.VerifyUsuarioSucursal, s.handlers.Sala.ObtenerListaSalas)
	v1Salas.Get("/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.ObtenerSalaById)

	v1AccionesSalas := v1UsuarioSucursal.Group("/acciones/salas")
	v1AccionesSalas.Use(middleware.HostnameMiddleware)
	v1AccionesSalas.Post("", middleware.VerifyUsuarioSucursal, s.handlers.Sala.AsignarTiempoUsoSala)
	v1AccionesSalas.Patch("/cancelar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.CancelarSala)
	v1AccionesSalas.Patch("/pausar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.PausarTiempoUsoSala)
	v1AccionesSalas.Patch("/reanudar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.ReanudarTiempoUsoSala)
	v1AccionesSalas.Patch("/incrementar/:salaId", middleware.VerifyUsuarioSucursal, s.handlers.Sala.IncrementarTiempoUsoSala)
}

func (s *Server) initEndPointsWS(app *fiber.App) {
	ws := app.Group("/ws")
	v1 := ws.Group("/v1")
	s.endPointsWS(v1)
	s.endPointsWSUsuarioSucursal(v1)
}

func (s *Server) endPointsWS(api fiber.Router) {
	// Websockets para monitoreo (Admin)
	v1Salas := api.Group("/salas")
	// Reemplazado ADMIN por permiso de ver sala
	v1Salas.Get("", middleware.VerifyPermission("sala:ver"), websocket.New(s.handlers.SalaWS.UsoSalas))
	v1Salas.Get("/:salaId", middleware.VerifyPermission("sala:ver"), websocket.New(s.handlers.SalaWS.UsoSala))
}

func (s *Server) endPointsWSUsuarioSucursal(api fiber.Router) {
	v1UsuarioSucursal := api.Group("/usuario-sucursal")
	v1Salas := v1UsuarioSucursal.Group("/salas")
	v1Salas.Get("", middleware.VerifyUsuarioSucursal, websocket.New(s.handlers.SalaWS.UsoSalasBySucursalId))
}
