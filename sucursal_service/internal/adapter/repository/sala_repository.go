package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type SalaRepository struct {
	pool *pgxpool.Pool
}

func (s SalaRepository) ObtenerListaUsoSalas(ctx context.Context, filtros map[string]string) (*[]domain.SalaDetail, error) {
	var filters []string
	var args []interface{}
	i := 1
	if fechaInicioStr := filtros["fechaInicio"]; fechaInicioStr != "" {
		// Parsear fecha + hora + offset
		fechaInicio, err := time.Parse(time.RFC3339, fechaInicioStr)
		if err != nil {
			log.Println("Error al convertir fechaInicio a time.Time:", err)
			return nil, datatype.NewBadRequestError("El valor de fechaInicio no es válido, formato esperado: YYYY-MM-DD HH:MM:SS ±HHMM")
		}

		filters = append(filters, fmt.Sprintf("us.inicio >= $%d", i))
		args = append(args, fechaInicio.UTC())
		i++
	}

	if fechaFinStr := filtros["fechaFin"]; fechaFinStr != "" {
		// Parsear fecha + hora + offset
		fechaFin, err := time.Parse(time.RFC3339, fechaFinStr)
		if err != nil {
			log.Println("Error al convertir fechaInicio a time.Time:", err)
			return nil, datatype.NewBadRequestError("El valor de fechaInicio no es válido, formato esperado: YYYY-MM-DD HH:MM:SS ±HHMM")
		}

		filters = append(filters, fmt.Sprintf("us.fin <= $%d", i))
		args = append(args, fechaFin.UTC())
		i++
	}

	if clienteIdStr := filtros["clienteId"]; clienteIdStr != "" {
		clienteId, err := strconv.Atoi(clienteIdStr)
		if err != nil {
			log.Println("Error al convertir clienteId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de clienteId no es válido")
		}
		filters = append(filters, fmt.Sprintf("c.id = $%d", i))
		args = append(args, clienteId)
		i++
	}

	if salaIdStr := filtros["salaId"]; salaIdStr != "" {
		salaId, err := strconv.Atoi(salaIdStr)
		if err != nil {
			log.Println("Error al convertir salaId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de salaId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s.id = $%d", i))
		args = append(args, salaId)
		i++
	}

	if sucursalIdStr := filtros["sucursalId"]; sucursalIdStr != "" {
		sucursalId, err := strconv.Atoi(sucursalIdStr)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		filters = append(filters, fmt.Sprintf("s2.id = $%d", i))
		args = append(args, sucursalId)
		i++
	}

	query := `
SELECT 
    s.id,
    s.nombre,
    s.estado,
    s.creado_en,
    s.actualizado_en,
    s.eliminado_en,
    jsonb_build_object(
        'id', s2.id,
        'nombre', s2.nombre,
        'estado', s2.estado,
        'creadoEn', s2.creado_en
    ) AS sucursal,
    jsonb_build_object(
        'id', p.id,
        'nombre', p.nombre,
        'estado', p.estado,
        'creadoEn', p.creado_en
    ) AS pais,
    jsonb_build_object(
		'id', d.id,
		'dispositivoId', d.dispositivo_id,
		'nombre', d.nombre,
		'estado', d.estado,
		'creadoEn', d.creado_en,
		'usuario', COALESCE(
		jsonb_build_object(
				'id', u.id,
				'username', u.username
			), '{}'::jsonb
		)
	) AS dispositivo,
    jsonb_build_object(
		'id',us.id,
		'cliente', (
			CASE WHEN c.id IS NOT NULL THEN jsonb_build_object(
				'id', c.id,
				'nombres', c.nombres,
				'apellidos', c.apellidos,
				'codigoPais', c.codigo_pais,
				'celular', c.celular,
				'fechaNacimiento', c.fecha_nacimiento,
				'estado', c.estado,
				'creadoEn', c.creado_en
			) ELSE '{}'::jsonb END
		),
		'inicio', us.inicio,
		'fin', us.fin,
		'pausadoEn', us.pausado_en,
		'duracionPausa', EXTRACT(EPOCH FROM COALESCE(us.duracion_pausa, '0')),
		'tiempoUso', EXTRACT(EPOCH FROM (COALESCE(us.fin, NOW()) - us.inicio - COALESCE(us.duracion_pausa, '0'))),
		'estado', us.estado
    ) AS uso
FROM uso_sala us
LEFT JOIN public.sala s on s.id = us.sala_id
LEFT JOIN public.sucursal s2 on s2.id = s.sucursal_id
LEFT JOIN public.pais p on p.id = s2.pais_id
LEFT JOIN public.cliente c ON c.id = us.cliente_id
LEFT JOIN public.dispositivo d ON s.dispositivo_id = d.id
LEFT JOIN public.usuario u ON d.usuario_id = u.id
`

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY us.inicio DESC"
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al ejecutar la consulta:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.SalaDetail, 0)
	for rows.Next() {
		var sala domain.SalaDetail
		err = rows.Scan(&sala.Id, &sala.Nombre, &sala.Estado, &sala.CreadoEn, &sala.ActualizadoEn, &sala.EliminadoEn,
			&sala.Sucursal, &sala.Pais, &sala.Dispositivo, &sala.Uso)
		if err != nil {
			log.Println("Error al escanear uso_sala:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, sala)
	}

	return &list, nil
}

func (s SalaRepository) EliminarSalaById(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sala SET estado='Inactivo',actualizado_en=now(),eliminado_en=now() WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Error al modificar sala:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sala no encontrada")
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

func (s SalaRepository) ActualizarUsoSalas(ctx context.Context) (*[]int, error) {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	// Finalizar usos de sala cuyo fin ya pasó y obtener sala_id
	query := `
		UPDATE uso_sala us
		SET estado = 'Finalizado', actualizado_en = NOW()
		WHERE us.estado = 'En uso'
		  AND us.fin IS NOT NULL
		  AND us.fin <= NOW()
		RETURNING us.sala_id
	`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		log.Println("Error al ejecutar UPDATE uso_sala:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	var salas []int
	for rows.Next() {
		var salaId int
		if err := rows.Scan(&salaId); err != nil {
			log.Println("Error al escanear sala_id:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		salas = append(salas, salaId)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error en iteración de filas uso_sala:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	if len(salas) > 0 {
		// Actualizar dispositivos en un solo query usando sub consulta
		query = `
			UPDATE dispositivo d
			SET estado = 'Inactivo'
			FROM sala s
			WHERE d.id = s.dispositivo_id
			  AND s.id = ANY($1::int[])
		`
		if _, err := tx.Exec(ctx, query, pq.Array(salas)); err != nil {
			log.Println("Error al actualizar dispositivos:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
	}

	// Confirmar de transacción
	if err = tx.Commit(ctx); err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	return &salas, nil
}

func (s SalaRepository) ObtenerListaSalasDetailByIds(ctx context.Context, ids []int) (*[]domain.SalaDetail, error) {
	query := `
SELECT 
    s.id,
    s.nombre,
    s.estado,
    s.creado_en,
    s.actualizado_en,
    s.eliminado_en,
    jsonb_build_object(
        'id', s2.id,
        'nombre', s2.nombre,
        'estado', s2.estado,
        'creadoEn', s2.creado_en
    ) AS sucursal,
    jsonb_build_object(
        'id', p.id,
        'nombre', p.nombre,
        'estado', p.estado,
        'creadoEn', p.creado_en
    ) AS pais,
    jsonb_build_object(
		'id', d.id,
		'dispositivoId', d.dispositivo_id,
		'nombre', d.nombre,
		'estado', d.estado,
		'creadoEn', d.creado_en,
		'usuario', COALESCE(
		jsonb_build_object(
				'id', u.id,
				'username', u.username
			), '{}'::jsonb
		)
	) AS dispositivo,
    (
        CASE WHEN  (us.estado != 'Finalizado' AND us.estado!= 'Cancelado') THEN
            jsonb_build_object(
                'cliente', (
                    CASE WHEN c.id IS NOT NULL THEN jsonb_build_object(
                        'id', c.id,
                        'nombres', c.nombres,
                        'apellidos', c.apellidos,
                        'codigoPais', c.codigo_pais,
                        'celular', c.celular,
                        'fechaNacimiento', c.fecha_nacimiento,
                        'estado', c.estado,
                        'creadoEn', c.creado_en
                    ) ELSE '{}'::jsonb END
                ),
				'id', us.id,
                'inicio', us.inicio,
                'fin', us.fin,
                'pausadoEn', us.pausado_en,
                'duracionPausa', EXTRACT(EPOCH FROM COALESCE(us.duracion_pausa, '0')),
                'tiempoUso', EXTRACT(EPOCH FROM (COALESCE(us.fin, NOW()) - us.inicio - COALESCE(us.duracion_pausa, '0'))),
                'estado', us.estado
            )
        ELSE 'null'::jsonb END
    ) AS uso
FROM sala s
LEFT JOIN public.sucursal s2 ON s2.id = s.sucursal_id
LEFT JOIN public.pais p ON p.id = s2.pais_id
LEFT JOIN LATERAL (
    SELECT * 
    FROM uso_sala 
    WHERE sala_id = s.id 
    ORDER BY inicio DESC
    LIMIT 1
) us ON true
LEFT JOIN public.cliente c ON c.id = us.cliente_id
LEFT JOIN public.dispositivo d ON s.dispositivo_id = d.id
LEFT JOIN public.usuario u ON d.usuario_id = u.id
WHERE s.id = ANY($1::int[])
`
	rows, err := s.pool.Query(ctx, query, pq.Array(ids))
	if err != nil {
		log.Println("Error al obtener lista:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var lista = make([]domain.SalaDetail, 0)
	defer rows.Close()
	for rows.Next() {
		var sala domain.SalaDetail
		err = rows.Scan(&sala.Id, &sala.Nombre, &sala.Estado, &sala.CreadoEn, &sala.ActualizadoEn, &sala.EliminadoEn, &sala.Sucursal, &sala.Pais, &sala.Dispositivo, &sala.Uso)
		if err != nil {
			log.Println("Error al obtener lista:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		lista = append(lista, sala)
	}
	return &lista, nil
}

func (s SalaRepository) ReanudarTiempoUsoSala(ctx context.Context, salaId *int) error {
	tx, err := s.pool.Begin(ctx)
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

	// Actualizar sesión pausada
	query := `
	UPDATE uso_sala
	SET estado = 'En uso',
		duracion_pausa = duracion_pausa + (NOW() - pausado_en),
		fin = fin + (NOW() - pausado_en),
		pausado_en = NULL,
		actualizado_en = NOW()
	WHERE sala_id = $1
	  AND estado = 'Pausado';
    `

	cmdTag, err := tx.Exec(ctx, query, *salaId)
	if err != nil {
		log.Println("Error al reanudar sala:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if cmdTag.RowsAffected() == 0 {
		return datatype.NewBadRequestError("No se pudo reanudar: La sala no está pausada")
	}
	var dispositivoId int
	query = `SELECT s.dispositivo_id FROM sala s WHERE s.id =$1 LIMIT 1`
	err = tx.QueryRow(ctx, query, *salaId).Scan(&dispositivoId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return datatype.NewNotFoundError("Dispositivo no encontrado de la sala")
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	query = `UPDATE dispositivo SET estado='Activo' WHERE id=$1`
	_, err = tx.Exec(ctx, query, dispositivoId)
	if err != nil {
		log.Println("Error al actualizar dispositivo:", err)
		return datatype.NewInternalServerErrorGeneric()
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

func (s SalaRepository) PausarTiempoUsoSala(ctx context.Context, salaId *int) error {
	tx, err := s.pool.Begin(ctx)
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

	query := `
	UPDATE uso_sala
	SET estado = 'Pausado',
		pausado_en = NOW(),
		actualizado_en = NOW()
	WHERE sala_id = $1
	  AND estado = 'En uso'
	  AND pausado_en IS NULL;
    `

	cmdTag, err := tx.Exec(ctx, query, *salaId)
	if err != nil {
		log.Println("Error al pausar sala:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if cmdTag.RowsAffected() == 0 {
		return datatype.NewBadRequestError("No se pudo pausar, la sala no está en uso o no existe")
	}

	var dispositivoId int
	query = `SELECT s.dispositivo_id FROM sala s WHERE s.id =$1 LIMIT 1`
	err = tx.QueryRow(ctx, query, *salaId).Scan(&dispositivoId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return datatype.NewNotFoundError("Dispositivo no encontrado de la sala")
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	query = `UPDATE dispositivo SET estado='Inactivo' WHERE id=$1`
	_, err = tx.Exec(ctx, query, dispositivoId)
	if err != nil {
		log.Println("Error al actualizar dispositivo:", err)
		return datatype.NewInternalServerErrorGeneric()
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

func (s SalaRepository) IncrementarTiempoUsoSala(ctx context.Context, salaId *int, request *domain.UsoSalaRequest) error {
	tx, err := s.pool.Begin(ctx)
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

	// Incrementar o reducir tiempo (TiempoUso en minutos)
	query := `
	UPDATE uso_sala
	SET fin = GREATEST(
				inicio,
				COALESCE(fin, inicio) + ($1 * INTERVAL '1 second')
			),	
		actualizado_en = NOW()
	WHERE sala_id = $2
	  AND estado = 'En uso';
    `

	cmdTag, err := tx.Exec(ctx, query, request.TiempoUso, *salaId)
	if err != nil {
		log.Println("Error al incrementar tiempo de uso:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if cmdTag.RowsAffected() == 0 {
		return datatype.NewBadRequestError("No se pudo incrementar tiempo: sala no existe o no está en uso")
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

func (s SalaRepository) CancelarSala(ctx context.Context, salaId *int) error {
	tx, err := s.pool.Begin(ctx)
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

	// Actualizar el estado a 'Cancelado' solo si no está finalizado
	query := `
        UPDATE uso_sala
        SET estado = 'Cancelado',
            actualizado_en = NOW()
        WHERE sala_id = $1
          AND estado IN ('En uso','Pausado')
    `

	cmdTag, err := tx.Exec(ctx, query, *salaId)
	if err != nil {
		log.Println("Error al cancelar uso de sala:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if cmdTag.RowsAffected() == 0 {
		return datatype.NewBadRequestError("No se pudo cancelar la sala: ya está finalizada o no existe")
	}

	var dispositivoId int
	query = `SELECT s.dispositivo_id FROM sala s WHERE s.id =$1 LIMIT 1`
	err = tx.QueryRow(ctx, query, *salaId).Scan(&dispositivoId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return datatype.NewNotFoundError("Dispositivo no encontrado de la sala")
		}
		return datatype.NewInternalServerErrorGeneric()
	}
	query = `UPDATE dispositivo SET estado='Inactivo' WHERE id=$1`
	_, err = tx.Exec(ctx, query, dispositivoId)
	if err != nil {
		log.Println("Error al actualizar dispositivo:", err)
		return datatype.NewInternalServerErrorGeneric()
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

func (s SalaRepository) AsignarTiempoUsoSala(ctx context.Context, request *domain.UsoSalaRequest) (*int64, error) {
	tx, err := s.pool.Begin(ctx)
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

	// Verificar si la sala ya está en uso
	var exists int
	checkQuery := `SELECT 1 FROM uso_sala WHERE sala_id=$1 AND estado='En uso' LIMIT 1`
	err = tx.QueryRow(ctx, checkQuery, request.SalaId).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Println("Error al verificar uso de sala:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	if exists == 1 {
		log.Println("La sala está en uso")
		return nil, datatype.NewBadRequestError("La sala ya se encuentra en uso")
	}

	// Insertar el nuevo uso
	insertQuery := `
INSERT INTO uso_sala(sala_id,cliente_id,inicio,fin,estado,pausado_en,duracion_pausa)
VALUES ($1, $2, NOW(), NOW() + ($3 * INTERVAL '1 second'), 'En uso', NULL, INTERVAL '0 second')
RETURNING id`

	var usoId int64
	err = tx.QueryRow(ctx, insertQuery, request.SalaId, request.ClienteId, request.TiempoUso).Scan(&usoId)
	if err != nil {
		log.Println("Error al insertar uso_sala:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	var dispositivoId int
	query := `SELECT s.dispositivo_id FROM sala s WHERE s.id =$1 LIMIT 1`
	err = tx.QueryRow(ctx, query, request.SalaId).Scan(&dispositivoId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Dispositivo no encontrado de la sala")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	query = `UPDATE dispositivo SET estado='Activo' WHERE id=$1`
	_, err = tx.Exec(ctx, query, dispositivoId)
	if err != nil {
		log.Println("Error al actualizar dispositivo:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	committed = true
	return &usoId, nil
}

func (s SalaRepository) RegistrarSala(ctx context.Context, request *domain.SalaRequest) (*int, error) {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `INSERT INTO sala(nombre, sucursal_id, dispositivo_id) 
	          VALUES ($1, $2, $3) 
	          RETURNING id`

	var id int
	err = tx.QueryRow(ctx, query, request.Nombre, request.SucursalId, request.DispositivoId).Scan(&id)
	if err != nil {
		log.Println("Error en registrar sala:", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				switch pgErr.ConstraintName {
				case "unique_nombre_sala":
					return nil, datatype.NewConflictError("Ya existe una sala con ese nombre en esa sucursal")
				case "unique_dispositivo_sala":
					return nil, datatype.NewConflictError("El dispositivo ya está asignado a una sala en esta sucursal")
				}
			case "23503": // foreign_key_violation
				switch pgErr.ConstraintName {
				case "sala_sucursal_id_fkey":
					return nil, datatype.NewBadRequestError("La sucursal especificada no existe")
				case "sala_dispositivo_id_fkey":
					return nil, datatype.NewBadRequestError("El dispositivo especificado no existe")
				}
			}
		}
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

func (s SalaRepository) ModificarSala(ctx context.Context, id *int, request *domain.SalaRequest) error {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sala SET nombre=$1,sucursal_id=$2,dispositivo_id=$3 WHERE id=$4`
	ct, err := tx.Exec(ctx, query, request.Nombre, request.SucursalId, request.DispositivoId, *id)
	if err != nil {
		log.Println("Error al modificar sala:", err)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505": // unique_violation
					if pgErr.ConstraintName == "unique_nombre_sala" {
						return datatype.NewConflictError("Ya existe una sala con ese nombre en esa sucursal")
					}
				case "23503": // foreign_key_violation
					if pgErr.ConstraintName == "sala_sucursal_id_fkey" {
						return datatype.NewBadRequestError("La sucursal especificada no existe")
					}
				}

			}
			return datatype.NewInternalServerErrorGeneric()
		}
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sala no encontrada")
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

func (s SalaRepository) ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error) {
	query := `
SELECT 
    s.id,
    s.nombre,
    s.estado,
    s.creado_en,
    s.actualizado_en,
    s.eliminado_en,
    jsonb_build_object(
        'id', s2.id,
        'nombre', s2.nombre,
        'estado', s2.estado,
        'creadoEn', s2.creado_en
    ) AS sucursal,
    jsonb_build_object(
        'id', p.id,
        'nombre', p.nombre,
        'estado', p.estado,
        'creadoEn', p.creado_en
    ) AS pais,
	jsonb_build_object(
		'id', d.id,
		'dispositivoId', d.dispositivo_id,
		'nombre', d.nombre,
		'estado', d.estado,
		'creadoEn', d.creado_en,
		'enLinea',d.en_linea,
		'usuario', COALESCE(
			jsonb_build_object(
				'id', u.id,
				'username', u.username
			), '{}'::jsonb
		)
	) AS dispositivo,
    (
        CASE WHEN  (us.estado != 'Finalizado' AND us.estado!= 'Cancelado') THEN
            jsonb_build_object(
				'id',us.id,
                'cliente', COALESCE(
                    jsonb_build_object(
                        'id', c.id,
                        'nombres', c.nombres,
                        'apellidos', c.apellidos,
                        'codigoPais', c.codigo_pais,
                        'celular', c.celular,
                        'fechaNacimiento', c.fecha_nacimiento,
                        'estado', c.estado,
                        'creadoEn', c.creado_en
                    ), '{}'::jsonb
                ),
                'inicio', us.inicio,
                'fin', us.fin,
                'pausadoEn', us.pausado_en,
                'duracionPausa', EXTRACT(EPOCH FROM COALESCE(us.duracion_pausa, '0')),
                'tiempoUso', EXTRACT(EPOCH FROM (COALESCE(us.fin, NOW()) - us.inicio - COALESCE(us.duracion_pausa, '0'))),
                'estado', us.estado
            )
        ELSE 'null'::jsonb
    END) AS uso,
    s.actualizado_en,
    s.eliminado_en
FROM sala s
LEFT JOIN public.sucursal s2 ON s2.id = s.sucursal_id
LEFT JOIN public.pais p ON p.id = s2.pais_id
LEFT JOIN LATERAL (
    SELECT * 
    FROM uso_sala 
    WHERE sala_id = s.id 
    ORDER BY inicio DESC 
    LIMIT 1
) us ON true
LEFT JOIN public.cliente c ON c.id = us.cliente_id
LEFT JOIN public.dispositivo d ON s.dispositivo_id = d.id
LEFT JOIN public.usuario u ON d.usuario_id = u.id
WHERE s.id = $1
LIMIT 1
`
	var sala domain.SalaDetail
	err := s.pool.QueryRow(ctx, query, *id).
		Scan(&sala.Id, &sala.Nombre, &sala.Estado, &sala.CreadoEn, &sala.ActualizadoEn, &sala.EliminadoEn, &sala.Sucursal,
			&sala.Pais, &sala.Dispositivo, &sala.Uso, &sala.ActualizadoEn, &sala.EliminadoEn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Sala no encontrada")
		}
	}
	return &sala, nil
}

func (s SalaRepository) ObtenerListaSalas(ctx context.Context, filtros map[string]string) (*[]domain.SalaInfo, error) {
	var filters []string
	var args []interface{}
	i := 1
	filters = append(filters, "s.eliminado_en IS NULL")
	// Si hay sucursalId en filtros
	if sucursalID := filtros["sucursalId"]; sucursalID != "" {
		filters = append(filters, fmt.Sprintf("s.sucursal_id = $%d", i))
		args = append(args, sucursalID)
		i++
	}

	//Si hay nombre en filtros
	if estado := filtros["estado"]; estado != "" {
		filters = append(filters, fmt.Sprintf("s.estado = $%d", i))
		args = append(args, estado)
		i++
	}

	query := `
SELECT 
    s.id,
    s.nombre,
    s.estado,
    s.creado_en,
    jsonb_build_object(
		'id', d.id,
		'dispositivoId', d.dispositivo_id,
		'nombre', d.nombre,
		'estado', d.estado,
		'creadoEn', d.creado_en,
		'enLinea',d.en_linea,
		'usuario', COALESCE(
		jsonb_build_object(
				'id', u.id,
				'username', u.username
			), '{}'::jsonb
		)
	) AS dispositivo,
    (
        CASE 
            WHEN (us.estado != 'Finalizado' AND us.estado!= 'Cancelado') THEN 
                json_build_object(
					'id',us.id,
                    'cliente', (
                        CASE WHEN c.id IS NOT NULL THEN json_build_object(
                            'id', c.id,
                            'nombres', c.nombres,
                            'apellidos', c.apellidos,
                            'codigoPais', c.codigo_pais,
                            'celular', c.celular,
                            'fechaNacimiento', c.fecha_nacimiento,
                            'estado', c.estado,
                            'creadoEn', c.creado_en
                        ) ELSE NULL END
                    ),
                    'inicio', us.inicio,
                    'fin', us.fin,
                    'pausadoEn', us.pausado_en,
                	'duracionPausa', EXTRACT(EPOCH FROM COALESCE(us.duracion_pausa, '0')),
                	'tiempoUso', EXTRACT(EPOCH FROM (COALESCE(us.fin, NOW()) - us.inicio - COALESCE(us.duracion_pausa, '0'))),
                    'estado', us.estado
                ) 
            ELSE NULL
        END
    ) AS uso,
    s.actualizado_en,
    s.eliminado_en
FROM sala s
LEFT JOIN LATERAL (
    SELECT *
    FROM uso_sala us
    WHERE us.sala_id = s.id
    ORDER BY us.inicio DESC
    LIMIT 1
) us ON TRUE
LEFT JOIN cliente c ON c.id = us.cliente_id
LEFT JOIN dispositivo d ON s.dispositivo_id = d.id
LEFT JOIN usuario u ON d.usuario_id = u.id
	`

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " ORDER BY s.id"

	list := make([]domain.SalaInfo, 0)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		log.Println("Error al obtener lista de salas:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()

	for rows.Next() {
		var sala domain.SalaInfo
		if err = rows.Scan(&sala.Id, &sala.Nombre, &sala.Estado, &sala.CreadoEn, &sala.Dispositivo, &sala.Uso, &sala.ActualizadoEn, &sala.EliminadoEn); err != nil {
			log.Println("Error al escanear sala:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, sala)
	}

	return &list, nil
}

func (s SalaRepository) HabilitarSala(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sala SET estado='Activo',actualizado_en=now() WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Error al modificar sala:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sala no encontrada")
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

func (s SalaRepository) DeshabilitarSala(ctx context.Context, id *int) error {
	// Iniciar transacción
	tx, err := s.pool.Begin(ctx)
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

	query := `UPDATE sala SET estado='Inactivo',actualizado_en=now() WHERE id=$1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		log.Println("Error al modificar sala:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Sala no encontrada")
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

func NewSalaRepository(pool *pgxpool.Pool) *SalaRepository {
	return &SalaRepository{pool: pool}
}

var _ port.SalaRepository = (*SalaRepository)(nil)
