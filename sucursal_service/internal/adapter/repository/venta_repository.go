package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
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

	// 2. Obtener SucursalId (basado en la SalaId del request)
	var sucursalId int
	querySucursal := `SELECT sucursal_id FROM sala WHERE id = $1`
	err = tx.QueryRow(ctx, querySucursal, request.SalaId).Scan(&sucursalId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewBadRequestError("La sala (POS) seleccionada no existe.")
		}
		log.Println("Error al buscar sucursalId por salaId:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	if request.UsoSalaId != nil {
		queryUsoSala := `
            UPDATE uso_sala 
            SET costo_tiempo=$1, 
                actualizado_en = NOW() 
            WHERE id = $2 
              AND estado IN ('En uso', 'Pausado')
        `

		ct, err := tx.Exec(ctx, queryUsoSala, request.CostoTiempo, *request.UsoSalaId)
		if err != nil {
			log.Println("Error al actualizar uso_sala:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		if ct.RowsAffected() == 0 {
			// Si el uso_sala existe pero no está 'En uso'/'Pausado', se rechaza la venta.
			return nil, datatype.NewBadRequestError("La sesión de uso de sala no está activa y no puede registrar más ventas.")
		}
	}
	// (Ya no es necesario obtener el 'codigoVenta' aquí, se hace en el INSERT)
	var totalVenta float64 = 0
	var detallesParaGuardar [][]interface{} // Para el CopyFrom

	// Queries que se usarán en el bucle
	queryGetPrecio := `SELECT precio FROM producto WHERE id = $1`
	queryFindStock := `
        SELECT i.stock, i.ubicacion_id
        FROM inventario i
        JOIN ubicacion u ON i.ubicacion_id = u.id
        WHERE i.producto_id = $1 
          AND u.sucursal_id = $2 
          AND u.es_vendible = true 
          AND u.estado = 'Activo'
        ORDER BY u.prioridad_venta
        FOR UPDATE OF i`
	queryRestaStock := `UPDATE inventario SET stock = stock - $1 WHERE producto_id = $2 AND ubicacion_id = $3`

	// 4. Bucle de Lógica de Inventario
	for _, detalleReq := range request.Detalles {
		if detalleReq.Cantidad <= 0 {
			return nil, datatype.NewBadRequestError("Cantidad de producto debe ser mayor a cero")
		}

		var precioVenta float64
		err = tx.QueryRow(ctx, queryGetPrecio, detalleReq.ProductoId).Scan(&precioVenta)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, datatype.NewBadRequestError(fmt.Sprintf("Producto ID %d no encontrado.", detalleReq.ProductoId))
			}
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		rows, err := tx.Query(ctx, queryFindStock, detalleReq.ProductoId, sucursalId)
		if err != nil {
			log.Println("Error al buscar stock:", err)
			rows.Close()
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		var stockDisponibleList []stockDisponible
		var stockTotalVendible = 0
		for rows.Next() {
			var s stockDisponible
			if err := rows.Scan(&s.Stock, &s.UbicacionId); err != nil {
				rows.Close()
				return nil, datatype.NewInternalServerErrorGeneric()
			}
			if s.Stock > 0 {
				stockDisponibleList = append(stockDisponibleList, s)
				stockTotalVendible += s.Stock
			}
		}
		rows.Close()

		if stockTotalVendible < int(detalleReq.Cantidad) {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("Stock insuficiente para producto ID %d. Solicitado: %d, Disponible: %d", detalleReq.ProductoId, detalleReq.Cantidad, stockTotalVendible))
		}

		cantidadARestar := int(detalleReq.Cantidad)
		totalVenta += precioVenta * float64(detalleReq.Cantidad)

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

			_, err = tx.Exec(ctx, queryRestaStock, cantidadDescontada, detalleReq.ProductoId, s.UbicacionId)
			if err != nil {
				log.Println("Error al restar stock:", err)
				return nil, datatype.NewInternalServerErrorGeneric()
			}

			detallesParaGuardar = append(detallesParaGuardar, []interface{}{
				nil,
				detalleReq.ProductoId,
				s.UbicacionId,
				cantidadDescontada,
				precioVenta,
			})
		}
	}

	totalVenta += request.CostoTiempo
	// 5. Insertar el Encabezado (venta)
	var ventaId int
	// CONSULTA CORREGIDA
	queryVenta := `
        INSERT INTO venta (codigo_venta, sucursal_id, sala_id, uso_sala_id, usuario_id, cliente_id, total, estado, creado_en)
        VALUES (nextval('seq_codigo_venta'), $1, $2, $3, $4, $5, $6, 'Pendiente', NOW())
        RETURNING id`

	err = tx.QueryRow(ctx, queryVenta,
		sucursalId,        // $1
		request.SalaId,    // $2
		request.UsoSalaId, // $3
		request.UsuarioId, // $4
		request.ClienteId, // $5
		totalVenta,        // $6
	).Scan(&ventaId)

	if err != nil {
		log.Println("Error al insertar encabezado de venta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 6. Insertar los Detalles (CopyFrom)
	for i := range detallesParaGuardar {
		detallesParaGuardar[i][0] = ventaId
	}
	columnasDetalle := []string{"venta_id", "producto_id", "ubicacion_id", "cantidad", "precio_venta"}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"detalle_venta"},
		columnasDetalle,
		pgx.CopyFromRows(detallesParaGuardar),
	)
	if err != nil {
		log.Println("Error durante la inserción masiva de detalles de venta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 7. Commit
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción de venta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	// Firma de retorno CORREGIDA
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

	// 2. Verificar estado de la venta (y bloquear la fila)
	var estadoActual string
	queryEstado := `SELECT estado FROM venta WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryEstado, *id).Scan(&estadoActual)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return datatype.NewNotFoundError("Venta no encontrada")
		}
		log.Println("Error al obtener estado de la venta:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if estadoActual == "Completado" {
		return datatype.NewBadRequestError("Esta venta ya fue completado.")
	}
	if estadoActual == "Anulada" {
		return datatype.NewBadRequestError("Esta venta ya fue anulada anteriormente.")
	}

	// 3. Obtener los detalles de la venta
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
	// (Solo si la venta no estaba 'Pendiente'. Si estaba 'Pendiente', el stock ya se descontó)

	querySumaInventario := `
       INSERT INTO inventario (producto_id, ubicacion_id, stock)
       VALUES ($1, $2, $3)
       ON CONFLICT (producto_id, ubicacion_id) 
       DO UPDATE SET
          stock = inventario.stock + EXCLUDED.stock;
    `
	// (Si la venta fue 'Completado' o 'Pendiente', el stock ya se descontó, así que lo devolvemos)
	for _, d := range detalles {
		if d.Cantidad > 0 {
			_, err = tx.Exec(ctx, querySumaInventario, d.ProductoId, d.UbicacionId, d.Cantidad)
			if err != nil {
				log.Println("Error al devolver stock (UPSERT):", err)
				return datatype.NewInternalServerErrorGeneric()
			}
		}
	}

	// (Opcional: ¿Qué hacer con los pagos? Si la venta estaba 'Completado',
	// deberías registrar una devolución de dinero, pero eso es una lógica más compleja
	// Por ahora, solo anulamos la venta).

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
	if math.Abs(totalPagoRequest-totalVenta) > epsilon {
		return nil, datatype.NewBadRequestError(
			fmt.Sprintf("El monto pagado (%.2f) no coincide con el total de la venta (%.2f).", totalPagoRequest, totalVenta),
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
        SET estado = 'Finalizado', actualizado_en = NOW() 
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
    ) AS sala,

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
			'precio',p.precio, -- (Usando 'precio' de tu DDL)
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
		Scan(&item.Id, &item.CodigoVenta, &item.Total, &item.Estado, &item.CreadoEn, &item.ActualizadoEn, &item.Usuario, &item.Cliente, &item.Sucursal, &item.Sala, &item.UsoSala, &item.Detalles, &item.Pagos)
	if err != nil {
		log.Println("Error al obtener compra:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Venta no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
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
	query := `
	SELECT 
		v.id,
		v.codigo_venta,
		v.total,
		v.estado,
		v.creado_en,
		v.actualizado_en,
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
	`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += `
GROUP BY v.id,c.id,d.id,ua.id,s1.id,s2.id,u.id
ORDER BY v.id
`
	rows, err := v.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.VentaInfo, 0)
	for rows.Next() {
		var item domain.VentaInfo
		err := rows.Scan(&item.Id, &item.CodigoVenta, &item.Total, &item.Estado, &item.CreadoEn, &item.ActualizadoEn, &item.Usuario, &item.Cliente, &item.Sucursal, &item.Sala)
		if err != nil {
			log.Println("Error al obtener lista de venta:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}

	return &list, nil
}

func NewVentaRepository(pool *pgxpool.Pool) *VentaRepository {
	return &VentaRepository{pool: pool}
}

var _ port.VentaRepository = (*VentaRepository)(nil)
