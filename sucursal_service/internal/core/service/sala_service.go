package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type SalaService struct {
	salaRepository port.SalaRepository
}

func (s SalaService) ObtenerListaUsoSalas(ctx context.Context, filtros map[string]string) (*[]domain.UsoSala, error) {
	return s.salaRepository.ObtenerListaUsoSalas(ctx, filtros)
}

func (s SalaService) EliminarSalaById(ctx context.Context, id *int) error {
	return s.salaRepository.EliminarSalaById(ctx, id)
}

func (s SalaService) ActualizarUsoSalas(ctx context.Context) (*[]int, error) {
	return s.salaRepository.ActualizarUsoSalas(ctx)
}

func (s SalaService) ObtenerListaSalasDetailByIds(ctx context.Context, ids []int) (*[]domain.SalaDetail, error) {
	return s.salaRepository.ObtenerListaSalasDetailByIds(ctx, ids)
}

func (s SalaService) IncrementarTiempoUsoSala(ctx context.Context, salaId *int, request *domain.UsoSalaRequest) error {
	return s.salaRepository.IncrementarTiempoUsoSala(ctx, salaId, request)
}

func (s SalaService) CancelarSala(ctx context.Context, salaId *int) error {
	return s.salaRepository.CancelarSala(ctx, salaId)
}

func (s SalaService) AsignarTiempoUsoSala(ctx context.Context, request *domain.UsoSalaRequest) error {
	return s.salaRepository.AsignarTiempoUsoSala(ctx, request)
}

func (s SalaService) PausarTiempoUsoSala(ctx context.Context, salaId *int) error {
	return s.salaRepository.PausarTiempoUsoSala(ctx, salaId)
}

func (s SalaService) ReanudarTiempoUsoSala(ctx context.Context, salaId *int) error {
	return s.salaRepository.ReanudarTiempoUsoSala(ctx, salaId)
}

func (s SalaService) RegistrarSala(ctx context.Context, request *domain.SalaRequest) (*int, error) {
	return s.salaRepository.RegistrarSala(ctx, request)
}

func (s SalaService) ModificarSala(ctx context.Context, id *int, request *domain.SalaRequest) error {
	return s.salaRepository.ModificarSala(ctx, id, request)
}

func (s SalaService) ObtenerSalaById(ctx context.Context, id *int) (*domain.SalaDetail, error) {
	return s.salaRepository.ObtenerSalaById(ctx, id)
}

func (s SalaService) ObtenerListaSalas(ctx context.Context, filtros map[string]string) (*[]domain.SalaInfo, error) {
	return s.salaRepository.ObtenerListaSalas(ctx, filtros)
}

func (s SalaService) HabilitarSala(ctx context.Context, id *int) error {
	return s.salaRepository.HabilitarSala(ctx, id)
}

func (s SalaService) DeshabilitarSala(ctx context.Context, id *int) error {
	return s.salaRepository.DeshabilitarSala(ctx, id)
}

func NewSalaService(salaRepository port.SalaRepository) *SalaService {
	return &SalaService{salaRepository: salaRepository}
}

var _ port.SalaService = (*SalaService)(nil)
