package port

import (
	"context"
	"multiroom/dispositivo-service/internal/core/domain"
)

type SalaRepository interface {
	ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error)
	ObtenerSalaByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.SalaDetail, error)
}

type SalaService interface {
	ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error)
	ObtenerSalaByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.SalaDetail, error)
}
