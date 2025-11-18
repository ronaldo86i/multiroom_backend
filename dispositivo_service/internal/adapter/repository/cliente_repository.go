package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/domain/datatype"
	"multiroom/dispositivo-service/internal/core/port"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClienteRepository struct {
	pool *pgxpool.Pool
}

func (c ClienteRepository) RegistrarCliente(ctx context.Context, request *domain.ClienteRequest) (*int64, error) {
	// Iniciar transacción
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar la transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	// Insertar usuario
	var clienteId int64
	query := `INSERT INTO cliente(celular, codigo_pais, nombres, apellidos, fecha_nacimiento) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Celular, request.CodigoPais, request.Nombres, request.Apellidos, request.FechaNacimiento).Scan(&clienteId)
	if err != nil {
		log.Println("Error en la transacción:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_cliente_celular" {
					return nil, datatype.NewConflictError("Ya existe un cliente con ese celular/teléfono")
				}
			case "23514": // check_violation
				if pgErr.ConstraintName == "check_fecha_nacimiento" {
					return nil, datatype.NewBadRequestError("La fecha de nacimiento no puede ser mayor a hoy")
				}
			}
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar la transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return &clienteId, nil
}

func (c ClienteRepository) ModificarCliente(ctx context.Context, id *int, request *domain.ClienteRequest) error {
	// Iniciar transacción
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	// Actualizar cliente
	query := `UPDATE cliente SET celular=$1,codigo_pais=$2,nombres=$3,apellidos=$4,fecha_nacimiento=$5,actualizado_en=now() WHERE id=$6`
	ct, err := tx.Exec(ctx, query, request.Celular, request.CodigoPais, request.Nombres, request.Apellidos, request.FechaNacimiento, *id)
	if err != nil {
		log.Println("Error en la transacción:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				if pgErr.ConstraintName == "unique_cliente_celular" {
					return datatype.NewConflictError("Ya existe un cliente con ese celular/teléfono")
				}
			case "23514": // check_violation
				if pgErr.ConstraintName == "check_fecha_nacimiento" {
					return datatype.NewBadRequestError("La fecha de nacimiento no puede ser mayor a hoy")
				}
			}
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		log.Println("Cliente no encontrado:", *id)
		return datatype.NewBadRequestError("Cliente no encontrado")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (c ClienteRepository) ObtenerListaClientes(ctx context.Context, filtros map[string]string) (*[]domain.ClienteInfo, error) {
	query := `SELECT c.id,c.nombres,c.apellidos,c.codigo_pais,c.celular,c.fecha_nacimiento,c.estado,c.creado_en FROM cliente c ORDER BY c.id`
	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		log.Println("Error al listar la clientes:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	var list = make([]domain.ClienteInfo, 0)
	for rows.Next() {
		var cliente domain.ClienteInfo
		err = rows.Scan(&cliente.Id, &cliente.Nombres, &cliente.Apellidos, &cliente.CodigoPais, &cliente.Celular, &cliente.FechaNacimiento, &cliente.Estado, &cliente.CreadoEn)
		if err != nil {
			log.Println("Error al listar la clientes:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, cliente)
	}
	return &list, nil
}

func (c ClienteRepository) ObtenerClienteDetailById(ctx context.Context, id *int) (*domain.ClienteDetail, error) {
	query := `SELECT c.id,c.nombres,c.apellidos,c.codigo_pais,c.celular,c.fecha_nacimiento,c.estado,c.creado_en,c.actualizado_en,c.eliminado_en FROM cliente c WHERE c.id=$1 LIMIT 1`
	var cliente domain.ClienteDetail
	err := c.pool.QueryRow(ctx, query, *id).
		Scan(&cliente.Id, &cliente.Nombres, &cliente.Apellidos, &cliente.CodigoPais, &cliente.Celular, &cliente.FechaNacimiento, &cliente.Estado, &cliente.CreadoEn, &cliente.ActualizadoEn, &cliente.EliminadoEn)
	if err != nil {
		log.Println("Error al obtener cliente", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Cliente no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &cliente, nil
}

func (c ClienteRepository) HabilitarCliente(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Transacción: Habilitar cliente
	// Actualizar cliente
	query := `UPDATE cliente SET estado='Activo',actualizado_en=now(),eliminado_en=NULL WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Error en la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Cliente no encontrado")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (c ClienteRepository) DeshabilitarCliente(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Transacción: Habilitar cliente
	// Actualizar cliente
	query := `UPDATE cliente SET estado='Inactivo',actualizado_en=now(),eliminado_en=now() WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Error en la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Cliente no encontrado")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar la transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func NewClienteRepository(pool *pgxpool.Pool) *ClienteRepository {
	return &ClienteRepository{pool: pool}
}

var _ port.ClienteRepository = (*ClienteRepository)(nil)
