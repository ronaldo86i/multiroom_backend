package repository

import (
	"context"
	"errors"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/domain/datatype"
	"multiroom/dispositivo-service/internal/core/port"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SalaRepository struct {
	pool *pgxpool.Pool
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

func (s SalaRepository) ObtenerSalaByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.SalaDetail, error) {
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
        'enLinea', d.en_linea,
        'usuario', COALESCE(
            jsonb_build_object(
                'id', u.id,
                'username', u.username
            ), '{}'::jsonb
        )
    ) AS dispositivo,
    (
        CASE WHEN (us.estado != 'Finalizado' AND us.estado != 'Cancelado') THEN
            jsonb_build_object(
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
WHERE d.dispositivo_id = $1 AND d.eliminado_en IS NULL
LIMIT 1
`
	var sala domain.SalaDetail
	err := s.pool.QueryRow(ctx, query, *dispositivoId).
		Scan(&sala.Id, &sala.Nombre, &sala.Estado, &sala.CreadoEn, &sala.ActualizadoEn, &sala.EliminadoEn,
			&sala.Sucursal, &sala.Pais, &sala.Dispositivo, &sala.Uso, &sala.ActualizadoEn, &sala.EliminadoEn)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	return &sala, nil
}

func NewSalaRepository(pool *pgxpool.Pool) *SalaRepository {
	return &SalaRepository{pool: pool}
}

var _ port.SalaRepository = (*SalaRepository)(nil)
