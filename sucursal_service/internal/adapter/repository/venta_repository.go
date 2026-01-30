package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VentaRepository struct {
	pool *pgxpool.Pool
}

func (v VentaRepository) ListarProductosVentas(ctx context.Context, filtros map[string]string) (*[]domain.ProductoVentaStat, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")

	var filters []string
	// $1 es el hostname para la foto
	var args = []interface{}{fullHostname}
	var j = 2

	// 1. Filtro Sucursal
	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s1.id = $%d", j))
		args = append(args, sucursalId)
		j++
	}

	// 2. Filtro Fecha Inicio (v.creado_en >= fecha)
	if fechaInicio := filtros["fechaInicio"]; fechaInicio != "" {
		// Se asume formato 'YYYY-MM-DD' o 'YYYY-MM-DD HH:MM:SS'
		filters = append(filters, fmt.Sprintf("v.creado_en >= $%d", j))
		args = append(args, fechaInicio)
		j++
	}

	// 3. Filtro Fecha Fin (v.creado_en <= fecha)
	if fechaFin := filtros["fechaFin"]; fechaFin != "" {
		// Se asume formato 'YYYY-MM-DD' o 'YYYY-MM-DD HH:MM:SS'
		// Nota: Si solo envías fecha, considera que '2023-01-01' es las 00:00.
		// Para incluir el día suele usarse la fecha siguiente o agregar hora 23:59:59 desde el frontend.
		filters = append(filters, fmt.Sprintf("v.creado_en <= $%d", j))
		args = append(args, fechaFin)
		j++
	}

	var list = make([]domain.ProductoVentaStat, 0)

	query := `
       SELECT
          json_build_object(
             'id', p.id,
             'nombre', p.nombre,
             'estado', p.estado,
             'urlFoto', ($1::text || p.id::text || '/' || p.foto),
             'esInventariable', p.es_inventariable,
             'creadoEn', p.creado_en,
             'actualizadoEn', p.actualizado_en,
             'eliminadoEn', p.eliminado_en
          ) as producto_info,
          -- (Cantidad * Precio) - DescuentoLinea
          COALESCE(SUM(dv.cantidad * dv.precio_venta - dv.descuento), 0) as total_ventas,
          COALESCE(SUM(dv.cantidad), 0) as cantidad_ventas
       FROM producto p
       LEFT JOIN detalle_venta dv ON p.id = dv.producto_id
       LEFT JOIN venta v ON dv.venta_id = v.id AND v.estado = 'Completado'
       LEFT JOIN sucursal s1 ON v.sucursal_id = s1.id 
    `

	// 4. Aplicar Filtros (WHERE)
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	// 5. Agrupación y Ordenamiento
	query += " GROUP BY p.id ORDER BY total_ventas DESC"

	// 6. Ejecutar Query
	rows, err := v.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al listar estadísticas de productos:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	// 7. Escanear Resultados
	for rows.Next() {
		var item domain.ProductoVentaStat
		err := rows.Scan(&item.Producto, &item.TotalVentas, &item.CantidadVentas)
		if err != nil {
			log.Println("Error al escanear:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		list = append(list, item)
	}

	if rows.Err() != nil {
		log.Println("Error en iteración de rows:", rows.Err())
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	return &list, nil
}

func (v VentaRepository) ListarVentas(ctx context.Context, filtros map[string]string) (*[]domain.VentaInfo, error) {
	var filters []string
	var args []interface{}
	var j = 1

	if usuarioIdStr := filtros["usuarioId"]; usuarioIdStr != "" {
		usuarioId, err := strconv.Atoi(usuarioIdStr)
		if err != nil {
			log.Println("Error al convertir usuarioId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de usuarioId no es válido")
		}
		filters = append(filters, fmt.Sprintf("v.usuario_id = $%d", j))
		args = append(args, usuarioId)
		j++
	}

	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s1.id = $%d", j))
		args = append(args, sucursalId)
		j++
	}

	// 2. Filtro Fecha Inicio (v.creado_en >= fecha)
	if fechaInicio := filtros["fechaInicio"]; fechaInicio != "" {
		// Se asume formato 'YYYY-MM-DD' o 'YYYY-MM-DD HH:MM:SS'
		filters = append(filters, fmt.Sprintf("v.creado_en >= $%d", j))
		args = append(args, fechaInicio)
		j++
	}

	// 3. Filtro Fecha Fin (v.creado_en <= fecha)
	if fechaFin := filtros["fechaFin"]; fechaFin != "" {
		// Se asume formato 'YYYY-MM-DD' o 'YYYY-MM-DD HH:MM:SS'
		filters = append(filters, fmt.Sprintf("v.creado_en <= $%d", j))
		args = append(args, fechaFin)
		j++
	}

	if salaIdStr := filtros["salaId"]; salaIdStr != "" {
		salaId, err := strconv.Atoi(salaIdStr)
		if err != nil {
			log.Println("Error al convertir salaId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de salaId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s2.id = $%d", j))
		args = append(args, salaId)
		j++
	}

	if usoSalaIdStr := filtros["usoSalaId"]; usoSalaIdStr != "" {
		usoSalaId, err := strconv.Atoi(usoSalaIdStr)
		if err != nil {
			log.Println("Error al convertir usoSalaId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de usoSalaId no es válido")
		}
		filters = append(filters, fmt.Sprintf("us.id = $%d", j))
		args = append(args, usoSalaId)
		j++
	}

	if estado := filtros["estado"]; estado != "" {
		filters = append(filters, fmt.Sprintf("v.estado = $%d", j))
		args = append(args, estado)
		j++
	}

	// Requiere hacer JOIN con detalle_venta (agregado en la query abajo)
	if productoIdStr := filtros["productoId"]; productoIdStr != "" {
		productoId, err := strconv.Atoi(productoIdStr)
		if err != nil {
			log.Println("Error al convertir productoId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de productoId no es válido")
		}
		// Usamos el alias 'dv' que definiremos en el JOIN
		filters = append(filters, fmt.Sprintf("dv.producto_id = $%d", j))
		args = append(args, productoId)
		j++
	}

	if fechaInicio := filtros["fechaInicio"]; fechaInicio != "" {
		filters = append(filters, fmt.Sprintf("v.creado_en >= $%d", j))
		args = append(args, fechaInicio)
		j++
	}

	if fechaFin := filtros["fechaFin"]; fechaFin != "" {
		filters = append(filters, fmt.Sprintf("v.creado_en <= $%d", j))
		args = append(args, fechaFin)
		j++
	}

	query := `
    SELECT 
       v.id,
       v.codigo_venta,
       v.total,
       v.estado,
       v.creado_en,
       v.actualizado_en,
       v.costo_tiempo_venta,
       v.descuento_general,
       v.observacion,
       json_build_object(
          'id',ua.id,
          'username',ua.username
       ) AS usuario,
       json_build_object(
          'id', c.id,
          'nombres', c.nombres,
          'apellidos', c.apellidos,
          'codigoPais', c.codigo_pais,
          'celular', c.celular,
          'fechaNacimiento', c.fecha_nacimiento,
          'estado', c.estado,
          'creadoEn', c.creado_en
       ) AS cliente,
       jsonb_build_object(
          'id', s1.id,
          'nombre', s1.nombre,
          'estado', s1.estado,
          'creadoEn', s1.creado_en
       ) AS sucursal,
       jsonb_build_object(
          'id',s2.id,
          'nombre',s2.nombre,
          'estado',s2.estado,
          'creadoEn',s2.creado_en,
          'actualizadoEn',s2.actualizado_en,
          'eliminadoEn',s2.eliminado_en,
          'dispositivo',jsonb_build_object(
             'id', d.id,
             'dispositivoId', d.dispositivo_id,
             'nombre', d.nombre,
             'estado', d.estado,
             'creadoEn', d.creado_en,
             'usuario', COALESCE(
             jsonb_build_object(
                'id', u.id,
                'username', u.username
             ), '{}'::jsonb)
          )
       ) AS sala
    FROM  venta v
    LEFT JOIN public.usuario_admin ua on v.usuario_id = ua.id
    LEFT JOIN public.cliente c on v.cliente_id = c.id
    LEFT JOIN public.sucursal s1 on v.sucursal_id = s1.id
    LEFT JOIN public.sala s2 on v.sala_id = s2.id
    LEFT JOIN public.dispositivo d on s2.dispositivo_id = d.id
    LEFT JOIN public.usuario u on d.usuario_id = u.id
    LEFT JOIN public.uso_sala us on v.uso_sala_id = us.id
    LEFT JOIN public.detalle_venta dv on v.id = dv.venta_id
    `

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	// El GROUP BY asegura que si una venta tiene el mismo producto varias veces
	// o si el JOIN duplica filas, solo obtengamos una fila por venta.
	query += `
    GROUP BY v.id, c.id, d.id, ua.id, s1.id, s2.id, u.id
    ORDER BY v.id DESC
    `
	// Nota: Cambié ORDER BY a DESC usualmente se quieren ver las ventas recientes primero,
	// puedes dejarlo sin DESC si prefieres orden ascendente.

	rows, err := v.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error ejecutando query ListarVentas:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	list := make([]domain.VentaInfo, 0)
	for rows.Next() {
		var item domain.VentaInfo
		err := rows.Scan(&item.Id, &item.CodigoVenta, &item.Total, &item.Estado, &item.CreadoEn, &item.ActualizadoEn, &item.CostoTiempoVenta, &item.DescuentoGeneral, &item.Observacion, &item.Usuario, &item.Cliente, &item.Sucursal, &item.Sala)
		if err != nil {
			log.Println("Error al obtener lista de venta:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}

	return &list, nil
}

func (v VentaRepository) RegistrarVenta(ctx context.Context, request *domain.VentaRequest) (*int, error) {
	type stockDisponible struct {
		UbicacionId int
		Stock       int
	}

	// 1. Iniciar transacción
	tx, err := v.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			if rollErr := tx.Rollback(ctx); rollErr != nil {
				log.Println("Error durante rollback:", rollErr)
			}
		}
	}()

	// 2. Validaciones iniciales
	if request.SucursalId <= 0 {
		return nil, datatype.NewBadRequestError("El ID de la sucursal es obligatorio.")
	}

	var sucursalExiste bool
	queryValidaSucursal := `SELECT EXISTS(SELECT 1 FROM sucursal WHERE id = $1 AND estado = 'Activo')`
	err = tx.QueryRow(ctx, queryValidaSucursal, request.SucursalId).Scan(&sucursalExiste)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if !sucursalExiste {
		return nil, datatype.NewBadRequestError(fmt.Sprintf("La sucursal con id %d no existe o está inactiva.", request.SucursalId))
	}

	// 3. Actualizar uso de sala (si aplica)
	if request.UsoSalaId != nil {
		queryUsoSala := `UPDATE uso_sala SET costo_tiempo = costo_tiempo + $1, actualizado_en = NOW() 
                         WHERE id = $2 AND estado IN ('En uso', 'Pausado') AND tipo = 'General'`
		ct, err := tx.Exec(ctx, queryUsoSala, request.CostoTiempo, *request.UsoSalaId)
		if err != nil || ct.RowsAffected() == 0 {
			return nil, datatype.NewBadRequestError("La sesión de uso de sala no está activa.")
		}
	}

	var totalVenta float64 = 0
	var detallesParaGuardar [][]interface{}

	// Consultas preparadas
	queryGetProductoInfo := `
        SELECT ps.precio, p.es_inventariable, p.nombre 
        FROM producto_sucursal ps
        JOIN producto p ON ps.producto_id = p.id
        WHERE ps.producto_id = $1 AND ps.sucursal_id = $2`

	queryFindStock := `
        SELECT i.stock, i.ubicacion_id
        FROM inventario i
        JOIN ubicacion u ON i.ubicacion_id = u.id
        WHERE i.producto_id = $1 AND u.sucursal_id = $2 AND u.es_vendible = true AND u.estado = 'Activo'
        ORDER BY u.prioridad_venta FOR UPDATE OF i`

	// 4. Bucle de Detalles de Venta
	for _, detalleReq := range request.Detalles {
		if detalleReq.Cantidad <= 0 {
			return nil, datatype.NewBadRequestError("Cantidad debe ser mayor a cero")
		}

		var precioVenta float64
		var esInventariable bool
		var nombreProducto string

		err = tx.QueryRow(ctx, queryGetProductoInfo, detalleReq.ProductoId, request.SucursalId).Scan(&precioVenta, &esInventariable, &nombreProducto)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, datatype.NewBadRequestError(fmt.Sprintf("Producto %d no disponible en esta sucursal.", detalleReq.ProductoId))
			}
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		subtotalBrutoLinea := precioVenta * float64(detalleReq.Cantidad)
		if detalleReq.Descuento > subtotalBrutoLinea {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("Descuento excesivo en %s", nombreProducto))
		}

		descuentoUnitario := detalleReq.Descuento / float64(detalleReq.Cantidad)
		totalVenta += subtotalBrutoLinea - detalleReq.Descuento

		// --- Lógica de Inventario ---
		if esInventariable {
			rows, err := tx.Query(ctx, queryFindStock, detalleReq.ProductoId, request.SucursalId)
			if err != nil {
				return nil, datatype.NewInternalServerErrorGeneric()
			}

			var stockDisponibleList []stockDisponible
			var stockTotalVendible = 0
			for rows.Next() {
				var s stockDisponible
				if err := rows.Scan(&s.Stock, &s.UbicacionId); err != nil {
					log.Println("Error al escanear:", err)
					return nil, datatype.NewInternalServerErrorGeneric()
				}
				if s.Stock > 0 {
					stockDisponibleList = append(stockDisponibleList, s)
					stockTotalVendible += s.Stock
				}
			}
			rows.Close()

			if stockTotalVendible < int(detalleReq.Cantidad) {
				return nil, datatype.NewBadRequestError(fmt.Sprintf("Stock insuficiente para %s. Disponible: %d", nombreProducto, stockTotalVendible))
			}

			cantidadARestar := int(detalleReq.Cantidad)
			for _, s := range stockDisponibleList {
				if cantidadARestar == 0 {
					break
				}

				cantidadDescontada := 0
				if cantidadARestar <= s.Stock {
					cantidadDescontada = cantidadARestar
					cantidadARestar = 0
				} else {
					cantidadDescontada = s.Stock
					cantidadARestar -= s.Stock
				}

				_, err = tx.Exec(ctx, `UPDATE inventario SET stock = stock - $1 WHERE producto_id = $2 AND ubicacion_id = $3`,
					cantidadDescontada, detalleReq.ProductoId, s.UbicacionId)
				if err != nil {
					return nil, datatype.NewInternalServerErrorGeneric()
				}

				detallesParaGuardar = append(detallesParaGuardar, []interface{}{
					nil, detalleReq.ProductoId, s.UbicacionId, cantidadDescontada, precioVenta, descuentoUnitario * float64(cantidadDescontada),
				})
			}
		} else {
			// Producto NO inventariable (Servicio): No resta stock, ubicación es NULL
			detallesParaGuardar = append(detallesParaGuardar, []interface{}{
				nil, detalleReq.ProductoId, nil, int(detalleReq.Cantidad), precioVenta, detalleReq.Descuento,
			})
		}
	}

	// 5. Totales Finales y Descuento General
	totalVenta += request.CostoTiempo
	if request.DescuentoGeneral > totalVenta {
		return nil, datatype.NewBadRequestError("El descuento general supera el total de la venta.")
	}
	totalVenta -= request.DescuentoGeneral

	// 6. Insertar Encabezado
	var ventaId int
	queryVenta := `
        INSERT INTO venta (codigo_venta, sucursal_id, sala_id, uso_sala_id, usuario_id, cliente_id, total, descuento_general, costo_tiempo_venta, observacion, estado, creado_en)
        VALUES (nextval('seq_codigo_venta'), $1, $2, $3, $4, $5, $6, $7, $8, $9, 'Completada', NOW())
        RETURNING id`

	err = tx.QueryRow(ctx, queryVenta, request.SucursalId, request.SalaId, request.UsoSalaId, request.UsuarioId, request.ClienteId, totalVenta, request.DescuentoGeneral, request.CostoTiempo, request.Observacion).Scan(&ventaId)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 7. Inserción Masiva de Detalles
	for i := range detallesParaGuardar {
		detallesParaGuardar[i][0] = ventaId
	}

	_, err = tx.CopyFrom(ctx, pgx.Identifier{"detalle_venta"},
		[]string{"venta_id", "producto_id", "ubicacion_id", "cantidad", "precio_venta", "descuento"},
		pgx.CopyFromRows(detallesParaGuardar))

	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 8. Confirmar Transacción
	if err = tx.Commit(ctx); err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &ventaId, nil
}

