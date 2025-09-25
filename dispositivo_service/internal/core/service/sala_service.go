package service

import (
	"context"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/port"
)

type SalaService struct {
	salaRepository port.SalaRepository
}

func (s SalaService) ObtenerSalaByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.SalaDetail, error) {
	return s.salaRepository.ObtenerSalaByDispositivoId(ctx, dispositivoId)
}

func (s SalaService) ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error) {
	return s.salaRepository.ObtenerSalaById(ctx, id)
}

func NewSalaRepository(salaRepository port.SalaRepository) *SalaService {
	return &SalaService{salaRepository: salaRepository}
}

var _ port.SalaService = (*SalaService)(nil)
