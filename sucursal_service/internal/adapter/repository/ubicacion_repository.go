package repository

import (
	"context"
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

type UbicacionRepository struct {
	pool *pgxpool.Pool
}

func (u UbicacionRepository) HabilitarUbicacion(ctx context.Context, id *int) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción")
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	query := `UPDATE ubicacion SET estado='Activo' WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Ha ocurrido un error en la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Ubicación no encontrado")
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (u UbicacionRepository) DeshabilitarUbicacion(ctx context.Context, id *int) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción")
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `UPDATE ubicacion SET estado='Inactivo' WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Ha ocurrido un error en la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Ubicación no encontrado")
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (u UbicacionRepository) RegistrarUbicacion(ctx context.Context, request *domain.UbicacionRequest) (*int, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción")
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	var ubicacionId int
	query := `INSERT INTO ubicacion(nombre, estado, es_vendible, prioridad_venta, sucursal_id) VALUES($1,$2,$3,$4,$5) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, request.Estado, request.EsVendible, request.PrioridadVenta, request.SucursalId).Scan(&ubicacionId)
	if err != nil {
		log.Println("Ha ocurrido un error en la transacción:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_ubicacion" {
					return nil, datatype.NewConflictError("Ya existe una ubicación con ese nombre en esa sucursal")
				}
				return nil, datatype.NewInternalServerErrorGeneric()
			}
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &ubicacionId, nil
}

func (u UbicacionRepository) ModificarUbicacionById(ctx context.Context, id *int, request *domain.UbicacionRequest) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción")
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	query := `UPDATE ubicacion SET nombre=$1,estado=$2,es_vendible=$3,prioridad_venta=$4 WHERE id=$5`
	ct, err := tx.Exec(ctx, query, request.Nombre, request.Estado, request.EsVendible, request.PrioridadVenta, *id)
	if err != nil {
		log.Println("Ha ocurrido un error en la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Ubicación no encontrado")
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (u UbicacionRepository) ListarUbicaciones(ctx context.Context, filtros map[string]string) (*[]domain.Ubicacion, error) {
	var filters []string
	var args []interface{}
	i := 1

	query := `
SELECT 
    u.id,
    u.nombre,
    u.estado,
    u.es_vendible,
    u.prioridad_venta,
        jsonb_build_object(
        'id', s.id,
        'nombre', s.nombre,
        'estado', s.estado,
        'creadoEn', s.creado_en
    ) AS sucursal
FROM ubicacion u
LEFT JOIN sucursal s on s.id = u.sucursal_id
`
	// Si hay sucursalId en filtros
	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("u.sucursal_id = $%d", i))
		args = append(args, sucursalId)
		i++
	}

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " ORDER BY u.id"

	rows, err := u.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al ejecutar query:", err, query, args)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	list := make([]domain.Ubicacion, 0)
	for rows.Next() {
		var item domain.Ubicacion
		err = rows.Scan(&item.Id, &item.Nombre, &item.Estado, &item.EsVendible, &item.PrioridadVenta, &item.Sucursal)
		if err != nil {
			log.Println("Error al escanear fila:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error durante la iteración de filas:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &list, nil
}

func (u UbicacionRepository) ObtenerUbicacionById(ctx context.Context, id *int) (*domain.Ubicacion, error) {
	var item domain.Ubicacion
	query := `
SELECT 
    u.id,
    u.nombre,
    u.estado,
    u.es_vendible,
    u.prioridad_venta,
        jsonb_build_object(
        'id', s.id,
        'nombre', s.nombre,
        'estado', s.estado,
        'creadoEn', s.creado_en
    ) AS sucursal
FROM ubicacion u
LEFT JOIN sucursal s on s.id = u.sucursal_id
WHERE u.id = $1
LIMIT 1
`
	err := u.pool.QueryRow(ctx, query, *id).Scan(&item.Id, &item.Nombre, &item.Estado, &item.EsVendible, &item.PrioridadVenta, &item.Sucursal)
	if err != nil {
		log.Println("Error al escanear ubicación por Id:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Ubicación no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func NewUbicacionRepository(pool *pgxpool.Pool) *UbicacionRepository {
	return &UbicacionRepository{pool: pool}
}

var _ port.UbicacionRepository = (*UbicacionRepository)(nil)
