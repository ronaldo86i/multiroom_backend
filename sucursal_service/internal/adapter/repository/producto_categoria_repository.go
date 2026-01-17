package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductoCategoriaRepository struct {
	pool *pgxpool.Pool
}

func (p ProductoCategoriaRepository) RegistrarCategoria(ctx context.Context, request *domain.ProductoCategoriaRequest) (*int, error) {
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
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

	query := `INSERT INTO categoria_producto(nombre, descripcion, estado) VALUES ($1, $2, $3) RETURNING id`

	var id int
	err = tx.QueryRow(ctx, query, request.Nombre, request.Descripcion, request.Estado).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError
		// Si es un error de DUPLICIDAD (Unique Violation - código 23505)
		// Esto pasa si el nombre de la categoría ya existe en la BD.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("La categoría '%s' ya existe.", request.Nombre))
		}
		// Cualquier otro error de base de datos
		log.Println("Error al insertar categoría:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &id, nil
}
func (p ProductoCategoriaRepository) ModificarCategoriaById(ctx context.Context, id *int, request *domain.ProductoCategoriaRequest) error {
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
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
	query := `UPDATE categoria_producto SET nombre=$1,estado=$2,descripcion=$3,actualizado_en=now() WHERE id=$4`
	ct, err := tx.Exec(ctx, query, request.Nombre, request.Estado, request.Descripcion, *id)
	if err != nil {
		var pgErr *pgconn.PgError
		// Si es un error de DUPLICIDAD (Unique Violation - código 23505)
		// Esto pasa si el nombre de la categoría ya existe en la BD.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return datatype.NewBadRequestError(fmt.Sprintf("La categoría '%s' ya existe.", request.Nombre))
		}
		// Cualquier otro error de base de datos
		log.Println("Error al insertar categoría:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewBadRequestError("Categoría no encontrada")
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

func (p ProductoCategoriaRepository) ListarCategorias(ctx context.Context, filtros map[string]string) (*[]domain.ProductoCategoria, error) {
	query := `SELECT cp.id,cp.nombre,cp.descripcion,cp.estado,cp.creado_en,cp.actualizado_en FROM categoria_producto cp`
	var filters []string
	var args []interface{}
	var i = 1
	if estado := filtros["estado"]; estado != "" {
		filters = append(filters, fmt.Sprintf("cp.estado = $%d", i))
		args = append(args, estado)
		i++
	}
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += ` ORDER BY cp.nombre ASC`
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, datatype.NewInternalServerError("Error al obtener lista de categorías de productos")
	}
	defer rows.Close()
	list := make([]domain.ProductoCategoria, 0)
	for rows.Next() {
		var item domain.ProductoCategoria
		err := rows.Scan(&item.Id, &item.Nombre, &item.Descripcion, &item.Estado, &item.CreadoEn, &item.ActualizadoEn)
		if err != nil {
			log.Println("Error al obtener lista de categorías de productos:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (p ProductoCategoriaRepository) ObtenerCategoriaById(ctx context.Context, id *int) (*domain.ProductoCategoria, error) {
	query := `SELECT cp.id,cp.nombre,cp.descripcion,cp.estado,cp.creado_en,cp.actualizado_en FROM categoria_producto cp WHERE cp.id=$1`
	var item domain.ProductoCategoria
	err := p.pool.QueryRow(ctx, query, *id).Scan(&item.Id, &item.Nombre, &item.Descripcion, &item.Estado, &item.CreadoEn, &item.ActualizadoEn)
	if err != nil {
		log.Println("Error al obtener categoría de producto:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func NewProductoCategoriaRepository(pool *pgxpool.Pool) *ProductoCategoriaRepository {
	return &ProductoCategoriaRepository{pool: pool}
}

var _ port.ProductoCategoriaRepository = (*ProductoCategoriaRepository)(nil)
