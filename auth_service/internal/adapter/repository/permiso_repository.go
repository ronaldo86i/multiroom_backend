package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PermisoRepository struct {
	pool *pgxpool.Pool
}

func (p PermisoRepository) ListarPermisos(ctx context.Context, _ map[string]string) (*[]domain.Permiso, error) {
	query := "SELECT p.id,p.nombre,p.descripcion,p.icono,p.creado_en FROM permiso p ORDER BY p.nombre"
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		log.Println("Error al consultar permisos:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.Permiso, 0)
	for rows.Next() {
		var item domain.Permiso
		err = rows.Scan(&item.Id, &item.Nombre, &item.Descripcion, &item.Icono, &item.CreadoEn)
		if err != nil {
			log.Println("Error al leer permiso:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (p PermisoRepository) ObtenerPermisoById(ctx context.Context, id *int) (*domain.Permiso, error) {
	var item domain.Permiso
	query := "SELECT p.id,p.nombre,p.descripcion,p.icono,p.creado_en FROM permiso p WHERE p.id=$1"
	err := p.pool.QueryRow(ctx, query, *id).Scan(&item.Id, &item.Nombre, &item.Descripcion, &item.Icono, &item.CreadoEn)
	if err != nil {
		log.Println("Error al consultar permiso:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Permiso no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func (p PermisoRepository) RegistrarPermiso(ctx context.Context, request *domain.PermisoRequest) (*int, error) {
	// 1. Iniciar Transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
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

	// 2. Insertar el PERMISO nuevo
	var id int
	query := `INSERT INTO permiso(nombre, icono, descripcion) VALUES ($1, $2, $3) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, request.Icono, request.Descripcion).Scan(&id)
	if err != nil {
		log.Println(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("El permiso '%s' ya existe.", request.Nombre))
		}
		log.Println("Error al insertar permiso:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// A. Buscamos el id del rol 'ADMIN' (No asumimos que es el ID 1)
	var adminRoleId int
	err = tx.QueryRow(ctx, "SELECT id FROM rol WHERE nombre = 'ADMIN'").Scan(&adminRoleId)

	if err != nil {
		// Si no encuentra el rol ADMIN, logueamos advertencia pero NO fallamos la creación del permiso.
		// Opcional: Si es crítico que ADMIN lo tenga, retorna error aquí.
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("ADVERTENCIA: Se creó el permiso pero no se pudo asignar al rol 'ADMIN' porque no existe.")
		} else {
			log.Println("Error buscando rol ADMIN:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
	} else {
		// B. Si encontramos el rol ADMIN, le asignamos el permiso
		queryAssign := `INSERT INTO rol_permiso (rol_id, permiso_id) VALUES ($1, $2)`
		_, err = tx.Exec(ctx, queryAssign, adminRoleId, id)
		if err != nil {
			log.Printf("Error asignando permiso %d al rol ADMIN (%d): %v", id, adminRoleId, err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
	}

	// 4. Confirmar Transacción
	err = tx.Commit(ctx)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return &id, nil
}

func (p PermisoRepository) ModificarPermisoById(ctx context.Context, id *int, request *domain.PermisoRequest) error {

	// 1. Iniciar Transacción
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

	// 2. Ejecutar UPDATE
	// IMPORTANTE: No incluimos 'nombre' en el SET para proteger la integridad del sistema.
	// Solo permitimos cambiar la descripción y el icono.
	query := `UPDATE permiso SET descripcion = $1, icono = $2 WHERE id = $3`

	ct, err := tx.Exec(ctx, query, request.Descripcion, request.Icono, *id)

	if err != nil {
		log.Println("Error al modificar permiso:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// 3. Validar si existía el registro
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("El permiso no existe")
	}

	// 4. Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return nil
}
func NewPermisoRepository(pool *pgxpool.Pool) *PermisoRepository {
	return &PermisoRepository{pool: pool}
}

var _ port.PermisoRepository = (*PermisoRepository)(nil)
