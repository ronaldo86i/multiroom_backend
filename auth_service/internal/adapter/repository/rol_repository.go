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

type RolRepository struct {
	pool *pgxpool.Pool
}

func (r RolRepository) RegistrarRol(ctx context.Context, request *domain.RolRequest) (*int, error) {
	// 1. Iniciar Transacción
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	var committed bool
	defer func() {
		if !committed {
			if Err := tx.Rollback(ctx); Err != nil {
				log.Println("Error durante rollback:", Err)
			}
		}
	}()

	// 2. Insertar el ROL
	var id int
	query := `INSERT INTO rol(nombre, estado) VALUES ($1,$2) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, request.Estado).Scan(&id)

	if err != nil {
		log.Println(err)
		var pgErr *pgconn.PgError
		// Error 23505: Unique Violation (Nombre duplicado)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("El rol '%s' ya existe.", request.Nombre))
		}
		log.Println("Error al insertar rol:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// 3. Insertar los PERMISOS (Tabla intermedia) -- SECCIÓN NUEVA --
	if len(request.PermisosIds) > 0 {
		queryPermisos := `INSERT INTO rol_permiso (rol_id, permiso_id) VALUES ($1, $2)`

		for _, permisoId := range request.PermisosIds {
			_, err := tx.Exec(ctx, queryPermisos, id, permisoId)
			if err != nil {
				log.Printf("Error al asignar permiso ID %d al rol %d: %v", permisoId, id, err)

				// IMPLEMENTACIÓN DE LA VALIDACIÓN
				var pgErr *pgconn.PgError

				// 23503 es el código PostgreSQL para violación de llave foránea
				if errors.As(err, &pgErr) && pgErr.Code == "23503" {
					// Como acabamos de crear el Rol, sabemos que el rol_id sí existe.
					// Por lo tanto, el que falló es el permiso_id.
					return nil, datatype.NewBadRequestError(fmt.Sprintf("El permiso con ID %d no existe.", permisoId))
				}

				// Si es otro error (conexión, disco lleno, etc.)
				return nil, datatype.NewInternalServerErrorGeneric()
			}
		}
	}

	// 4. Confirmar Transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al hacer commit:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return &id, nil
}

func (r RolRepository) ModificarRolById(ctx context.Context, id *int, request *domain.RolRequest) error {
	// 1. Iniciar Transacción
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		return datatype.NewInternalServerErrorGeneric()
	}

	var committed bool
	defer func() {
		if !committed {
			if Err := tx.Rollback(ctx); Err != nil {
				log.Println("Error durante rollback:", Err)
			}
		}
	}()

	// =========================================================================
	// 🛡️ PROTECCIÓN DE ROL CRÍTICO (ADMIN)
	// =========================================================================
	// Antes de modificar, consultamos qué rol es.
	var currentNombre string
	err = tx.QueryRow(ctx, "SELECT nombre FROM rol WHERE id = $1", *id).Scan(&currentNombre)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return datatype.NewNotFoundError("Rol no encontrado")
		}
		log.Println("Error verificando rol:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// Si intentan tocar el ADMIN, bloqueamos.
	if currentNombre == "ADMIN" {
		return datatype.NewForbiddenError("Acción denegada: El rol 'ADMIN' es crítico para el sistema y no se puede modificar.")
	}
	// =========================================================================

	// 2. Actualizar datos básicos del ROL
	query := `UPDATE rol SET nombre=$1, estado=$2 WHERE id=$3`

	ct, err := tx.Exec(ctx, query, request.Nombre, request.Estado, *id)
	if err != nil {
		log.Println(err)
		var pgErr *pgconn.PgError
		// Error de Nombre Duplicado (Si cambias el nombre a uno que ya existe)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return datatype.NewBadRequestError(fmt.Sprintf("El rol '%s' ya existe.", request.Nombre))
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	// (Opcional) Doble verificación, aunque el SELECT anterior ya validó si existe.
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Rol no encontrado")
	}

	// 3. Sincronizar PERMISOS

	// 3.1. Borrar permisos antiguos (Wipe)
	queryDelete := `DELETE FROM rol_permiso WHERE rol_id = $1`
	_, err = tx.Exec(ctx, queryDelete, *id)
	if err != nil {
		log.Println("Error limpiando permisos antiguos:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// 3.2. Insertar permisos nuevos (Replace)
	if len(request.PermisosIds) > 0 {
		queryInsert := `INSERT INTO rol_permiso (rol_id, permiso_id) VALUES ($1, $2)`

		for _, permisoId := range request.PermisosIds {
			_, err := tx.Exec(ctx, queryInsert, *id, permisoId)
			if err != nil {
				log.Printf("Error al asignar permiso ID %d al rol %d: %v", permisoId, *id, err)

				var pgErr *pgconn.PgError
				// Error 23503: Foreign Key Violation (El permiso enviado no existe)
				if errors.As(err, &pgErr) && pgErr.Code == "23503" {
					return datatype.NewBadRequestError(fmt.Sprintf("El permiso con ID %d no existe.", permisoId))
				}

				return datatype.NewInternalServerErrorGeneric()
			}
		}
	}

	// 4. Confirmar Transacción
	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (r RolRepository) ListarRoles(ctx context.Context, _ map[string]string) (*[]domain.RolInfo, error) {
	query := `SELECT r.id,r.nombre,r.estado,r.creado_en FROM rol r ORDER BY r.nombre`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		log.Println("Error al consultar roles:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Permiso no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.RolInfo, 0)
	for rows.Next() {
		var item domain.RolInfo
		err := rows.Scan(&item.Id, &item.Nombre, &item.Estado, &item.CreadoEn)
		if err != nil {
			log.Println("Error al leer permiso:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (r RolRepository) ObtenerRolById(ctx context.Context, id *int) (*domain.Rol, error) {
	query := `
    SELECT 
        r.id,
        r.nombre,
        r.estado,
        r.creado_en,
        -- Agregamos los permisos como un Array JSON
        COALESCE(
           json_agg(
              json_build_object(
                 'id', p.id,
                 'nombre', p.nombre,
                 'descripcion', p.descripcion,
                 'icono', p.icono,
              	 'creadoEn',p.creado_en
              )
           ) FILTER (WHERE p.id IS NOT NULL), -- Filtramos nulos si el rol no tiene permisos
           '[]'
        ) AS permisos
    FROM rol r
    LEFT JOIN rol_permiso rp ON r.id = rp.rol_id
    LEFT JOIN permiso p ON rp.permiso_id = p.id
    WHERE r.id = $1
    GROUP BY r.id
    `
	var item domain.Rol
	err := r.pool.QueryRow(ctx, query, *id).Scan(&item.Id, &item.Nombre, &item.Estado, &item.CreadoEn, &item.Permisos)
	if err != nil {
		log.Println("Error al consultar permiso:", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Rol no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func NewRolRepository(pool *pgxpool.Pool) *RolRepository {
	return &RolRepository{pool: pool}
}

var _ port.RolRepository = (*RolRepository)(nil)
