package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductoRepository struct {
	pool *pgxpool.Pool
}

func (p ProductoRepository) ObtenerProductoSucursalById(ctx context.Context, id *int) (*domain.ProductoSucursalInfo, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")

	query := `
	SELECT 
		ps.id,
		ps.precio,
		ps.estado,
	json_build_object(
		'id', p.id,
		'nombre', p.nombre,
		'estado', p.estado,
		'urlFoto', ($1::text || p.id::text || '/' || p.foto),
		'esInventariable', p.es_inventariable,
		'creadoEn', p.creado_en,
		'actualizadoEn', p.actualizado_en,
		'eliminadoEn', p.eliminado_en
	) as producto_info
	FROM producto_sucursal ps
	INNER JOIN producto p ON ps.producto_id = p.id
	WHERE ps.id = $2
    `
	var item domain.ProductoSucursalInfo
	// Escaneo directo a la estructura anidada
	err := p.pool.QueryRow(ctx, query, fullHostname, *id).Scan(
		&item.Id,
		&item.Precio,
		&item.Estado,
		&item.Producto,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("El producto en sucursal no existe")
		}
		log.Println("Error al obtener ProductoSucursal:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	return &item, nil
}

func (p ProductoRepository) ActualizarProductoSucursal(ctx context.Context, id *int, req *domain.ProductoSucursalUpdateRequest) error {
	// Validaciones básicas
	if req.Precio < 0 {
		return datatype.NewBadRequestError("El precio no puede ser negativo")
	}
	if req.Estado == "" {
		return datatype.NewBadRequestError("El estado es obligatorio")
	}

	query := `
        UPDATE producto_sucursal
        SET 
            precio = $1,
            estado = $2
        WHERE id = $3
    `

	cmdTag, err := p.pool.Exec(ctx, query, req.Precio, req.Estado, *id)
	if err != nil {
		log.Println("Error al actualizar ProductoSucursal:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if cmdTag.RowsAffected() == 0 {
		return datatype.NewNotFoundError("No se encontró el registro para actualizar")
	}

	return nil
}

func (p ProductoRepository) ListarProductosPorSucursal(ctx context.Context, filtros map[string]string) (*[]domain.ProductoSucursalInfo, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")

	// Validamos Sucursal
	sucursalIdStr := filtros["sucursalId"]
	if sucursalIdStr == "" {
		return nil, datatype.NewBadRequestError("El parámetro sucursalId es obligatorio para listar precios.")
	}

	sucursalId, err := strconv.Atoi(sucursalIdStr)
	if err != nil {
		return nil, datatype.NewBadRequestError("sucursalId inválido")
	}

	var filters []string
	// $1 es hostname, $2 es sucursalId
	var args = []interface{}{fullHostname, sucursalId}
	var i = 3

	// Filtro Categoría
	if val, ok := filtros["categoriaId"]; ok {
		if val == "null" {
			filters = append(filters, "p.categoria_id IS NULL")
		} else {
			id, err := strconv.Atoi(val)
			if err == nil {
				filters = append(filters, fmt.Sprintf("p.categoria_id = $%d", i))
				args = append(args, id)
				i++
			} else {
				return nil, datatype.NewBadRequestError("El valor de categoriaId no es válido")
			}
		}
	}

	// Filtro Buscador (Opcional, por si buscan por nombre)
	if busqueda, ok := filtros["q"]; ok && busqueda != "" {
		filters = append(filters, fmt.Sprintf("p.nombre ILIKE $%d", i))
		args = append(args, "%"+busqueda+"%")
		i++
	}

	query := `
	SELECT 
		ps.id,
		ps.estado,
		ps.precio,
        json_build_object(
            'id', p.id,
            'nombre', p.nombre,
            'estado', p.estado,
            'urlFoto', ($1::text || p.id::text || '/' || p.foto),
            'esInventariable', p.es_inventariable,
            'creadoEn', p.creado_en,
            'actualizadoEn', p.actualizado_en,
            'eliminadoEn', p.eliminado_en
        ) as producto_info
	FROM producto_sucursal ps
	INNER JOIN producto p ON ps.producto_id = p.id
	WHERE ps.sucursal_id = $2
    `

	if len(filters) > 0 {
		query += " AND " + strings.Join(filters, " AND ")
	}

	query += ` ORDER BY p.nombre ASC` // Orden alfabético suele ser mejor para ventas

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error query listar productos sucursal:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	list := make([]domain.ProductoSucursalInfo, 0)
	for rows.Next() {
		var item domain.ProductoSucursalInfo
		// Asegúrate de que tu struct ProductoInfo tenga los tags correctos para recibir estos datos
		err = rows.Scan(
			&item.Id,
			&item.Estado,
			&item.Precio,
			&item.Producto,
		)
		if err != nil {
			log.Println("Error al escanear producto:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (p ProductoRepository) HabilitarProductoById(ctx context.Context, productoId *int) error {
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	query := `UPDATE producto SET estado='Activo' WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *productoId)
	if err != nil {
		log.Println("Error al actualizar producto:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Producto no encontrado")
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

func (p ProductoRepository) DeshabilitarProductoById(ctx context.Context, productoId *int) error {
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	query := `UPDATE producto SET estado='Inactivo' WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *productoId)
	if err != nil {
		log.Println("Error al actualizar producto:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Producto no encontrado")
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

func (p ProductoRepository) ListarProductosMasVendidos(ctx context.Context, filtros map[string]string) (*[]domain.ProductoStat, error) {

	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")

	// 1. PREPARAR VARIABLES (Punteros para permitir nil/NULL)
	// De esta forma, pgx enviará NULL a la base de datos si no hay filtro
	var categoriaId *int
	if val, ok := filtros["categoriaId"]; ok && val != "" && val != "null" {
		id, _ := strconv.Atoi(val)
		categoriaId = &id
	}

	var sucursalId *int
	if val, ok := filtros["sucursalId"]; ok && val != "" {
		id, _ := strconv.Atoi(val)
		sucursalId = &id
	}

	var fechaInicio *string
	if val, ok := filtros["fechaInicio"]; ok && val != "" {
		fechaInicio = &val
	}

	var fechaFin *string
	if val, ok := filtros["fechaFin"]; ok && val != "" {
		fechaFin = &val
	}

	// 2. QUERY ESTÁTICA
	query := `
    SELECT 
        COALESCE(v_stats.total_dinero, 0) as total_ventas,
        COALESCE(v_stats.total_cantidad, 0) as cantidad_ventas,
        COALESCE(c_stats.total_dinero, 0) as total_compras,
        COALESCE(c_stats.total_cantidad, 0) as cantidad_compras,

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

        -- Ventas Diarias
        COALESCE((
            SELECT json_agg(vd) FROM (
                SELECT 
                    TO_CHAR(v.creado_en, 'YYYY-MM-DD') as fecha, 
                    SUM(dv.cantidad * dv.precio_venta) as total 
                FROM detalle_venta dv
                JOIN venta v ON dv.venta_id = v.id
                WHERE dv.producto_id = p.id 
                  AND v.estado = 'Completado'
                  -- LOGICA SQL: Si el parametro $3 es NULL, ignora el filtro
                  AND ($3::int IS NULL OR v.sucursal_id = $3)
                  AND ($4::date IS NULL OR v.creado_en >= $4::date)
                  AND ($5::date IS NULL OR v.creado_en <= $5::date)
                GROUP BY TO_CHAR(v.creado_en, 'YYYY-MM-DD')
                ORDER BY fecha DESC
                LIMIT 30 
            ) vd
        ), '[]') as ventas_diarias,

        -- Compras Diarias
        COALESCE((
            SELECT json_agg(cd) FROM (
                SELECT 
                    TO_CHAR(c.creado_en, 'YYYY-MM-DD') as fecha, 
                    SUM(dc.cantidad * dc.precio_compra) as total
                FROM detalle_compra dc
                JOIN compra c ON dc.compra_id = c.id
                WHERE dc.producto_id = p.id 
                  AND c.estado = 'Completado'
                  AND ($3::int IS NULL OR c.sucursal_id = $3)
                  AND ($4::date IS NULL OR c.creado_en >= $4::date)
                  AND ($5::date IS NULL OR c.creado_en <= $5::date)
                GROUP BY TO_CHAR(c.creado_en, 'YYYY-MM-DD')
                ORDER BY fecha DESC
                LIMIT 30
            ) cd
        ), '[]') as compras_diarias

    FROM producto p

    -- Stats Totales Venta
    LEFT JOIN LATERAL (
        SELECT 
            SUM(dv.cantidad * dv.precio_venta) as total_dinero,
            SUM(dv.cantidad) as total_cantidad
        FROM detalle_venta dv
        JOIN venta v ON dv.venta_id = v.id
        WHERE dv.producto_id = p.id 
          AND v.estado = 'Completado'
          AND ($3::int IS NULL OR v.sucursal_id = $3)
          AND ($4::date IS NULL OR v.creado_en >= $4::date)
          AND ($5::date IS NULL OR v.creado_en <= $5::date)
    ) v_stats ON true

    -- Stats Totales Compra
    LEFT JOIN LATERAL (
        SELECT 
            SUM(dc.cantidad * dc.precio_compra) as total_dinero,
            SUM(dc.cantidad) as total_cantidad
        FROM detalle_compra dc
        JOIN compra c ON dc.compra_id = c.id
        WHERE dc.producto_id = p.id 
          AND c.estado = 'Completado'
          AND ($3::int IS NULL OR c.sucursal_id = $3)
          AND ($4::date IS NULL OR c.creado_en >= $4::date)
          AND ($5::date IS NULL OR c.creado_en <= $5::date)
    ) c_stats ON true

    WHERE ($2::int IS NULL OR p.categoria_id = $2) -- Filtro Categoria Principal

    ORDER BY v_stats.total_cantidad DESC NULLS LAST, p.nombre ASC
    `

	// 3. EJECUTAR
	// Pasamos los punteros. Si son nil, PostgreSQL recibe NULL y la condición "IS NULL" se vuelve verdadera.
	rows, err := p.pool.Query(ctx, query, fullHostname, categoriaId, sucursalId, fechaInicio, fechaFin)

	if err != nil {
		log.Println("Error query stats:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	list := make([]domain.ProductoStat, 0)
	for rows.Next() {
		var item domain.ProductoStat
		err = rows.Scan(
			&item.TotalVentas,
			&item.CantidadVentas,
			&item.TotalCompras,
			&item.CantidadCompras,
			&item.Producto,
			&item.VentasDiarias,
			&item.ComprasDiarias,
		)
		if err != nil {
			log.Println("Error al escanear estadísticas:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}

	return &list, nil
}

func (p ProductoRepository) RegistrarProducto(ctx context.Context, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) (*int, error) {
	nombreArchivo := strings.ToLower(fileHeader.Filename)
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	var productoId int
	query := `INSERT INTO producto(nombre, estado, foto,es_inventariable,categoria_id) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, request.Estado, nombreArchivo, request.EsInventariable, request.CategoriaId).Scan(&productoId)
	if err != nil {
		log.Println("Error al actualizar producto:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_producto" {
					return nil, datatype.NewConflictError("Ya existe un producto con ese nombre")
				}
				return nil, datatype.NewInternalServerErrorGeneric()
			}
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	query = `INSERT INTO producto_sucursal (producto_id, sucursal_id, precio, estado)SELECT $1, id, $2, 'Activo' FROM sucursal`
	_, err = tx.Exec(ctx, query, productoId, request.Precio)
	if err != nil {
		log.Println("Error asignando sucursales:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	routeDir := fmt.Sprintf("./public/uploads/productos/%d", productoId)
	defer func() {
		if !committed {
			_ = util.File.DeleteAllFiles(routeDir)
		}
	}()
	// Crear directorio
	err = util.File.MakeDir(routeDir)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Abrir archivo
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error al abrir archivo:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Guardar archivo
	err = util.File.SaveFile(routeDir, nombreArchivo, file)
	if err != nil {
		log.Println("Error al guardar imagen:", err)
		return nil, datatype.NewInternalServerError("Error al guardar imagen")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return &productoId, nil
}

func (p ProductoRepository) ModificarProductoById(ctx context.Context, productoId *int, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) error {
	nombreArchivo := strings.ToLower(fileHeader.Filename)
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	query := `UPDATE producto SET nombre=$1,foto=$2,estado=$3,actualizado_en=now(), categoria_id=$4 WHERE id=$5`
	ct, err := tx.Exec(ctx, query, request.Nombre, nombreArchivo, request.Estado, request.Precio, request.CategoriaId, *productoId)
	if err != nil {
		log.Println("Error al actualizar producto:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_producto" {
					return datatype.NewConflictError("Ya existe un producto con ese nombre")
				}
				// Otra violación única
				return datatype.NewInternalServerErrorGeneric()
			}
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Producto no encontrado")
	}

	// Respaldar archivos existentes
	route := fmt.Sprintf("./public/uploads/productos/%d", *productoId)
	backupFiles, err := util.File.BackupFiles(route)
	if err != nil {
		log.Println(err)
		return datatype.NewInternalServerErrorGeneric()
	}
	defer func() {
		if !committed {
			_ = util.File.RestoreFiles(backupFiles, route)
		}
	}()
	// Eliminar archivos
	err = util.File.DeleteAllFiles(route)
	if err != nil {
		log.Println("Error al eliminar archivos:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// Crear directorio
	routeDir := fmt.Sprintf("./public/uploads/productos/%d", *productoId)
	err = util.File.MakeDir(routeDir)
	if err != nil {
		log.Println("Error al crear directorio:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	// Abrir archivo
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error al abrir archivo")
		return datatype.NewInternalServerErrorGeneric()
	}

	// Guardar archivo
	err = util.File.SaveFile(routeDir, nombreArchivo, file)
	if err != nil {
		log.Println("Error al guardar imagen:", err)
		return datatype.NewInternalServerError("Error al guardar imagen")
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

func (p ProductoRepository) ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.ProductoInfo, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	var filters []string
	var args = []interface{}{fullHostname}
	var i = 2
	if val, ok := filtros["categoriaId"]; ok {
		// Caso A: Quieres buscar productos SIN categoría (NULL)
		if val == "null" {
			filters = append(filters, "p.categoria_id IS NULL")
		} else {
			// Caso B: Quieres buscar una categoría específica (ID entero)
			// Convertimos a entero para seguridad (opcional pero recomendado)
			id, err := strconv.Atoi(val)
			if err == nil {
				filters = append(filters, fmt.Sprintf("p.categoria_id = $%d", i))
				args = append(args, id)
				i++
			} else {
				return nil, datatype.NewBadRequestError("El valor de categoriaId no es válido")
			}
		}
	}

	query := `
SELECT 
    p.id,
    p.nombre,
    p.estado,
    ($1::text || p.id::text || '/' || p.foto) AS url_foto,
    p.es_inventariable,
    p.creado_en,
    p.actualizado_en,
    p.eliminado_en 
FROM producto p
`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += ` ORDER BY p.id`
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.ProductoInfo, 0)
	for rows.Next() {
		var item domain.ProductoInfo
		err = rows.Scan(&item.Id, &item.Nombre, &item.Estado, &item.UrlFoto, &item.EsInventariable, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn)
		if err != nil {
			log.Println("Error al escanear producto:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (p ProductoRepository) ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `
SELECT 
	p.id,
	p.nombre,
	p.estado,
	($1::text || p.id::text || '/' || p.foto) AS url_foto,
	p.es_inventariable,
	p.creado_en,
	p.actualizado_en,
	p.eliminado_en,
	CASE 
		WHEN c.id IS NOT NULL THEN json_build_object(
			'id', c.id,
			'nombre', c.nombre,
			'descripcion', c.descripcion,
			'estado', c.estado
		)
		ELSE NULL
	END AS categoria
FROM producto p 
LEFT JOIN categoria_producto c ON p.categoria_id = c.id
WHERE p.id = $2 
LIMIT 1`
	var item domain.Producto
	err := p.pool.QueryRow(ctx, query, fullHostname, *productoId).
		Scan(&item.Id, &item.Nombre, &item.Estado, &item.UrlFoto, &item.EsInventariable, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn, &item.Categoria)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Producto no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}
func NewProductoRepository(pool *pgxpool.Pool) *ProductoRepository {
	return &ProductoRepository{pool: pool}
}

var _ port.ProductoRepository = (*ProductoRepository)(nil)
