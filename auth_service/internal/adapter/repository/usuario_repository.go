package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
)

type UsuarioRepository struct {
	pool *pgxpool.Pool
}

func (u UsuarioRepository) DeshabilitarUsuario(ctx context.Context, id *int) error {
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

	query := `UPDATE usuario d SET estado = 'Inactivo',eliminado_en = now() WHERE id = $1`
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

func (u UsuarioRepository) HabilitarUsuario(ctx context.Context, id *int) error {
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
	query := `UPDATE usuario d SET estado = 'Activo',eliminado_en = NULL WHERE id = $1`
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

func (u UsuarioRepository) ObtenerListaUsuarios(ctx context.Context) (*[]domain.UsuarioInfo, error) {
	var list = make([]domain.UsuarioInfo, 0)
	query := `SELECT u.id,u.username,u.estado,u.creado_en FROM usuario u`
	rows, err := u.pool.Query(ctx, query)
	if err != nil {
		log.Println("Error al obtener lista de usuarios:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.UsuarioInfo
		err = rows.Scan(&item.Id, &item.Username, &item.Estado, &item.CreadoEn)
		if err != nil {
			log.Println("Error al escanear usuario:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (u UsuarioRepository) RegistrarUsuario(ctx context.Context, request *domain.UsuarioRequest) (*int, error) {
	// Iniciar transacción
	tx, err := u.pool.Begin(ctx)
	if err != nil {
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
	query := `INSERT INTO usuario (username, password_hash) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Username, string(passwordHash)).Scan(&usuarioId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, datatype.NewConflictError("El nombre de usuario ya está registrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &usuarioId, nil
}

func (u UsuarioRepository) ObtenerUsuarioById(ctx context.Context, id *int) (*domain.Usuario, error) {
	var usuario domain.Usuario
	query := `SELECT u.id,u.username,u.password_hash,u.estado,u.creado_en FROM usuario u WHERE u.id = $1 LIMIT 1`
	err := u.pool.QueryRow(ctx, query, *id).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn)
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
func (u UsuarioRepository) ObtenerUsuarioByUsername(ctx context.Context, username *string) (*domain.Usuario, error) {
	var usuario domain.Usuario
	query := `SELECT u.id,u.username,u.password_hash,u.estado,u.creado_en FROM usuario u WHERE u.username = $1 LIMIT 1`
	err := u.pool.QueryRow(ctx, query, *username).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn)
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

func NewUsuarioRepository(pool *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{pool: pool}
}

var _ port.UsuarioRepository = (*UsuarioRepository)(nil)
