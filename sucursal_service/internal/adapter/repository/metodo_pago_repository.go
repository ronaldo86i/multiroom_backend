package repository

import (
	"context"
	"fmt"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MetodoPagoRepository struct {
	pool *pgxpool.Pool
}

func (m MetodoPagoRepository) ListarMetodosPago(ctx context.Context, filtros map[string]string) (*[]domain.MetodoPago, error) {
	var filters []string
	var args []interface{}
	var j = 1

	if estado := filtros["estado"]; estado != "" {
		filters = append(filters, fmt.Sprintf("m.estado = $%d", j))
		args = append(args, estado)
		j++
	}

	query := `SELECT m.id,m.nombre,m.estado FROM metodo_pago m`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += ` ORDER BY m.id`
	rows, err := m.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.MetodoPago, 0)
	for rows.Next() {
		var item domain.MetodoPago
		err = rows.Scan(&item.Id, &item.Nombre, &item.Estado)
		if err != nil {
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func NewMetodoPagoRepository(pool *pgxpool.Pool) *MetodoPagoRepository {
	return &MetodoPagoRepository{pool: pool}
}

var _ port.MetodoPagoRepository = (*MetodoPagoRepository)(nil)
