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

type InventarioRepository struct {
	pool *pgxpool.Pool
}

func (i InventarioRepository) ListarTransferencias(ctx context.Context, filtros map[string]string) (*[]domain.TransferenciaInventarioInfo, error) {
	var filters []string
	var args []interface{}
	var j = 1
	if usuarioIdStr := filtros["usuarioId"]; usuarioIdStr != "" {
		usuarioId, err := strconv.Atoi(usuarioIdStr)
		if err != nil {
			log.Println("Error al convertir usuarioId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de usuarioId no es válido")
		}
		filters = append(filters, fmt.Sprintf("u.id = $%d", j))
		args = append(args, usuarioId)
		j++
	}

	if ubicacionOrigenIdStr := filtros["ubicacionOrigenId"]; ubicacionOrigenIdStr != "" {
		ubicacionOrigenId, err := strconv.Atoi(ubicacionOrigenIdStr)
		if err != nil {
			log.Println("Error al convertir ubicacionOrigenId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de ubicacionOrigenId no es válido")
		}
		filters = append(filters, fmt.Sprintf("uo.id = $%d", j))
		args = append(args, ubicacionOrigenId)
		j++
	}

	if ubicacionDestinoIdStr := filtros["ubicacionDestinoId"]; ubicacionDestinoIdStr != "" {
		ubicacionDestinoId, err := strconv.Atoi(ubicacionDestinoIdStr)
		if err != nil {
			log.Println("Error al convertir ubicacionDestinoId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de ubicacionDestinoId no es válido")
		}
		filters = append(filters, fmt.Sprintf("ud.id = $%d", j))
		args = append(args, ubicacionDestinoId)
		j++
	}

	query := `
SELECT 
    t.id,
    t.motivo,
    t.fecha,
    json_build_object(
    	'id',u.id,
		'username',u.username
    ) AS usuario,
	json_build_object(
		'id',uo.id,
		'nombre',uo.nombre,
		'estado',uo.estado,
		'esVendible',uo.es_vendible,
		'prioridadVenta',uo.prioridad_venta
    ) AS ubicacion_origen,
	json_build_object(
		'id',ud.id,
		'nombre',ud.nombre,
		'estado',ud.estado,
		'esVendible',ud.es_vendible,
		'prioridadVenta',ud.prioridad_venta
    ) AS ubicacion_destino
FROM transferencia t
LEFT JOIN public.usuario_admin u ON t.usuario_id = u.id
LEFT JOIN public.ubicacion uo ON t.ubicacion_origen_id = uo.id
LEFT JOIN public.ubicacion ud ON t.ubicacion_destino_id = ud.id
`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY t.id DESC"
	rows, err := i.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al ejecutar consulta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.TransferenciaInventarioInfo, 0)
	for rows.Next() {
		var item domain.TransferenciaInventarioInfo
		err := rows.Scan(&item.Id, &item.Motivo, &item.Fecha, &item.Usuario, &item.UbicacionOrigen, &item.UbicacionDestino)
		if err != nil {
			log.Println("Error al escanear ajuste_inventario", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (i InventarioRepository) ObtenerTransferenciaById(ctx context.Context, id *int) (*domain.TransferenciaInventario, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `
SELECT 
    t.id,
    t.motivo,
    t.fecha,
    json_build_object(
    	'id',u.id,
		'username',u.username
    ) AS usuario,
	json_build_object(
		'id',uo.id,
		'nombre',uo.nombre,
		'estado',uo.estado,
		'esVendible',uo.es_vendible,
		'prioridadVenta',uo.prioridad_venta
    ) AS ubicacion_origen,
	json_build_object(
		'id',ud.id,
		'nombre',ud.nombre,
		'estado',ud.estado,
		'esVendible',ud.es_vendible,
		'prioridadVenta',ud.prioridad_venta
    ) AS ubicacion_destino,
    COALESCE(
       json_agg(
          CASE 
             WHEN dt.id IS NULL THEN NULL
             ELSE json_build_object(
             	'id',dt.id,
                'cantidad',dt.cantidad,
                'producto', json_build_object(
                    'id',p.id,
                    'nombre',p.nombre,
                    'estado',p.estado,
                    'urlFoto',($1::text || p.id::text || '/' || p.foto),
                    'precio',p.precio,
                    'creadoEn',p.creado_en,
                    'actualizadoEn',p.actualizado_en,
                    'eliminadoEn',p.eliminado_en
                )
             )
          END
       ), 
       '[]'
    ) AS detalles
FROM transferencia t
LEFT JOIN public.usuario_admin u ON t.usuario_id = u.id
LEFT JOIN public.ubicacion uo ON t.ubicacion_origen_id = uo.id
LEFT JOIN public.ubicacion ud ON t.ubicacion_destino_id = ud.id
LEFT JOIN detalle_transferencia dt ON t.id = dt.transferencia_id
LEFT JOIN public.producto p on dt.producto_id = p.id
WHERE t.id = $2
GROUP BY t.id,u.id,uo.id,ud.id
LIMIT 1
`
	var item domain.TransferenciaInventario
	err := i.pool.QueryRow(ctx, query, fullHostname, *id).Scan(&item.Id, &item.Motivo, &item.Fecha, &item.Usuario, &item.UbicacionOrigen, &item.UbicacionDestino, &item.Detalles)
	if err != nil {
		log.Println("Error al consultar:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Ajuste de inventario no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func (i InventarioRepository) ObtenerAjusteById(ctx context.Context, id *int) (*domain.AjusteInventario, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `
SELECT 
    ai.id,
    ai.tipo_ajuste,
    ai.motivo,
    json_build_object(
        'id',ua.id,
        'username',ua.username
    ) AS usuario,
    json_build_object(
        'id',s.id,
        'nombre',s.nombre,
        'estado',s.estado,
        'creadoEn',s.creado_en
    ) AS sucursal,
    COALESCE(
       json_agg(
          CASE 
             WHEN dai.id IS NULL THEN NULL
             ELSE json_build_object(
             	'id',dai.id,
                'cantidad',dai.cantidad,
                'producto', json_build_object(
                    'id',p.id,
                    'nombre',p.nombre,
                    'estado',p.estado,
                    'urlFoto',($1::text || p.id::text || '/' || p.foto),
                    'precio',p.precio,
                    'creadoEn',p.creado_en,
                    'actualizadoEn',p.actualizado_en,
                    'eliminadoEn',p.eliminado_en
                ),
                'ubicacion', json_build_object(
                    'id',u.id,
                    'nombre',u.nombre,
                    'estado',u.estado,
                    'esVendible',u.es_vendible,
                    'prioridadVenta',u.prioridad_venta
                )
             )
          END
       ), 
       '[]'
    ) AS detalles
FROM ajuste_inventario ai
LEFT JOIN public.usuario_admin ua on ai.usuario_id = ua.id
LEFT JOIN public.sucursal s on ai.sucursal_id = s.id
LEFT JOIN public.detalle_ajuste_inventario dai on ai.id = dai.ajuste_inventario_id
LEFT JOIN public.producto p on dai.producto_id = p.id
LEFT JOIN public.ubicacion u on dai.ubicacion_id = u.id
WHERE ai.id = $2
GROUP BY ai.id, ua.id, s.id
`
	var item domain.AjusteInventario
	err := i.pool.QueryRow(ctx, query, fullHostname, *id).Scan(&item.Id, &item.TipoAjuste, &item.Motivo, &item.Usuario, &item.Sucursal, &item.Detalles)
	if err != nil {
		log.Println("Error al consultar:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Ajuste de inventario no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func (i InventarioRepository) ListarAjustes(ctx context.Context, filtros map[string]string) (*[]domain.AjusteInventarioInfo, error) {
	var filters []string
	var args []interface{}
	var j = 1
	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s.id = $%d", j))
		args = append(args, sucursalId)
		j++
	}

	query := `
SELECT 
    ai.id,
    ai.tipo_ajuste,
    ai.motivo,
    json_build_object(
    	'id',ua.id,
    	'username',ua.username
    ) AS usuario,
    json_build_object(
    	'id',s.id,
    	'nombre',s.nombre,
    	'estado',s.estado,
    	'creadoEn',s.creado_en
    ) AS sucursal
FROM ajuste_inventario ai
LEFT JOIN public.usuario_admin ua on ai.usuario_id = ua.id
LEFT JOIN public.sucursal s on ai.sucursal_id = s.id
`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY ai.id DESC"
	rows, err := i.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al ejecutar consulta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.AjusteInventarioInfo, 0)
	for rows.Next() {
		var item domain.AjusteInventarioInfo
		err := rows.Scan(&item.Id, &item.TipoAjuste, &item.Motivo, &item.Usuario, &item.Sucursal)
		if err != nil {
			log.Println("Error al escanear ajuste_inventario", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (i InventarioRepository) RegistrarTransferencia(ctx context.Context, request *domain.TransferenciaRequest) (*int, error) {

	// 0. Validaciones de entrada
	if request.UbicacionOrigenId == request.UbicacionDestinoId {
		return nil, datatype.NewBadRequestError("La ubicación de origen y destino no pueden ser la misma.")
	}
	if len(request.Detalles) == 0 {
		return nil, datatype.NewBadRequestError("La transferencia debe contener al menos un detalle.")
	}

	// 1. Iniciar transacción
	tx, err := i.pool.Begin(ctx)
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

	// 2. INSERTAR ENCABEZADO (transferencia)
	var transferenciaId int
	queryEncabezado := `
        INSERT INTO transferencia (ubicacion_origen_id, ubicacion_destino_id, usuario_id, motivo, fecha) 
        VALUES($1, $2, $3, $4, NOW()) 
        RETURNING id`

	err = tx.QueryRow(ctx, queryEncabezado, request.UbicacionOrigenId, request.UbicacionDestinoId, request.UsuarioId, request.Motivo).Scan(&transferenciaId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			log.Println("Error de FK al insertar encabezado de transferencia:", pgErr.Message)
			var errorMsg string

			switch pgErr.ConstraintName {
			case "transferencia_ubicacion_origen_id_fkey":
				errorMsg = "La ubicación de origen seleccionada no existe."
			case "transferencia_ubicacion_destino_id_fkey":
				errorMsg = "La ubicación de destino seleccionada no existe."
			case "transferencia_usuario_id_fkey":
				errorMsg = "El usuario administrador que registra la transferencia no existe."
			default:
				errorMsg = "Error de referencia: uno de los campos del encabezado es inválido."
			}
			return nil, datatype.NewBadRequestError(errorMsg)
		}
		log.Println("Error al insertar encabezado de transferencia:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 3. BUCLE: Aplicar Lógica de Stock y Recolectar para CopyFrom
	var rowsDetalle [][]interface{}

	queryCheckStock := `SELECT stock FROM inventario WHERE producto_id = $1 AND ubicacion_id = $2 FOR UPDATE`

	queryUpdateOrigen := `UPDATE inventario SET stock = stock - $1 WHERE producto_id = $2 AND ubicacion_id = $3`

	querySumaDestino := `
       INSERT INTO inventario (producto_id, ubicacion_id, stock)
       VALUES ($1, $2, $3)
       ON CONFLICT (producto_id, ubicacion_id) 
       DO UPDATE SET stock = inventario.stock + EXCLUDED.stock;
    `

	for _, detalle := range request.Detalles {

		if detalle.Cantidad <= 0 {
			return nil, datatype.NewBadRequestError("La cantidad a transferir debe ser mayor a cero.")
		}

		// RESTAR de la Ubicación de Origen (con chequeo de stock)
		var stockOrigen int64
		err = tx.QueryRow(ctx, queryCheckStock, detalle.ProductoId, request.UbicacionOrigenId).Scan(&stockOrigen)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, datatype.NewBadRequestError(fmt.Sprintf("No hay stock del producto %d en la ubicación de origen.", detalle.ProductoId))
			}
			log.Println("Error al consultar stock de origen:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		if stockOrigen < detalle.Cantidad {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("Stock insuficiente en Origen para producto %d. Actual: %d, Se intenta transferir: %d", detalle.ProductoId, stockOrigen, detalle.Cantidad))
		}

		_, err = tx.Exec(ctx, queryUpdateOrigen, detalle.Cantidad, detalle.ProductoId, request.UbicacionOrigenId)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				return nil, datatype.NewBadRequestError(fmt.Sprintf("Error de referencia en el encabezado: %s", pgErr.ConstraintName))
			}
			// Fallback por si la validación de Go falla
			if errors.As(err, &pgErr) && pgErr.Code == "23514" && pgErr.ConstraintName == "check_diferente_ubicacion" {
				return nil, datatype.NewBadRequestError("La ubicación de origen y destino no pueden ser la misma.")
			}
			log.Println("Error al restar stock de origen:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		// SUMAR a la Ubicación de Destino (UPSERT)
		_, err = tx.Exec(ctx, querySumaDestino, detalle.ProductoId, request.UbicacionDestinoId, detalle.Cantidad)
		if err != nil {
			log.Println("Error al sumar stock a destino (UPSERT):", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		// 3c. Recolectar para el INSERT masivo de detalles
		rowsDetalle = append(rowsDetalle, []interface{}{
			transferenciaId,
			detalle.ProductoId,
			detalle.Cantidad,
		})
	}

	// INSERTAR DETALLES (Usando CopyFrom)
	columnasDetalle := []string{"transferencia_id", "producto_id", "cantidad"}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"detalle_transferencia"},
		columnasDetalle,
		pgx.CopyFromRows(rowsDetalle),
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			return nil, datatype.NewBadRequestError("Error en la cantidad de un detalle: la cantidad debe ser mayor a cero.")
		}
		log.Println("Error durante la inserción masiva de detalles de transferencia:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 5. Commit de la transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción de transferencia:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &transferenciaId, nil
}

func (i InventarioRepository) RegistrarAjusteConDetalle(ctx context.Context, request *domain.AjusteInventarioRequest) (*int, error) {

	// Iniciar transacción
	tx, err := i.pool.Begin(ctx)
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

	// VALIDACIÓN N+1: Verificar que todas las ubicaciones pertenezcan a la sucursal
	var ubicacionIds []int
	for _, detalle := range request.Detalles {
		ubicacionIds = append(ubicacionIds, detalle.UbicacionId)

		if detalle.Cantidad == 0 {
			// La BD ya lo valida (CHECK <> 0), pero fallar aquí es más rápido.
			return nil, datatype.NewBadRequestError("La cantidad a ajustar no puede ser cero.")
		}
	}

	var countUbicaciones int
	queryValidarUbi := `SELECT COUNT(id) FROM ubicacion WHERE sucursal_id = $1 AND id = ANY($2)`

	err = tx.QueryRow(ctx, queryValidarUbi, request.SucursalId, ubicacionIds).Scan(&countUbicaciones)
	if err != nil {
		log.Println("Error al validar ubicaciones en lote:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if countUbicaciones != len(ubicacionIds) {
		return nil, datatype.NewBadRequestError("Una o más ubicaciones no son válidas o no pertenecen a la sucursal del ajuste.")
	}

	// INSERTAR ENCABEZADO (ajuste_inventario)
	var ajusteId int
	queryEncabezado := `
        INSERT INTO ajuste_inventario (usuario_id, sucursal_id, motivo, tipo_ajuste, fecha) 
        VALUES($1, $2, $3, $4, NOW()) 
        RETURNING id`

	err = tx.QueryRow(ctx, queryEncabezado, request.UsuarioId, request.SucursalId, request.Motivo, request.TipoAjuste).Scan(&ajusteId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("Error de referencia en el encabezado: %s", pgErr.ConstraintName))
		}
		log.Println("Error al insertar encabezado de ajuste:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// BUCLE: Aplicar Ajustes de Stock y Recolectar para CopyFrom
	var rowsDetalle [][]interface{}

	querySumaInventario := `
       INSERT INTO inventario (producto_id, ubicacion_id, stock)
       VALUES ($1, $2, $3)
       ON CONFLICT (producto_id, ubicacion_id) 
       DO UPDATE SET
          stock = inventario.stock + EXCLUDED.stock;
    `

	queryRestaInventario := `
        UPDATE inventario 
        SET stock = stock - $1 
        WHERE producto_id = $2 AND ubicacion_id = $3
    `
	// Query para la validación de CARGA_INICIAL
	queryCheckExist := `SELECT 1 FROM inventario WHERE producto_id = $1 AND ubicacion_id = $2`

	for _, detalle := range request.Detalles {

		// 4a. VALIDACIÓN DE CARGA_INICIAL (si aplica)
		if request.TipoAjuste == "CARGA_INICIAL" {
			var existe int
			errStock := tx.QueryRow(ctx, queryCheckExist, detalle.ProductoId, detalle.UbicacionId).Scan(&existe)

			if errStock == nil {
				// Lo encontró (incluso con stock 0), rechazamos.
				return nil, datatype.NewBadRequestError(fmt.Sprintf(
					"CARGA_INICIAL fallida: El producto %d en la ubicación %d ya existe en el inventario.",
					detalle.ProductoId, detalle.UbicacionId,
				))
			}
			if !errors.Is(errStock, pgx.ErrNoRows) {
				// Error de BD real
				log.Println("Error al verificar stock para CARGA_INICIAL:", errStock)
				return nil, datatype.NewInternalServerErrorGeneric()
			}
			// Si es pgx.ErrNoRows, continúa (es el caso OK)
		}

		// APLICAR AJUSTE (Lógica Separada)
		if detalle.Cantidad > 0 {
			// --- LÓGICA DE SUMA (Entrada) ---
			_, err = tx.Exec(ctx, querySumaInventario, detalle.ProductoId, detalle.UbicacionId, detalle.Cantidad)
			if err != nil {
				log.Println("Error al sumar stock (UPSERT):", err)
				return nil, datatype.NewInternalServerErrorGeneric()
			}

		} else if detalle.Cantidad < 0 {
			cantidadARestar := -detalle.Cantidad

			ct, err := tx.Exec(ctx, queryRestaInventario, cantidadARestar, detalle.ProductoId, detalle.UbicacionId)

			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23514" {
					// Atrapa el CHECK (stock >= 0)
					return nil, datatype.NewBadRequestError(fmt.Sprintf("Stock insuficiente para el producto %d en la ubicación %d.", detalle.ProductoId, detalle.UbicacionId))
				}
				log.Println("Error al restar stock (UPDATE):", err)
				return nil, datatype.NewInternalServerErrorGeneric()
			}

			if ct.RowsAffected() == 0 {
				// No se puede restar de un producto/ubicación que no existe en el inventario
				return nil, datatype.NewBadRequestError(fmt.Sprintf("No se puede restar stock: El producto %d no existe en la ubicación %d.", detalle.ProductoId, detalle.UbicacionId))
			}
		}

		// Recolectar para el INSERT masivo de detalles
		rowsDetalle = append(rowsDetalle, []interface{}{
			ajusteId,
			detalle.ProductoId,
			detalle.UbicacionId,
			detalle.Cantidad,
		})
	}

	// INSERTAR DETALLES (Usando CopyFrom)
	columnasDetalle := []string{"ajuste_inventario_id", "producto_id", "ubicacion_id", "cantidad"}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"detalle_ajuste_inventario"},
		columnasDetalle,
		pgx.CopyFromRows(rowsDetalle),
	)
	if err != nil {
		log.Println("Error durante la inserción masiva de detalles de ajuste:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 6. Commit de la transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &ajusteId, nil
}

func (i InventarioRepository) ListarInventario(ctx context.Context, filtros map[string]string) (*[]domain.Inventario, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")

	var filters []string
	var args = []interface{}{fullHostname}
	var j = 2
	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s.id = $%d", j))
		args = append(args, sucursalId)
		j++
	}

	if ubicacionIdStr := filtros["ubicacionId"]; ubicacionIdStr != "" {
		ubicacionId, err := strconv.Atoi(ubicacionIdStr)
		if err != nil {
			log.Println("Error al convertir ubicacionId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de ubicacionId no es válido")
		}
		filters = append(filters, fmt.Sprintf("u.id = $%d", j))
		args = append(args, ubicacionId)
		j++
	}

	query := `
SELECT 
    i.id,
    i.stock,
	json_build_object(
		'id',p.id,
		'nombre',p.nombre,
		'estado',p.estado,
		'urlFoto',($1::text || p.id::text || '/' || p.foto),
		'precio',p.precio,
		'creadoEn',p.creado_en,
		'actualizadoEn',p.actualizado_en,
		'eliminadoEn',p.eliminado_en
	) AS producto,
	json_build_object(
		'id',u.id,
		'nombre',u.nombre,
		'estado',u.estado,
		'esVendible',u.es_vendible,
		'prioridadVenta',u.prioridad_venta
	) AS ubicacion
FROM inventario i
LEFT JOIN public.producto p on p.id = i.producto_id
LEFT JOIN public.ubicacion u on i.ubicacion_id = u.id
LEFT JOIN public.sucursal s on s.id = u.sucursal_id
`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY i.producto_id,i.ubicacion_id "
	rows, err := i.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al ejecutar consulta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.Inventario, 0)
	for rows.Next() {
		var item domain.Inventario
		err := rows.Scan(&item.Id, &item.Stock, &item.Producto, &item.Ubicacion)
		if err != nil {
			log.Println("Error al escanear inventario:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func NewInventarioRepository(pool *pgxpool.Pool) *InventarioRepository {
	return &InventarioRepository{pool: pool}
}

var _ port.InventarioRepository = (*InventarioRepository)(nil)
