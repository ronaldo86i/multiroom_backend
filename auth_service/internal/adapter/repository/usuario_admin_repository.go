package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UsuarioAdminRepository struct {
	pool *pgxpool.Pool
}

func (u UsuarioAdminRepository) ListarUsuariosAdmin(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioAdminInfo, error) {
	var filters []string
	var args []interface{}
	var i = 1

	// Construcción de filtros dinámicos
	if val, ok := filtros["search"]; ok && val != "" {
		filters = append(filters, fmt.Sprintf("ua.username ILIKE $%d", i))
		args = append(args, "%"+val+"%") // Búsqueda parcial insensible a mayúsculas
		i++
	}

	if val, ok := filtros["estado"]; ok && val != "" {
		filters = append(filters, fmt.Sprintf("ua.estado = $%d", i))
		args = append(args, val)
		i++
	}

	// Consulta
	query := `
    SELECT 
        ua.id,
        ua.username,
        -- ua.password_hash, -- No devolver el hash en listados por seguridad
        ua.estado,
        ua.creado_en
    FROM usuario_admin ua
    `

	// Inyectar Filtros
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " ORDER BY ua.id DESC"

	// Ejecución
	rows, err := u.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al listar usuarios:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	// Escaneo
	list := make([]domain.UsuarioAdminInfo, 0)

	for rows.Next() {
		var item domain.UsuarioAdminInfo

		// pgx des-serializa automáticamente los JSON arrays a tus structs/slices en Go
		err = rows.Scan(&item.Id, &item.Username, &item.Estado, &item.CreadoEn)
		if err != nil {
			log.Println("Error escanear usuario_admin:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}

		list = append(list, item)
	}

	if rows.Err() != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	return &list, nil
}

func (u UsuarioAdminRepository) RegistrarUsuarioAdmin(ctx context.Context, request *domain.UsuarioAdminRequest) (*int, error) {
	// Hashear la contraseña (CRÍTICO)
	// Costo 10 o 12 es estándar hoy en día.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error al hashear password:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// Iniciar Transacción
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return nil, datatype.NewInternalServerError("Error al iniciar transacción")
	}
	// Defer Rollback: Si la función termina y no se hizo Commit, se deshacen los cambios automáticamente.
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Insertar Usuario
	queryUsuario := `
        INSERT INTO usuario_admin (username, password_hash, estado)
        VALUES ($1, $2, $3)
        RETURNING id
    `
	var newId int
	err = tx.QueryRow(ctx, queryUsuario, request.Username, string(hashedPassword), request.Estado).Scan(&newId)
	if err != nil {
		// Manejo de error por Usuario Duplicado (Código PG 23505)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, datatype.NewConflictError("El nombre de usuario ya existe")
		}
		log.Println("Error al insertar usuario:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	// Asignar Roles (Insertar en tabla intermedia)
	if len(request.RolesIds) > 0 {
		queryRoles := `INSERT INTO usuario_admin_rol (usuario_admin_id, rol_id) VALUES ($1, $2)`
		for _, rolId := range request.RolesIds {
			_, err = tx.Exec(ctx, queryRoles, newId, rolId)
			if err != nil {
				log.Printf("Error al asignar rol %d: %v", rolId, err)
				return nil, datatype.NewInternalServerError("Error al asignar roles")
			}
		}
	}

	// Asignar Sucursales (Insertar en tabla intermedia)
	if len(request.SucursalesIds) > 0 {
		querySucursal := `INSERT INTO usuario_admin_sucursal (usuario_admin_id, sucursal_id) VALUES ($1, $2)`
		for _, sucId := range request.SucursalesIds {
			_, err = tx.Exec(ctx, querySucursal, newId, sucId)
			if err != nil {
				log.Printf("Error al asignar sucursal %d: %v", sucId, err)
				// Aquí podrías validar si el error es de llave foránea (sucursal no existe)
				return nil, datatype.NewInternalServerError("Error al asignar sucursales")
			}
		}
	}

	// Commit
	err = tx.Commit(ctx)
	if err != nil {
		return nil, datatype.NewInternalServerError("Error al confirmar la transacción")
	}
	committed = true
	return &newId, nil
}

func (u UsuarioAdminRepository) ModificarUsuarioAdminById(ctx context.Context, id *int, request *domain.UsuarioAdminRequest) error {

	// Iniciar Transacción
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return datatype.NewInternalServerError("Error al iniciar transacción")
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Actualizar Datos Básicos (Usuario, Estado y opcionalmente Password)
	var queryUpdate string
	var args []interface{}

	if request.Password != "" {
		// CASO A: El usuario envió una nueva contraseña -> La hasheamos y actualizamos
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			return datatype.NewInternalServerError("Error al procesar la contraseña")
		}

		queryUpdate = `
            UPDATE usuario_admin 
            SET username = $1, password_hash = $2, estado = $3
            WHERE id = $4
        `
		args = []interface{}{request.Username, string(hashedPassword), request.Estado, *id}
	} else {
		// CASO B: El usuario NO envió contraseña -> Mantenemos la actual, solo actualizamos datos
		queryUpdate = `
            UPDATE usuario_admin 
            SET username = $1, estado = $2
            WHERE id = $3
        `
		args = []interface{}{request.Username, request.Estado, *id}
	}

	cmdTag, err := tx.Exec(ctx, queryUpdate, args...)
	if err != nil {
		// Manejo de duplicados si intenta poner un username que ya existe
		if strings.Contains(err.Error(), "23505") { // Código SQL State para Unique Violation
			return datatype.NewConflictError("El nombre de usuario ya está en uso")
		}
		log.Println("Error al actualizar usuario_admin:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if cmdTag.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Usuario no encontrado")
	}

	// Sincronizar ROLES
	// Borramos las relaciones viejas
	_, err = tx.Exec(ctx, `DELETE FROM usuario_admin_rol WHERE usuario_admin_id = $1`, *id)
	if err != nil {
		return datatype.NewInternalServerError("Error al limpiar roles antiguos")
	}

	// Insertamos los nuevos (si hay)
	if len(request.RolesIds) > 0 {
		queryRoles := `INSERT INTO usuario_admin_rol (usuario_admin_id, rol_id) VALUES ($1, $2)`
		for _, rolId := range request.RolesIds {
			_, err = tx.Exec(ctx, queryRoles, *id, rolId)
			if err != nil {
				// Aquí podría fallar si envía un rol_id que no existe
				log.Printf("Error asignando rol %d: %v", rolId, err)
				return datatype.NewInternalServerError("Error al asignar nuevos roles")
			}
		}
	}

	// Sincronizar SUCURSALES
	// Borramos relaciones viejas
	_, err = tx.Exec(ctx, `DELETE FROM usuario_admin_sucursal WHERE usuario_admin_id = $1`, *id)
	if err != nil {
		return datatype.NewInternalServerError("Error al limpiar sucursales antiguas")
	}

	// Insertamos las nuevas
	if len(request.SucursalesIds) > 0 {
		querySucursal := `INSERT INTO usuario_admin_sucursal (usuario_admin_id, sucursal_id) VALUES ($1, $2)`
		for _, sucId := range request.SucursalesIds {
			_, err = tx.Exec(ctx, querySucursal, *id, sucId)
			if err != nil {
				log.Printf("Error asignando sucursal %d: %v", sucId, err)
				return datatype.NewInternalServerError("Error al asignar nuevas sucursales")
			}
		}
	}

	// Confirmar cambios
	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerError("Error al confirmar transacción")
	}
	committed = true
	return nil
}

