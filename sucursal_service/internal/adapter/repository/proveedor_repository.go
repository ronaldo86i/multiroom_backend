package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProveedorRepository struct {
	pool *pgxpool.Pool
}

func (p ProveedorRepository) ListarProveedores(ctx context.Context, filtros map[string]string) (*[]domain.Proveedor, error) {
	query := `SELECT p.id,p.nombre,p.estado,p.email,p.celular,p.telefono,p.creado_en,p.actualizado_en,p.eliminado_en FROM proveedor p`
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.Proveedor, 0)
	for rows.Next() {
		var item domain.Proveedor
		err = rows.Scan(&item.Id, &item.Nombre, &item.Estado, &item.Email, &item.Celular, &item.Telefono, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn)
		if err != nil {
			log.Println("Error scanning rows:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (p ProveedorRepository) ObtenerProveedorById(ctx context.Context, id *int) (*domain.Proveedor, error) {
	query := `SELECT p.id,p.nombre,p.estado,p.email,p.celular,p.telefono,p.creado_en,p.actualizado_en,p.eliminado_en FROM proveedor p WHERE p.id=$1 LIMIT 1`
	var proveedor domain.Proveedor
	err := p.pool.QueryRow(ctx, query, *id).
		Scan(&proveedor.Id, &proveedor.Nombre, &proveedor.Estado, &proveedor.Email, &proveedor.Celular, &proveedor.Telefono, &proveedor.CreadoEn, &proveedor.ActualizadoEn, &proveedor.EliminadoEn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Proveedor no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &proveedor, nil
}

func (p ProveedorRepository) RegistrarProveedor(ctx context.Context, request *domain.ProveedorRequest) (*int, error) {
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
	var proveedorId int
	query := `INSERT INTO proveedor(nombre, estado,email,celular,telefono) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, request.Estado, request.Email, request.Celular, request.Telefono).Scan(&proveedorId)
	if err != nil {
		log.Println("Error al insertar proveedor:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return &proveedorId, nil
}

func (p ProveedorRepository) ModificarProveedor(ctx context.Context, id *int, request *domain.ProveedorRequest) error {
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

	query := `UPDATE proveedor SET nombre=$1,estado=$2,email=$3,celular=$4,telefono=$5,actualizado_en=now() WHERE id=$6`
	ct, err := tx.Exec(ctx, query, request.Nombre, request.Estado, request.Email, request.Celular, request.Telefono, *id)
	if err != nil {
		log.Println("Error al modificar proveedor:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Proveedor no encontrado")
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

func NewProveedorRepository(pool *pgxpool.Pool) *ProveedorRepository {
	return &ProveedorRepository{pool: pool}
}

var _ port.ProveedorRepository = (*ProveedorRepository)(nil)