func (v VentaRepository) AnularVentaById(ctx context.Context, id *int) error {
	type detalleVentaSimple struct {
		ProductoId  int
		UbicacionId int
		Cantidad    int
	}

	// 1. Iniciar transacción
	tx, err := v.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			if rollErr := tx.Rollback(ctx); rollErr != nil {
				log.Println("Error durante rollback:", rollErr)
			}
		}
	}()

	// 2. Obtener datos de la venta (Estado, UsoSala y CostoTiempo)
	var estadoActual string
	var usoSalaId *int64
	var costoTiempoVenta float64

	// MODIFICADO: Ahora traemos también el uso_sala_id y el costo_tiempo_venta
	queryDatosVenta := `
        SELECT estado, uso_sala_id, costo_tiempo_venta 
        FROM venta 
        WHERE id = $1 
        FOR UPDATE`

	err = tx.QueryRow(ctx, queryDatosVenta, *id).Scan(&estadoActual, &usoSalaId, &costoTiempoVenta)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return datatype.NewNotFoundError("Venta no encontrada")
		}
		log.Println("Error al obtener datos de la venta:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if estadoActual == "Anulada" {
		return datatype.NewBadRequestError("Esta venta ya fue anulada anteriormente.")
	}

	// --- NUEVA LÓGICA: REVERTIR COSTO DE TIEMPO EN USO_SALA ---
	// Si la venta estaba ligada a una sala y tenía un costo de tiempo asociado, lo restamos.
	if usoSalaId != nil && costoTiempoVenta > 0 {
		queryRevertirCosto := `
            UPDATE uso_sala 
            SET 
                costo_tiempo = costo_tiempo - $1, 
                actualizado_en = NOW() 
            WHERE id = $2`

		_, err = tx.Exec(ctx, queryRevertirCosto, costoTiempoVenta, *usoSalaId)
		if err != nil {
			log.Println("Error al revertir el costo de tiempo en uso_sala:", err)
			return datatype.NewInternalServerErrorGeneric()
		}
	}
	// -----------------------------------------------------------

	// 3. Obtener los detalles de la venta (Productos)
	queryDetalles := `
        SELECT producto_id, ubicacion_id, cantidad 
        FROM detalle_venta 
        WHERE venta_id = $1`

	rows, err := tx.Query(ctx, queryDetalles, *id)
	if err != nil {
		log.Println("Error al obtener detalles de la venta:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	var detalles []detalleVentaSimple
	for rows.Next() {
		var d detalleVentaSimple
		if err := rows.Scan(&d.ProductoId, &d.UbicacionId, &d.Cantidad); err != nil {
			rows.Close()
			return datatype.NewInternalServerErrorGeneric()
		}
		detalles = append(detalles, d)
	}
	rows.Close()

	// 4. REVERTIR EL INVENTARIO (Devolver el stock)
	querySumaInventario := `
       INSERT INTO inventario (producto_id, ubicacion_id, stock)
       VALUES ($1, $2, $3)
       ON CONFLICT (producto_id, ubicacion_id) 
       DO UPDATE SET
          stock = inventario.stock + EXCLUDED.stock;
    `
	for _, d := range detalles {
		if d.Cantidad > 0 {
			// Nota: Si el producto no es inventariable (ubicacion_id es NULL), esto fallaría.
			// Deberías validar que d.UbicacionId > 0 o manejar el NULL en la query.
			// Asumiendo que detalle_venta guarda NULL en ubicacion_id para servicios:
			if d.UbicacionId > 0 {
				_, err = tx.Exec(ctx, querySumaInventario, d.ProductoId, d.UbicacionId, d.Cantidad)
				if err != nil {
					log.Println("Error al devolver stock (UPSERT):", err)
					return datatype.NewInternalServerErrorGeneric()
				}
			}
		}
	}

	// 5. Actualizar estado de la Venta
	queryUpdateVenta := `UPDATE venta SET estado = 'Anulada', actualizado_en = NOW() WHERE id = $1`

	_, err = tx.Exec(ctx, queryUpdateVenta, *id)
	if err != nil {
		log.Println("Error al actualizar estado de la venta a Anulada:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// 6. Commit
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción de anulación:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (v VentaRepository) RegistrarPagoVenta(ctx context.Context, ventaId *int, request *domain.RegistrarPagosRequest) (*[]int, error) {
	var pagosValidos []domain.PagoRequest
	for _, pago := range request.Pagos {
		if pago.Monto > 0 {
			pagosValidos = append(pagosValidos, pago)
		}
	}
	// Reemplazamos la lista original con la lista limpia
	request.Pagos = pagosValidos

	// 0. Validar el input
	if len(request.Pagos) == 0 {
		return nil, datatype.NewBadRequestError("Se debe proporcionar al menos un método de pago.")
	}

	// 1. Iniciar transacción y defer rollback
	tx, err := v.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			if rollErr := tx.Rollback(ctx); rollErr != nil {
				log.Println("Error durante rollback:", rollErr)
			}
		}
	}()

	// 2. Obtener la Venta y BLOQUEAR LA FILA
	var totalVenta float64
	var estadoVenta string
	queryLockVenta := `SELECT total, estado FROM venta WHERE id = $1 FOR UPDATE`

	err = tx.QueryRow(ctx, queryLockVenta, *ventaId).Scan(&totalVenta, &estadoVenta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("La venta no fue encontrada.")
		}
		log.Println("Error al bloquear la venta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// Validar el estado de la venta
	if estadoVenta == "Completado" || estadoVenta == "Anulada" {
		return nil, datatype.NewBadRequestError(fmt.Sprintf("Esta venta está en estado '%s' y no se puede pagar.", estadoVenta))
	}

	// 3. Validar PAGO EXACTO (Ni más, ni menos)
	var totalPagoRequest float64 = 0
	for _, pago := range request.Pagos {
		if pago.Monto <= 0 {
			return nil, datatype.NewBadRequestError("El monto de un pago no puede ser cero o negativo.")
		}
		totalPagoRequest += pago.Monto
	}

	const epsilon = 0.001
	if totalPagoRequest < (totalVenta - epsilon) {
		return nil, datatype.NewBadRequestError(
			fmt.Sprintf("El monto pagado (%.2f) es menor al total de la venta (%.2f).", totalPagoRequest, totalVenta),
		)
	}

	// 4. Insertar los nuevos pagos
	var pagoIds []int
	queryPago := `
        INSERT INTO venta_pago (venta_id, metodo_pago_id, monto, referencia) 
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	for _, pago := range request.Pagos {
		var pagoId int
		err = tx.QueryRow(ctx, queryPago, *ventaId, pago.MetodoPagoId, pago.Monto, pago.Referencia).Scan(&pagoId)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, datatype.NewBadRequestError("El método de pago seleccionado no existe.")
			}
			log.Println("Error al insertar el pago:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		pagoIds = append(pagoIds, pagoId)
	}

	// 5. Actualizar estado de la Venta a 'Completado'
	queryUpdateVenta := `UPDATE venta SET estado = 'Completado', actualizado_en = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, queryUpdateVenta, *ventaId)
	if err != nil {
		log.Println("Error al actualizar estado de la venta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 6. Sincronizar el USO DE SALA (Finalizar la sesión si la venta la incluye)
	// Cuando el pago es exacto, cerramos la sesión (solo si aún está activa)
	queryFinalizeUsoSala := `
        UPDATE uso_sala us 
        SET actualizado_en = NOW() 
        FROM venta v 
        WHERE v.id = $1 
          AND us.id = v.uso_sala_id 
          AND us.estado NOT IN ('Finalizado', 'Cancelado');
    `
	// Si la venta está ligada a un uso_sala, este UPDATE lo marca como Finalizado.
	// Si la venta es 'al paso' (uso_sala_id es NULL), esta consulta no hace nada (lo cual es correcto).
	_, err = tx.Exec(ctx, queryFinalizeUsoSala, *ventaId)
	if err != nil {
		log.Println("Error al finalizar uso_sala:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 7. Commit
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción de pago:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &pagoIds, nil
}

func (v VentaRepository) ObtenerVenta(ctx context.Context, id *int) (*domain.Venta, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `
SELECT 
    v.id,
    v.codigo_venta,
    v.total,
    v.estado,
    v.creado_en,
    v.actualizado_en, 
	v.costo_tiempo_venta,
	v.descuento_general,
	v.observacion,
    -- Construye el objeto 'usuario' (el admin/cajero)
    json_build_object(
       'id',ua.id,
       'username',ua.username
    ) AS usuario,
    -- Construye el objeto 'cliente' (enlazado a la venta)
    json_build_object(
       'id', c.id,
       'nombres', c.nombres,
       'apellidos', c.apellidos,
       'codigoPais', c.codigo_pais,
       'celular', c.celular,
       'fechaNacimiento', c.fecha_nacimiento,
       'estado', c.estado,
       'creadoEn', c.creado_en
    ) AS cliente,

    -- Construye el objeto 'sucursal'
    jsonb_build_object(
       'id', s1.id,
       'nombre', s1.nombre,
       'estado', s1.estado,
       'creadoEn', s1.creado_en
    ) AS sucursal,

    -- Construye el objeto 'sala' (con su dispositivo y usuario)
    (CASE WHEN (s2.id IS NOT NULL) THEN
		jsonb_build_object(
			'id',s2.id,
			'nombre',s2.nombre,
			'estado',s2.estado,
			'creadoEn',s2.creado_en,
			'actualizadoEn',s2.actualizado_en,
			'eliminadoEn',s2.eliminado_en,
			'dispositivo',jsonb_build_object(
				'id', d.id,
				'dispositivoId', d.dispositivo_id,
				'nombre', d.nombre,
				'estado', d.estado,
				'creadoEn', d.creado_en,
				'usuario', COALESCE(
			jsonb_build_object(
				 'id', u.id,
				 'username', u.username
			), '{}'::jsonb)
		)
	) ELSE 'null'::jsonb END) AS sala,

    -- Construye el objeto 'uso' (uso_sala)
    (
       CASE WHEN  (us.id IS NOT NULL)THEN
          jsonb_build_object(
             'cliente', (
                CASE WHEN c1.id IS NOT NULL THEN jsonb_build_object(
                   'id', c1.id,
                   'nombres', c1.nombres,
                   'apellidos', c1.apellidos,
                   'codigoPais', c1.codigo_pais,
                   'celular', c1.celular,
                   'fechaNacimiento', c1.fecha_nacimiento,
                   'estado', c1.estado,
                   'creadoEn', c1.creado_en
                ) ELSE '{}'::jsonb END
             ),
             'id', us.id,
             'inicio', us.inicio,
             'fin', us.fin,
             'pausadoEn', us.pausado_en,
             'costoTiempo',us.costo_tiempo,
             'duracionPausa', EXTRACT(EPOCH FROM COALESCE(us.duracion_pausa, '0')),
             'tiempoUso', EXTRACT(EPOCH FROM (COALESCE(us.fin, NOW()) - us.inicio - COALESCE(us.duracion_pausa, '0'))),
             'estado', us.estado
          )
       ELSE 'null'::jsonb END
    ) AS uso,
    
    -- SUB-CONSULTA 1: Obtiene solo los detalles
    (SELECT COALESCE(json_agg(
        json_build_object(
			'producto', json_build_object(
			'id',p.id,
			'nombre',p.nombre,
			'estado',p.estado,
			'urlFoto',($1::text || p.id::text || '/' || p.foto),
			'creadoEn',p.creado_en,
			'actualizadoEn',p.actualizado_en,
			'eliminadoEn',p.eliminado_en
			),
			'ubicacion', json_build_object(
				'id',u2.id,
				'nombre',u2.nombre,
				'estado',u2.estado,
				'esVendible',u2.es_vendible,
				'prioridadVenta',u2.prioridad_venta
			),
            'id',dv.id,
            'cantidad',dv.cantidad,
            'descuento',dv.descuento,
            'precioVenta',dv.precio_venta
        )
    ORDER BY dv.id), '[]')
     FROM public.detalle_venta dv
     LEFT JOIN public.producto p on dv.producto_id = p.id
     LEFT JOIN public.ubicacion u2 on dv.ubicacion_id = u2.id
     WHERE dv.venta_id = v.id
    ) AS detalles,
    
    -- SUB-CONSULTA 2: Obtiene solo los pagos
    (SELECT COALESCE(json_agg(
        json_build_object(
           'metodoPago', json_build_object(
               'id',mp.id,
               'nombre',mp.nombre,
               'estado',mp.estado
           ),
             'monto',vp.monto,
             'referencia',vp.referencia
        )
    ORDER BY vp.id), '[]')
     FROM public.venta_pago vp
     LEFT JOIN public.metodo_pago mp on vp.metodo_pago_id = mp.id
     WHERE vp.venta_id = v.id
    ) AS pagos
FROM venta v
LEFT JOIN public.usuario_admin ua on v.usuario_id = ua.id
LEFT JOIN public.cliente c on v.cliente_id = c.id
LEFT JOIN public.sucursal s1 on v.sucursal_id = s1.id
LEFT JOIN public.sala s2 on v.sala_id = s2.id
LEFT JOIN public.dispositivo d on s2.dispositivo_id = d.id
LEFT JOIN public.usuario u on d.usuario_id = u.id
LEFT JOIN public.uso_sala us on v.uso_sala_id = us.id
LEFT JOIN public.cliente c1 on us.cliente_id = c1.id
WHERE v.id = $2
GROUP BY v.id, ua.id, c.id, s1.id, s2.id, d.id, u.id, us.id, c1.id
LIMIT 1
	`
	var item domain.Venta
	err := v.pool.QueryRow(ctx, query, fullHostname, *id).
		Scan(&item.Id, &item.CodigoVenta, &item.Total, &item.Estado, &item.CreadoEn, &item.ActualizadoEn, &item.CostoTiempoVenta, &item.DescuentoGeneral, &item.Observacion, &item.Usuario, &item.Cliente, &item.Sucursal, &item.Sala, &item.UsoSala, &item.Detalles, &item.Pagos)
	if err != nil {
		log.Println("Error al obtener compra:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Venta no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func NewVentaRepository(pool *pgxpool.Pool) *VentaRepository {
	return &VentaRepository{pool: pool}
}

var _ port.VentaRepository = (*VentaRepository)(nil)
