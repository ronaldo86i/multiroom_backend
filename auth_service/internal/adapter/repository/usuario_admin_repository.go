package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
)

type UsuarioAdminRepository struct {
	pool *pgxpool.Pool
}

func (u UsuarioAdminRepository) ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error) {
	var usuario domain.UsuarioAdmin
	query := `
SELECT 
    u.id,
    u.username,
    u.password_hash,
    u.estado,
    u.creado_en,
    COALESCE(
        json_agg(
            json_build_object(
                'id', r.id,
                'nombre', r.nombre
            )
        ) FILTER (WHERE r.estado='Activo'), '[]'
    ) AS roles
FROM usuario_admin u
INNER JOIN usuario_admin_rol uar ON u.id = uar.usuario_admin_id
LEFT JOIN rol r ON uar.rol_id = r.id
WHERE u.username = $1
GROUP BY u.id
LIMIT 1;
`
	err := u.pool.QueryRow(ctx, query, *username).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn, &usuario.Roles)
	if err != nil {
		log.Println("Error al obtener usuario_admin:", err)
		// Si no hay registros
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Usuario no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &usuario, nil
}

func (u UsuarioAdminRepository) ObtenerUsuarioAdminById(ctx context.Context, id *int) (*domain.UsuarioAdmin, error) {
	var usuario domain.UsuarioAdmin
	query := `
SELECT 
    u.id,
    u.username,
    u.password_hash,
    u.estado,
    u.creado_en,
    COALESCE(
        json_agg(
            json_build_object(
                'id', r.id,
                'nombre', r.nombre
            )
        ) FILTER (WHERE r.estado='Activo'), '[]'
    ) AS roles
FROM usuario_admin u
INNER JOIN usuario_admin_rol uar ON u.id = uar.usuario_admin_id
LEFT JOIN rol r ON uar.rol_id = r.id
WHERE u.id = $1
GROUP BY u.id
LIMIT 1;
`
	err := u.pool.QueryRow(ctx, query, *id).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn, &usuario.Roles)
	if err != nil {
		log.Println("Error al obtener usuario_admin:", err)
		// Si no hay registros
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Usuario no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &usuario, nil
}

func NewUsuarioAdminRepository(pool *pgxpool.Pool) *UsuarioAdminRepository {
	return &UsuarioAdminRepository{pool: pool}
}

var _ port.UsuarioAdminRepository = (*UsuarioAdminRepository)(nil)
