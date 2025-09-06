package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
)

type UsuarioSucursalRepository struct {
	pool *pgxpool.Pool
}

func (u UsuarioSucursalRepository) ObtenerUsuarioSucursalByUsername(ctx context.Context, username *string) (*domain.UsuarioSucursal, error) {
	var usuario domain.UsuarioSucursal
	query := `
	SELECT 
		us.id, 
		us.username, 
		us.password_hash,
		us.estado,
		us.creado_en,
		us.eliminado_en,
		json_build_object(
			'id', s.id,
			'nombre',s.nombre,
			'creadoEn',s.creado_en
		) AS sucursal
	FROM usuario_sucursal us
	LEFT JOIN public.sucursal s on us.sucursal_id = s.id
	WHERE us.username = $1
	LIMIT 1
`
	err := u.pool.QueryRow(ctx, query, *username).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn, &usuario.EliminadoEn, &usuario.Sucursal)
	if err != nil {
		log.Println("Error al escanear el usuario", err)
		// Si no hay registros
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Usuario no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &usuario, nil
}

func (u UsuarioSucursalRepository) RegistrarUsuarioSucursal(ctx context.Context, request *domain.UsuarioSucursalRequest) (*int, error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	// Generar hash de la contraseña
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, datatype.NewInternalServerError("Error al hashear la contraseña")
	}

	var usuarioId int
	query := `INSERT INTO usuario_sucursal (username, password_hash,sucursal_id) VALUES ($1, $2,$3) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Username, string(passwordHash), request.SucursalId).Scan(&usuarioId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "usuario_sucursal_username_key" {
				return nil, datatype.NewConflictError("El nombre de usuario ya está registrado")
			} else if pgErr.ConstraintName == "usuario_sucursal_sucursal_id_key" {
				return nil, datatype.NewConflictError("La sucursal ya se encuentra ocupada")
			}
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return &usuarioId, nil
}

func (u UsuarioSucursalRepository) ModificarUsuarioSucursal(ctx context.Context, id *int, request *domain.UsuarioSucursalRequest) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		return datatype.NewInternalServerErrorGeneric()
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	// Generar hash de la contraseña
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return datatype.NewInternalServerError("Error al hashear la contraseña")
	}

	query := `UPDATE usuario_sucursal SET username=$1,password_hash=$2,sucursal_id=$3 WHERE id=$4`
	ct, err := tx.Exec(ctx, query, request.Username, string(passwordHash), request.SucursalId, *id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ColumnName == "username" {
				return datatype.NewConflictError("El nombre de usuario ya está registrado")
			} else if pgErr.ColumnName == "sucursal_id" {
				return datatype.NewConflictError("La sucursal ya se encuentra ocupada")
			}
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Usuario no encontrado")
	}
	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (u UsuarioSucursalRepository) ObtenerListaUsuariosSucursal(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioSucursalInfo, error) {
	query := `
SELECT 
    us.id, 
    us.username, 
    us.estado,
    us.creado_en,
    json_build_object(
    	'id', s.id,
    	'nombre',s.nombre,
    	'creadoEn',s.creado_en
    ) AS sucursal
FROM usuario_sucursal us
LEFT JOIN public.sucursal s on us.sucursal_id = s.id
ORDER BY us.id`
	rows, err := u.pool.Query(ctx, query)
	if err != nil {
		log.Println("Error al obtener lista de usuarios:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	var list = make([]domain.UsuarioSucursalInfo, 0)
	for rows.Next() {
		var item domain.UsuarioSucursalInfo
		err = rows.Scan(&item.Id, &item.Username, &item.Estado, &item.CreadoEn, &item.Sucursal)
		if err != nil {
			log.Println("Error al escanear usuario:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}

	return &list, nil
}

func (u UsuarioSucursalRepository) ObtenerUsuarioSucursalById(ctx context.Context, id *int) (*domain.UsuarioSucursal, error) {
	var usuario domain.UsuarioSucursal
	query := `
SELECT 
    us.id, 
    us.username, 
    us.password_hash,
    us.estado,
    us.creado_en,
    us.eliminado_en,
    json_build_object(
    	'id', s.id,
    	'nombre',s.nombre,
    	'creadoEn',s.creado_en
    ) AS sucursal
FROM usuario_sucursal us
LEFT JOIN public.sucursal s on us.sucursal_id = s.id
WHERE us.id = $1
LIMIT 1
`
	err := u.pool.QueryRow(ctx, query, *id).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn, &usuario.EliminadoEn, &usuario.Sucursal)
	if err != nil {
		// Si no hay registros
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Usuario no encontrado")
		}
		log.Println("Error al escanear el usuario")
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &usuario, nil
}

func (u UsuarioSucursalRepository) HabilitarUsuarioSucursal(ctx context.Context, id *int) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		return datatype.NewInternalServerErrorGeneric()
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `UPDATE usuario_sucursal SET estado = 'Activo',eliminado_en = NULL WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Usuario no encontrado")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (u UsuarioSucursalRepository) DeshabilitarUsuarioSucursal(ctx context.Context, id *int) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		return datatype.NewInternalServerErrorGeneric()
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `UPDATE usuario_sucursal SET estado = 'Inactivo',eliminado_en = now() WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Usuario no encontrado")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func NewUsuarioSucursalRepository(pool *pgxpool.Pool) *UsuarioSucursalRepository {
	return &UsuarioSucursalRepository{pool: pool}
}

var _ port.UsuarioSucursalRepository = (*UsuarioSucursalRepository)(nil)
