package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/domain/datatype"
	"multiroom/dispositivo-service/internal/core/port"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DispositivoRepository struct {
	pool *pgxpool.Pool
}

func (d DispositivoRepository) ActualizarDispositivoEnLinea(ctx context.Context, id *int, enLinea *bool) error {
	tx, err := d.pool.Begin(ctx)
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

	query := `UPDATE dispositivo d SET en_linea=$1 WHERE id = $2`
	ct, err := tx.Exec(ctx, query, *enLinea, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Dispositivo no encontrado")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (d DispositivoRepository) EliminarDispositivoById(ctx context.Context, id *int) error {
	tx, err := d.pool.Begin(ctx)
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

	query := `UPDATE dispositivo d SET estado = 'Inactivo', eliminado_en = now() WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Dispositivo no encontrado")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (d DispositivoRepository) ObtenerDispositivoByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.DispositivoInfo, error) {
	query := `
SELECT 
    d.id,
    d.dispositivo_id,
    d.nombre,
    d.estado,
    d.creado_en,
    json_build_object(
    'id',u.id,
    'username',u.username
    ) as usuario 
FROM dispositivo d
    LEFT JOIN public.usuario u on u.id = d.usuario_id
WHERE d.dispositivo_id = $1 AND d.eliminado_en IS NULL
LIMIT 1
`
	var item domain.DispositivoInfo
	err := d.pool.QueryRow(ctx, query, *dispositivoId).Scan(&item.Id, &item.DispositivoId, &item.Nombre, &item.Estado, &item.CreadoEn, &item.Usuario)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Dispositivo no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func (d DispositivoRepository) ObtenerDispositivoById(ctx context.Context, id *int) (*domain.DispositivoInfo, error) {
	query := `
SELECT 
    d.id,
    d.dispositivo_id,
    d.nombre,
    d.estado,
    d.creado_en,
    json_build_object(
    'id',u.id,
    'username',u.username
    ) as usuario 
FROM dispositivo d
    LEFT JOIN public.usuario u on u.id = d.usuario_id
WHERE d.id = $1
LIMIT 1
`
	var item domain.DispositivoInfo
	err := d.pool.QueryRow(ctx, query, *id).Scan(&item.Id, &item.DispositivoId, &item.Nombre, &item.Estado, &item.CreadoEn, &item.Usuario)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Dispositivo no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func (d DispositivoRepository) DeshabilitarDispositivo(ctx context.Context, id *int) error {
	tx, err := d.pool.Begin(ctx)
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

	query := `UPDATE dispositivo d SET estado = 'Inactivo' WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Dispositivo no encontrado")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (d DispositivoRepository) HabilitarDispositivo(ctx context.Context, id *int) error {
	tx, err := d.pool.Begin(ctx)
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
	query := `UPDATE dispositivo d SET estado = 'Activo' WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Dispositivo no encontrado")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (d DispositivoRepository) ObtenerListaDispositivosByUsuarioId(ctx context.Context, usuarioId *int) (*[]domain.DispositivoInfo, error) {
	query := `
SELECT 
    d.id,
    d.dispositivo_id,
    d.nombre,
    d.estado,
    d.creado_en,
    json_build_object(
    'id',u.id,
    'username',u.username
    ) as usuario 
FROM dispositivo d
    LEFT JOIN public.usuario u on u.id = d.usuario_id
WHERE d.usuario_id = $1
`

	rows, err := d.pool.Query(ctx, query, *usuarioId)
	if err != nil {
		log.Printf("Error al obtener el dispositivos: %s", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	var list = make([]domain.DispositivoInfo, 0)
	for rows.Next() {
		var item domain.DispositivoInfo
		err = rows.Scan(&item.Id, &item.DispositivoId, &item.Nombre, &item.Estado, &item.CreadoEn, &item.Usuario)
		if err != nil {
			log.Printf("Error al obtener el dispositivos: %s", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (d DispositivoRepository) RegistrarDispositivo(ctx context.Context, request *domain.DispositivoRequest) error {
	// Iniciar transacción
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		log.Printf("Error al iniciar transacción: %s", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Insertar dispositivo del usuario
	query := `INSERT INTO dispositivo(dispositivo_id, nombre, usuario_id) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, query, request.DispositivoId, request.Nombre, request.UsuarioId)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return datatype.NewBadRequestError("El dispositivo ya se encuentra registrado")
			case "23514": // check_violation
				return datatype.NewBadRequestError(fmt.Sprintf("Violación de regla de negocio: %s", pgErr.ConstraintName))
			case "23502": // not_null_violation
				return datatype.NewBadRequestError(fmt.Sprintf("Falta un campo obligatorio: %s", pgErr.ColumnName))
			case "22P02": // invalid_text_representation (por ejemplo, error al insertar texto donde se espera número)
				return datatype.NewBadRequestError("Tipo de dato inválido")
			}
		}
	}

	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return nil
}

func (d DispositivoRepository) ObtenerListaDispositivos(ctx context.Context, filtros map[string]string) (*[]domain.DispositivoInfo, error) {
	// Extraer y validar parámetros de filtros
	dispositivoId := strings.TrimSpace(filtros["dispositivoId"])
	estado := strings.TrimSpace(filtros["estado"])

	// Construir la query dinámicamente para mejor rendimiento
	query := `
        SELECT 
            d.id,
            d.dispositivo_id,
            d.nombre,
            d.estado,
            d.creado_en,
            COALESCE(
                json_build_object(
                    'id', u.id,
                    'username', u.username
                ),
                '{}'::json
            ) as usuario 
        FROM dispositivo d
        LEFT JOIN usuario u ON u.id = d.usuario_id`

	// Construir condiciones WHERE dinámicamente
	var conditions []string
	var args []interface{}
	argIndex := 1
	conditions = append(conditions, "d.eliminado_en IS NULL")

	if dispositivoId != "" {
		conditions = append(conditions, fmt.Sprintf("d.dispositivo_id = $%d", argIndex))
		args = append(args, dispositivoId)
		argIndex++
	}

	if estado != "" {
		conditions = append(conditions, fmt.Sprintf("d.estado = $%d", argIndex))
		args = append(args, estado)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY d.usuario_id, d.creado_en DESC"

	// Ejecutar la consulta
	rows, err := d.pool.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Error al obtener dispositivos: %v", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	// Procesar resultados
	var dispositivos = make([]domain.DispositivoInfo, 0)
	for rows.Next() {
		var dispositivo domain.DispositivoInfo
		if err := rows.Scan(
			&dispositivo.Id,
			&dispositivo.DispositivoId,
			&dispositivo.Nombre,
			&dispositivo.Estado,
			&dispositivo.CreadoEn,
			&dispositivo.Usuario,
		); err != nil {
			log.Printf("Error al escanear dispositivo: %v", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		dispositivos = append(dispositivos, dispositivo)
	}

	// Verificar errores de iteración
	if err := rows.Err(); err != nil {
		log.Printf("Error durante la iteración de filas: %v", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	return &dispositivos, nil
}

func NewDispositivoRepository(pool *pgxpool.Pool) *DispositivoRepository {
	return &DispositivoRepository{pool: pool}
}

var _ port.DispositivoRepository = (*DispositivoRepository)(nil)
