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
                    'precio',pr.precio,
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
GROUP BY c.id,p.id,ua.id,s.id
	`
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
	// Iniciar transacción
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

	// Insertar la cabecera de la compra
	var compraId int
	query := `INSERT INTO compra(usuario_admin_id, proveedor_id, sucursal_id, estado,codigo_compra) VALUES($1, $2, $3, $4,nextval('seq_codigo_compra')) RETURNING id`

	err = tx.QueryRow(ctx, query, request.UsuarioId, request.ProveedorId, request.SucursalId, "Pendiente").Scan(&compraId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			// Error 23503 = Foreign Key Violation
			log.Println("Error de FK al insertar cabecera:", pgErr.Message)
			var errorMsg string

			switch pgErr.ConstraintName {
			case "compra_proveedor_id_fkey":
				errorMsg = "El proveedor seleccionado no existe."
			case "compra_sucursal_id_fkey":
				errorMsg = "La sucursal seleccionada no existe."
			case "compra_usuario_admin_id_fkey":
				errorMsg = "El usuario administrador seleccionado no existe."
			default:
				// Captura cualquier otra violación de FK
				errorMsg = "Error de referencia: uno de los campos es inválido."
			}
			return nil, datatype.NewBadRequestError(errorMsg)
		}
		// Si no es un 23503, es un error interno.
		log.Println("Error al insertar cabecera de compra:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// No permitir una compra sin detalles.
	if len(request.Detalles) == 0 {
		return nil, datatype.NewBadRequestError("Debe haber un detalle de compra")
	}

	// Validación en lote
	// Recolectar Ids únicos de productos y ubicaciones
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

	// Validar todas las ubicaciones de una sola vez
	var countUbicaciones int
	queryValidarUbi := `SELECT COUNT(id) FROM ubicacion WHERE sucursal_id = $1 AND id = ANY($2)`

	err = tx.QueryRow(ctx, queryValidarUbi, request.SucursalId, idsUbicaciones).Scan(&countUbicaciones)
	if err != nil {
		log.Println("Error al validar ubicaciones en lote:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if countUbicaciones != len(idsUbicaciones) {
		log.Println("Error de validación: una o más ubicaciones no son válidas o no pertenecen a la sucursal.")
		return nil, datatype.NewBadRequestError("Una ubicación no pertenece a la sucursal de la compra.")
	}

	// Validar todos los productos de una sola vez
	var countProductos int
	queryValidarProd := `SELECT COUNT(id) FROM producto WHERE id = ANY($1)`

	err = tx.QueryRow(ctx, queryValidarProd, idsProductos).Scan(&countProductos)
	if err != nil {
		log.Println("Error al validar productos en lote:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if countProductos != len(idsProductos) {
		log.Println("Error de validación: uno o más productos no existen.")
		return nil, datatype.NewBadRequestError("Uno de los productos no existe.")
	}

	// Bucle de Inserción (Ahora es "tonto" y rápido)
	// Usamos pgx.CopyFrom para una inserción masiva, que es aún más rápida
	// que un bucle de 'tx.Exec'

	// Preparamos los datos para la copia masiva
	var rows [][]interface{}
	for _, detalle := range request.Detalles {
		rows = append(rows, []interface{}{
			compraId,
			detalle.ProductoId,
			detalle.UbicacionId,
			detalle.Cantidad,
			detalle.PrecioCompra,
			detalle.PrecioVenta,
		})
	}

	// Ejecutamos la copia masiva
	columnas := []string{"compra_id", "producto_id", "ubicacion_id", "cantidad", "precio_compra", "precio_venta"}

	copyCount, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"detalle_compra"},
		columnas,
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		log.Println("Error durante la inserción masiva de detalles:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if int(copyCount) != len(request.Detalles) {
		log.Printf("Error de conteo en inserción masiva. Esperado: %d, Insertado: %d\n", len(request.Detalles), copyCount)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return &compraId, nil
}

func (c CompraRepository) ModificarOrdenCompra(ctx context.Context, id *int, request *domain.CompraRequest) error {
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

	// Verificar estado
	// No podemos modificar una compra que ya fue recibida o cancelada.
	// Usamos 'FOR UPDATE' para bloquear la fila y evitar que otro proceso
	// la confirme mientras la estamos modificando.
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
		log.Printf("Intento de modificar compra %d con estado no editable %s\n", *id, estadoActual)
		return datatype.NewBadRequestError(fmt.Sprintf("No se puede modificar una compra en estado '%s'", estadoActual))
	}

	// No permitir una compra sin detalles.
	if len(request.Detalles) == 0 {
		return datatype.NewBadRequestError("Debe haber un detalle de compra")
	}

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

	// Validar ubicaciones
	var countUbicaciones int
	queryValidarUbi := `SELECT COUNT(id) FROM ubicacion WHERE sucursal_id = $1 AND id = ANY($2)`
	err = tx.QueryRow(ctx, queryValidarUbi, request.SucursalId, idsUbicaciones).Scan(&countUbicaciones)
	if err != nil {
		log.Println("Error al validar ubicaciones en lote:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if countUbicaciones != len(idsUbicaciones) {
		return datatype.NewBadRequestError("Una ubicación no pertenece a la sucursal de la compra.")
	}

	// Validar productos
	var countProductos int
	queryValidarProd := `SELECT COUNT(id) FROM producto WHERE id = ANY($1)`
	err = tx.QueryRow(ctx, queryValidarProd, idsProductos).Scan(&countProductos)
	if err != nil {
		log.Println("Error al validar productos en lote:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if countProductos != len(idsProductos) {
		return datatype.NewBadRequestError("Uno de los productos no existe.")
	}

	// Actualizar cabecera
	queryUpdate := `UPDATE compra SET proveedor_id = $1, sucursal_id = $2, actualizado_en = CURRENT_TIMESTAMP WHERE id = $3`
	ct, err := tx.Exec(ctx, queryUpdate, request.ProveedorId, request.SucursalId, *id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			// Error 23503 = Foreign Key Violation
			log.Println("Error de FK al insertar cabecera:", pgErr.Message)
			var errorMsg string

			switch pgErr.ConstraintName {
			case "compra_proveedor_id_fkey":
				errorMsg = "El proveedor seleccionado no existe."
			case "compra_sucursal_id_fkey":
				errorMsg = "La sucursal seleccionada no existe."
			default:
				// Captura cualquier otra violación de FK
				errorMsg = "Error de referencia: uno de los campos es inválido."
			}
			return datatype.NewBadRequestError(errorMsg)
		}
		// Si no es un 23503, es un error interno.
		log.Println("Error al insertar cabecera de compra:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		// Esto no debería pasar gracias al 'SELECT FOR UPDATE' anterior, pero es una buena defensa
		return datatype.NewNotFoundError("Compra no encontrada")
	}

	// Eliminar detalles antiguos
	queryDelete := `DELETE FROM detalle_compra WHERE compra_id = $1`
	_, err = tx.Exec(ctx, queryDelete, *id)
	if err != nil {
		log.Println("Error al eliminar detalles antiguos:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// Insertar nuevos detalles (Usando CopyFrom)

	var rows [][]interface{}
	for _, detalle := range request.Detalles {
		rows = append(rows, []interface{}{
			*id, // ID de la compra existente
			detalle.ProductoId,
			detalle.UbicacionId,
			detalle.Cantidad,
			detalle.PrecioCompra,
			detalle.PrecioVenta,
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
		log.Println("Error durante la inserción masiva de detalles:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if int(copyCount) != len(request.Detalles) {
		log.Printf("Error de conteo en inserción masiva. Esperado: %d, Insertado: %d\n", len(request.Detalles), copyCount)
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

	// Actualizar precios en la tabla 'producto' ---
	// Esta consulta actualiza el 'precio_venta' de todos los productos involucrados en esta compra.
	queryUpdatePrecios := `
        UPDATE producto AS p
        SET 
            precio = dc.precio_venta
        FROM (
            -- Obtenemos el precio de cada producto en esta compra
            SELECT DISTINCT ON (producto_id)
                producto_id,
                precio_venta,
                precio_compra
            FROM detalle_compra
            WHERE compra_id = $1
        ) AS dc
        WHERE p.id = dc.producto_id;
    `
	_, err = tx.Exec(ctx, queryUpdatePrecios, *id)
	if err != nil {
		log.Println("Error al actualizar precios de productos:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

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