func (u UsuarioAdminRepository) ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/paises/")
	var usuario domain.UsuarioAdmin
	var query = `
SELECT 
    u.id,
    u.username,
    u.password_hash,
    u.estado,
    u.creado_en,
    -- 1. ROLES
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'id', r.id,
                'nombre', r.nombre,
                'estado', r.estado,
                'creadoEn', r.creado_en
            )
        ) FILTER (WHERE r.id IS NOT NULL), '[]'
    ) AS roles,
    
    -- 2. SUCURSALES
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'id', s.id,
                'nombre', s.nombre,
                'estado', s.estado,
                'creadoEn', s.creado_en,
                'actualizadoEn', s.actualizado_en,
                'eliminadoEn', s.eliminado_en,
                'pais', jsonb_build_object(
                    'id', pa.id,
                    'nombre', pa.nombre,
                    'codigoLocal', pa.codigo_local,
                    -- CORRECCIÓN: Usar pa.id (Pais) en vez de p.id (Permiso)
                    'urlFoto', ($1::text || pa.id::text || '/' || pa.archivo), 
                    'estado', pa.estado,
                    'creadoEn', pa.creado_en
                )
            )
        ) FILTER (WHERE s.id IS NOT NULL), '[]'
    ) AS sucursales,
    
    -- 3. PERMISOS
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'id', p.id,
                'nombre', p.nombre,
                'descripcion', p.descripcion,
                'icono', p.icono,
                'creadoEn', p.creado_en
            )
        ) FILTER (WHERE p.id IS NOT NULL), '[]'
    ) AS permisos

FROM usuario_admin u
-- Joins para ROLES y PERMISOS
LEFT JOIN usuario_admin_rol uar ON u.id = uar.usuario_admin_id
LEFT JOIN rol r ON uar.rol_id = r.id AND r.estado = 'Activo'
LEFT JOIN rol_permiso rp ON r.id = rp.rol_id
LEFT JOIN permiso p ON rp.permiso_id = p.id
-- Joins para SUCURSALES y PAIS
LEFT JOIN usuario_admin_sucursal uas ON u.id = uas.usuario_admin_id
LEFT JOIN sucursal s ON uas.sucursal_id = s.id
LEFT JOIN pais pa ON s.pais_id = pa.id

WHERE u.username = $2
GROUP BY u.id
LIMIT 1;
`
	err := u.pool.QueryRow(ctx, query, fullHostname, *username).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn, &usuario.Roles, &usuario.Sucursales, &usuario.Permisos)
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
	var fullHostname string
	if val, ok := ctx.Value("fullHostname").(string); ok {
		fullHostname = fmt.Sprintf("%s%s", val, "/uploads/paises/")
	} else {
		// Valor por defecto o manejo de error si es crítico
		fullHostname = "/uploads/paises/"
	}

	var usuario domain.UsuarioAdmin
	query := `
SELECT 
    u.id,
    u.username,
    u.password_hash,
    u.estado,
    u.creado_en,
    -- 1. ROLES
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'id', r.id,
                'nombre', r.nombre,
                'estado', r.estado,
                'creadoEn', r.creado_en
            )
        ) FILTER (WHERE r.id IS NOT NULL), '[]'
    ) AS roles,
    
    -- 2. SUCURSALES
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'id', s.id,
                'nombre', s.nombre,
                'estado', s.estado,
                'creadoEn', s.creado_en,
                'actualizadoEn', s.actualizado_en,
                'eliminadoEn', s.eliminado_en,
                'pais', jsonb_build_object(
                    'id', pa.id,
                    'nombre', pa.nombre,
                    'codigoLocal', pa.codigo_local,
                    -- CORRECCIÓN: Usar pa.id (Pais) en vez de p.id (Permiso)
                    'urlFoto', ($1::text || pa.id::text || '/' || pa.archivo), 
                    'estado', pa.estado,
                    'creadoEn', pa.creado_en
                )
            )
        ) FILTER (WHERE s.id IS NOT NULL), '[]'
    ) AS sucursales,
    
    -- 3. PERMISOS
    COALESCE(
        jsonb_agg(
            DISTINCT jsonb_build_object(
                'id', p.id,
                'nombre', p.nombre,
                'descripcion', p.descripcion,
                'icono', p.icono,
                'creadoEn', p.creado_en
            )
        ) FILTER (WHERE p.id IS NOT NULL), '[]'
    ) AS permisos

FROM usuario_admin u
-- Joins para ROLES y PERMISOS
LEFT JOIN usuario_admin_rol uar ON u.id = uar.usuario_admin_id
LEFT JOIN rol r ON uar.rol_id = r.id AND r.estado = 'Activo'
LEFT JOIN rol_permiso rp ON r.id = rp.rol_id
LEFT JOIN permiso p ON rp.permiso_id = p.id
-- Joins para SUCURSALES y PAIS
LEFT JOIN usuario_admin_sucursal uas ON u.id = uas.usuario_admin_id
LEFT JOIN sucursal s ON uas.sucursal_id = s.id
LEFT JOIN pais pa ON s.pais_id = pa.id

WHERE u.id = $2
GROUP BY u.id
LIMIT 1;
`
	err := u.pool.QueryRow(ctx, query, fullHostname, *id).Scan(&usuario.Id, &usuario.Username, &usuario.PasswordHash, &usuario.Estado, &usuario.CreadoEn, &usuario.Roles, &usuario.Sucursales, &usuario.Permisos)
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
