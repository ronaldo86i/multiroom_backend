package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type MetodoPagoService struct {
	metodoPagoRepository port.MetodoPagoRepository
}

func (m MetodoPagoService) ListarMetodosPago(ctx context.Context, filtros map[string]string) (*[]domain.MetodoPago, error) {
	return m.metodoPagoRepository.ListarMetodosPago(ctx, filtros)
}

func NewMetodoPagoService(metodoPagoRepository port.MetodoPagoRepository) *MetodoPagoService {
	return &MetodoPagoService{metodoPagoRepository: metodoPagoRepository}
}

var _ port.MetodoPagoService = (*MetodoPagoService)(nil)
