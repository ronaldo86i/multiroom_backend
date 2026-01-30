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

type CompraRepository struct {
	pool *pgxpool.Pool
}

func (c CompraRepository) ListarCompras(ctx context.Context, filtros map[string]string) (*[]domain.CompraInfo, error) {
	var filters []string
	var args []interface{}
	var j = 1
	if usuarioIdStr := filtros["usuarioId"]; usuarioIdStr != "" {
		usuarioId, err := strconv.Atoi(usuarioIdStr)
		if err != nil {
			log.Println("Error al convertir usuarioId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de usuarioId no es válido")
		}
		filters = append(filters, fmt.Sprintf("c.usuario_admin_id = $%d", j))
		args = append(args, usuarioId)
		j++
	}

	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("c.sucursal_id = $%d", j))
		args = append(args, sucursalId)
		j++
	}

	if proveedorIdStr := filtros["proveedorId"]; proveedorIdStr != "" {
		proveedorId, err := strconv.Atoi(proveedorIdStr)
		if err != nil {
			log.Println("Error al convertir proveedorId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de proveedorId no es válido")
		}
		filters = append(filters, fmt.Sprintf("c.proveedor_id = $%d", j))
		args = append(args, proveedorId)
		j++
	}

	if estado := filtros["estado"]; estado != "" {
		filters = append(filters, fmt.Sprintf("c.estado = $%d", j))
		args = append(args, estado)
		j++
	}

	query := `
SELECT 
    c.id,
    c.codigo_compra,
    c.estado,
    c.creado_en,
    c.actualizado_en,
    c.eliminado_en,
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
    json_build_object(
    	'id',p.id,
      	'nombre',p.nombre,
    	'estado',p.estado,
    	'email',p.email,
    	'telefono',p.telefono,
    	'celular',p.celular,
    	'creadoEn',p.creado_en,
    	'actualizadoEn',p.actualizado_en,
    	'eliminadoEn',p.eliminado_en
    ) AS proveedor
FROM compra c
LEFT JOIN public.sucursal s on c.sucursal_id = s.id
LEFT JOIN public.proveedor p on p.id = c.proveedor_id
LEFT JOIN public.usuario_admin ua on ua.id = c.usuario_admin_id
`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += ` ORDER BY c.id`
	rows, err := c.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al obtener lista de compra:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.CompraInfo, 0)
	for rows.Next() {
		var item domain.CompraInfo
		err := rows.Scan(&item.Id, &item.CodigoCompra, &item.Estado, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn, &item.Usuario, &item.Sucursal, &item.Proveedor)
		if err != nil {
			log.Println("Error scanning rows:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (c CompraRepository) ObtenerCompraById(ctx context.Context, id *int) (*domain.Compra, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `
SELECT 
    c.id,
    c.codigo_compra,
    c.estado,
    c.creado_en,
    c.actualizado_en,
    c.eliminado_en,
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
    json_build_object(
       'id',p.id,
       'nombre',p.nombre,
       'estado',p.estado,
       'email',p.email,
       'telefono',p.telefono,
       'celular',p.celular,
       'creadoEn',p.creado_en,
       'actualizadoEn',p.actualizado_en,
       'eliminadoEn',p.eliminado_en
    ) AS proveedor,
    COALESCE(
       json_agg(
          CASE 
             WHEN dc.id IS NULL THEN NULL
             ELSE json_build_object(
                'cantidad',dc.cantidad,
                'precioCompra',dc.precio_compra,
                'precioVenta',dc.precio_venta,
                'producto', json_build_object(
                	'id',pr.id,
                   	'nombre',pr.nombre,
                	'estado',pr.estado,
                    'urlFoto',($1::text || pr.id::text || '/' || pr.foto),
                    'creadoEn',pr.creado_en,
                	'actualizadoEn',pr.actualizado_en,
                	'eliminadoEn',pr.eliminado_en
                ),
                'ubicacion', json_build_object(
                	'id',ub.id,
                   	'nombre',ub.nombre,
                	'estado',ub.estado,
                    'esVendible',ub.es_vendible,
                	'prioridadVenta',ub.prioridad_venta
                )
             )
          END
       ), 
       '[]'
    ) AS detalles
FROM compra c
LEFT JOIN public.sucursal s on c.sucursal_id = s.id
LEFT JOIN public.proveedor p on p.id = c.proveedor_id
LEFT JOIN public.usuario_admin ua on ua.id = c.usuario_admin_id
LEFT JOIN public.detalle_compra dc on c.id = dc.compra_id
LEFT JOIN public.producto pr on dc.producto_id = pr.id
LEFT JOIN public.ubicacion ub on dc.ubicacion_id = ub.id
WHERE c.id=$2 
GROUP BY c.id,p.id,ua.id,s.id`

	var item domain.Compra
	err := c.pool.QueryRow(ctx, query, fullHostname, *id).
		Scan(&item.Id, &item.CodigoCompra, &item.Estado, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn, &item.Usuario, &item.Sucursal, &item.Proveedor, &item.Detalles)
	if err != nil {
		log.Println("Error al obtener compra:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Compra no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func (c CompraRepository) RegistrarOrdenCompra(ctx context.Context, request *domain.CompraRequest) (*int, error) {
	// 1. Iniciar transacción
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(ctx); err != nil {
				log.Println("Error durante rollback:", err)
			}
		}
	}()

	// 2. Insertar la cabecera (Sin cambios)
	var compraId int
	query := `INSERT INTO compra(usuario_admin_id, proveedor_id, sucursal_id, estado, codigo_compra) 
              VALUES($1, $2, $3, $4, nextval('seq_codigo_compra')) RETURNING id`

	err = tx.QueryRow(ctx, query, request.UsuarioId, request.ProveedorId, request.SucursalId, "Pendiente").Scan(&compraId)
	if err != nil {
		// ... (Mismo manejo de errores de FK que antes) ...
		log.Println("Error al insertar cabecera:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if len(request.Detalles) == 0 {
		return nil, datatype.NewBadRequestError("Debe haber un detalle de compra")
	}

	// 3. Preparación de datos (Recolección de IDs)
	mapUbicaciones := make(map[int]struct{})
	mapProductos := make(map[int]struct{})
	var idsUbicaciones []int
	var idsProductos []int

	for _, detalle := range request.Detalles {
		if _, ok := mapUbicaciones[detalle.UbicacionId]; !ok {
			mapUbicaciones[detalle.UbicacionId] = struct{}{}
			idsUbicaciones = append(idsUbicaciones, detalle.UbicacionId)
		}
		if _, ok := mapProductos[detalle.ProductoId]; !ok {
			mapProductos[detalle.ProductoId] = struct{}{}
			idsProductos = append(idsProductos, detalle.ProductoId)
		}
	}

	// --- VALIDACIÓN UBICACIONES (Sin cambios) ---
	var countUbicaciones int
	queryValidarUbi := `SELECT COUNT(id) FROM ubicacion WHERE sucursal_id = $1 AND id = ANY($2)`
	err = tx.QueryRow(ctx, queryValidarUbi, request.SucursalId, idsUbicaciones).Scan(&countUbicaciones)
	if err != nil || countUbicaciones != len(idsUbicaciones) {
		return nil, datatype.NewBadRequestError("Una ubicación no es válida o no pertenece a la sucursal.")
	}

	// --- VALIDACIÓN Y OBTENCIÓN DE PRECIOS ACTUALES (AQUÍ USAMOS TU QUERY) ---
	// Usamos tu consulta para validar existencia Y obtener precio al mismo tiempo
	queryPrecios := `
       SELECT p.id, ps.precio 
       FROM producto p
       LEFT JOIN public.producto_sucursal ps on p.id = ps.producto_id
       WHERE p.id = ANY($1) 
         AND ps.sucursal_id = $2 
         AND p.es_inventariable IS TRUE`

	rowsPrecios, err := tx.Query(ctx, queryPrecios, idsProductos, request.SucursalId)
	if err != nil {
		log.Println("Error al consultar precios/validar productos:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rowsPrecios.Close()

	mapPreciosActuales := make(map[int]float64)
	countProductosEncontrados := 0

	for rowsPrecios.Next() {
		var pId int
		var pPrecio float64
		if err := rowsPrecios.Scan(&pId, &pPrecio); err != nil {
			log.Println("Error escaneando precios:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		mapPreciosActuales[pId] = pPrecio
		countProductosEncontrados++
	}

	// Validación estricta: Si no encontramos todos los IDs, es que alguno no existe en la sucursal o no es inventariable
	if countProductosEncontrados != len(idsProductos) {
		return nil, datatype.NewBadRequestError("Uno o más productos no están asignados a esta sucursal o no son inventariables.")
	}

	// 4. Inserción Masiva con Lógica de Precio
	var rows [][]interface{}
	for _, detalle := range request.Detalles {

		precioVentaFinal := detalle.PrecioVenta

		// LA LÓGICA CLAVE: Si viene 0, ponemos el precio que acabamos de consultar
		if precioVentaFinal == 0 {
			if precioActual, ok := mapPreciosActuales[detalle.ProductoId]; ok {
				precioVentaFinal = precioActual
			}
		}

		rows = append(rows, []interface{}{
			compraId,
			detalle.ProductoId,
			detalle.UbicacionId,
			detalle.Cantidad,
			detalle.PrecioCompra,
			precioVentaFinal, // Guardamos el precio correcto
		})
	}

	columnas := []string{"compra_id", "producto_id", "ubicacion_id", "cantidad", "precio_compra", "precio_venta"}

	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"detalle_compra"},
		columnas,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		log.Println("Error copyFrom:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if int(copyCount) != len(request.Detalles) {
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 5. Commit
	err = tx.Commit(ctx)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return &compraId, nil
}

func (c CompraRepository) ModificarOrdenCompra(ctx context.Context, id *int, request *domain.CompraRequest) error {
	// 1. Iniciar transacción
	tx, err := c.pool.Begin(ctx)
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

	// 2. Verificar estado (Bloqueo pesimista)
	var estadoActual string
	queryEstado := `SELECT estado FROM compra WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryEstado, *id).Scan(&estadoActual)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return datatype.NewNotFoundError("Compra no encontrada")
		}
		log.Println("Error al obtener estado:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if estadoActual != "Pendiente" {
		return datatype.NewBadRequestError(fmt.Sprintf("No se puede modificar una compra en estado '%s'", estadoActual))
	}

	if len(request.Detalles) == 0 {
		return datatype.NewBadRequestError("Debe haber un detalle de compra")
	}

	// 3. Recolección de IDs
	mapUbicaciones := make(map[int]struct{})
	mapProductos := make(map[int]struct{})
	var idsUbicaciones []int
	var idsProductos []int

	for _, detalle := range request.Detalles {
		if _, ok := mapUbicaciones[detalle.UbicacionId]; !ok {
			mapUbicaciones[detalle.UbicacionId] = struct{}{}
			idsUbicaciones = append(idsUbicaciones, detalle.UbicacionId)
		}
		if _, ok := mapProductos[detalle.ProductoId]; !ok {
			mapProductos[detalle.ProductoId] = struct{}{}
			idsProductos = append(idsProductos, detalle.ProductoId)
		}
	}

	// 4. Validar Ubicaciones
	var countUbicaciones int
	queryValidarUbi := `SELECT COUNT(id) FROM ubicacion WHERE sucursal_id = $1 AND id = ANY($2)`
	err = tx.QueryRow(ctx, queryValidarUbi, request.SucursalId, idsUbicaciones).Scan(&countUbicaciones)
	if err != nil || countUbicaciones != len(idsUbicaciones) {
		return datatype.NewBadRequestError("Una ubicación no pertenece a la sucursal de la compra.")
	}

	// 5. Validar Productos y Obtener Precios Actuales
	// (Lógica agregada para mantener precios si vienen en 0)
	queryPrecios := `
       SELECT p.id, ps.precio 
       FROM producto p
       LEFT JOIN public.producto_sucursal ps on p.id = ps.producto_id
       WHERE p.id = ANY($1) 
         AND ps.sucursal_id = $2 
         AND p.es_inventariable IS TRUE`

	rowsPrecios, err := tx.Query(ctx, queryPrecios, idsProductos, request.SucursalId)
	if err != nil {
		log.Println("Error al consultar precios:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	defer rowsPrecios.Close()

	mapPreciosActuales := make(map[int]float64)
	countProductosEncontrados := 0

	for rowsPrecios.Next() {
		var pId int
		var pPrecio float64
		if err := rowsPrecios.Scan(&pId, &pPrecio); err != nil {
			log.Println("Error escaneando precios:", err)
			return datatype.NewInternalServerErrorGeneric()
		}
		mapPreciosActuales[pId] = pPrecio
		countProductosEncontrados++
	}

	if countProductosEncontrados != len(idsProductos) {
		return datatype.NewBadRequestError("Uno o más productos no existen en esta sucursal o no son inventariables.")
	}

	// 6. Actualizar Cabecera
	queryUpdate := `UPDATE compra SET proveedor_id = $1, sucursal_id = $2, actualizado_en = CURRENT_TIMESTAMP WHERE id = $3`
	ct, err := tx.Exec(ctx, queryUpdate, request.ProveedorId, request.SucursalId, *id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			// Manejo de errores de FK
			var errorMsg string
			switch pgErr.ConstraintName {
			case "compra_proveedor_id_fkey":
				errorMsg = "El proveedor seleccionado no existe."
			case "compra_sucursal_id_fkey":
				errorMsg = "La sucursal seleccionada no existe."
			default:
				errorMsg = "Error de referencia en cabecera."
			}
			return datatype.NewBadRequestError(errorMsg)
		}
		log.Println("Error update cabecera:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Compra no encontrada")
	}

	// 7. Eliminar detalles antiguos
	queryDelete := `DELETE FROM detalle_compra WHERE compra_id = $1`
	_, err = tx.Exec(ctx, queryDelete, *id)
	if err != nil {
		log.Println("Error delete detalles:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// 8. Insertar nuevos detalles (Con lógica de precio 0)
	var rows [][]interface{}
	for _, detalle := range request.Detalles {

		precioVentaFinal := detalle.PrecioVenta

		// Si envían 0, mantenemos el precio actual de la DB
		if precioVentaFinal == 0 {
			if precioActual, ok := mapPreciosActuales[detalle.ProductoId]; ok {
				precioVentaFinal = precioActual
			}
		}

		rows = append(rows, []interface{}{
			*id, // ID de la compra existente
			detalle.ProductoId,
			detalle.UbicacionId,
			detalle.Cantidad,
			detalle.PrecioCompra,
			precioVentaFinal, // Usamos el precio calculado
		})
	}

	columnas := []string{"compra_id", "producto_id", "ubicacion_id", "cantidad", "precio_compra", "precio_venta"}

	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"detalle_compra"},
		columnas,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		log.Println("Error CopyFrom:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if int(copyCount) != len(request.Detalles) {
		return datatype.NewInternalServerErrorGeneric()
	}

	// 9. Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error commit:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return nil
}

func (c CompraRepository) ConfirmarRecepcionCompra(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := c.pool.Begin(ctx)
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

	// Verificar estado de la compra (y bloquear la fila)
	var estadoActual string
	queryEstado := `SELECT estado FROM compra WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryEstado, *id).Scan(&estadoActual)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return datatype.NewNotFoundError("Compra no encontrada")
		}
		log.Println("Error al obtener estado de la compra:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if estadoActual != "Pendiente" {
		log.Printf("Intento de confirmar compra %d con estado %s\n", *id, estadoActual)
		return datatype.NewBadRequestError("Esta compra no está pendiente, no se puede confirmar.")
	}

	// Actualizar inventario
	queryUpsertInventario := `
       INSERT INTO inventario (producto_id, ubicacion_id, stock)
       (
          SELECT 
             producto_id, 
             ubicacion_id, 
             cantidad 
          FROM detalle_compra 
          WHERE compra_id = $1
       )
       ON CONFLICT (producto_id, ubicacion_id) 
       DO UPDATE SET
          stock = inventario.stock + EXCLUDED.stock;
    `
	_, err = tx.Exec(ctx, queryUpsertInventario, *id)
	if err != nil {
		log.Println("Error al actualizar el inventario (UPSERT):", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// Actualizar precios en la tabla 'producto_sucursal' ---
	// Actualizar precios en la tabla 'producto_sucursal'
	queryUpdatePrecios := `
        UPDATE producto_sucursal AS ps
        SET 
            precio = dc.precio_venta
        FROM (
            SELECT DISTINCT ON (producto_id)
                producto_id,
                precio_venta
            FROM detalle_compra
            WHERE compra_id = $1
        ) AS dc
        WHERE ps.producto_id = dc.producto_id
          AND ps.sucursal_id = (SELECT sucursal_id FROM compra WHERE id = $1);
    `

	cmdTag, err := tx.Exec(ctx, queryUpdatePrecios, *id)
	if err != nil {
		log.Println("Error al actualizar precios de productos:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// Opcional: Loguear cuántos precios se actualizaron
	log.Printf("Se actualizaron precios de %d productos en la sucursal", cmdTag.RowsAffected())

	// Actualizar estado de compra
	queryUpdateCompra := `UPDATE compra SET estado = 'Completado', actualizado_en = CURRENT_TIMESTAMP WHERE id = $1`

	ct, err := tx.Exec(ctx, queryUpdateCompra, *id)
	if err != nil {
		log.Println("Error al actualizar estado de la compra:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		log.Println("Error, la compra no se actualizó (RowsAffected 0)")
		return datatype.NewInternalServerErrorGeneric()
	}

	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return nil
}

func NewCompraRepository(pool *pgxpool.Pool) *CompraRepository {
	return &CompraRepository{pool: pool}
}

var _ port.CompraRepository = (*CompraRepository)(nil)
