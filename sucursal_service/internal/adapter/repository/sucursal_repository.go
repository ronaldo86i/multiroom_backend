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

type SucursalRepository struct {
	pool *pgxpool.Pool
}

func (s SucursalRepository) RegistrarSucursal(ctx context.Context, request *domain.SucursalRequest) (*int, error) {
	// 1. Iniciar Transacción
	tx, err := s.pool.Begin(ctx)
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

	// 2. Insertar la Sucursal
	var sucursalId int
	query := `INSERT INTO sucursal (nombre, estado, pais_id) VALUES ($1, 'Activo', $2) RETURNING id`

	err = tx.QueryRow(ctx, query, request.Nombre, request.PaisId).Scan(&sucursalId)
	if err != nil {
		log.Println("Ha ocurrido un error en la transacción:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_nombre_sucursal" {
					return nil, datatype.NewConflictError("Ya existe una sucursal con ese nombre en ese país")
				}
				return nil, datatype.NewInternalServerErrorGeneric()
			}
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 3. Asignar todos los productos existentes a esta nueva sucursal
	queryAsignarProductos := `
        INSERT INTO producto_sucursal (producto_id, sucursal_id, precio, estado)
        SELECT id, $1, 0, 'Activo'
        FROM producto
    `

	_, err = tx.Exec(ctx, queryAsignarProductos, sucursalId)
	if err != nil {
		log.Println("Error al asignar productos a la nueva sucursal:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 4. Confirmar Transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &sucursalId, nil
}

func (s SucursalRepository) ModificarSucursal(ctx context.Context, id *int, request *domain.SucursalRequest) error {
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sucursal SET nombre=$1,pais_id=$2,actualizado_en=now() WHERE id=$3`
	ct, err := tx.Exec(ctx, query, request.Nombre, request.PaisId, *id)
	if err != nil {
		log.Println("Ha ocurrido un error en la transacción:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_nombre_sucursal" {
					return datatype.NewConflictError("Ya existe una sucursal con ese nombre en ese país")
				}
				// Otra violación única
				return datatype.NewInternalServerErrorGeneric()
			}
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sucursal no encontrado")
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return nil
}

func (s SucursalRepository) ObtenerSucursalById(ctx context.Context, id *int) (*domain.Sucursal, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/paises/")
	query := `
SELECT 
    s.id,
    s.nombre,
    s.estado,
    s.creado_en,
    s.actualizado_en,
    s.eliminado_en,
    json_build_object(
    	'id', p.id,
    	'nombre',p.nombre,
    	'urlFoto',($1::text || p.id::text || '/' || p.archivo),
    	'estado',p.estado,
    	'creadoEn',p.creado_en
    ) AS pais
FROM sucursal s
LEFT JOIN public.pais p on s.pais_id = p.id
WHERE s.id=$2 
LIMIT 1
`

	var sucursal domain.Sucursal
	err := s.pool.QueryRow(ctx, query, fullHostname, *id).Scan(&sucursal.Id, &sucursal.Nombre, &sucursal.Estado, &sucursal.CreadoEn, &sucursal.ActualizadoEn, &sucursal.EliminadoEn, &sucursal.Pais)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Sucursal no encontrado")
		}
	}

	return &sucursal, nil
}

func (s SucursalRepository) ObtenerListaSucursales(ctx context.Context, filtros map[string]string) (*[]domain.SucursalInfo, error) {
	var filters []string
	var args []interface{}
	i := 1

	// Si hay paisId en filtros
	if paisIdStr := filtros["paisId"]; paisIdStr != "" {
		paisId, err := strconv.Atoi(paisIdStr)
		if err != nil {
			log.Println("Error al convertir paisId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de paisId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s.pais_id = $%d", i))
		args = append(args, paisId)
		i++
	}

	query := `SELECT s.id, s.nombre, s.estado, s.creado_en FROM sucursal s`

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " ORDER BY s.id"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al ejecutar query:", err, query, args)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	var list []domain.SucursalInfo
	for rows.Next() {
		var sucursal domain.SucursalInfo
		err = rows.Scan(&sucursal.Id, &sucursal.Nombre, &sucursal.Estado, &sucursal.CreadoEn)
		if err != nil {
			log.Println("Error al escanear fila:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, sucursal)
	}

	return &list, nil
}

func (s SucursalRepository) HabilitarSucursal(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sucursal SET estado= 'Activo',actualizado_en=now(),eliminado_en=NULL WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sucursal no encontrado")
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

func (s SucursalRepository) DeshabilitarSucursal(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sucursal SET estado= 'Inactivo',actualizado_en=now(),eliminado_en=now() WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sucursal no encontrado")
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

func NewSucursalRepository(pool *pgxpool.Pool) SucursalRepository {
	return SucursalRepository{pool: pool}
}

var _ port.SucursalRepository = (*SucursalRepository)(nil)
