package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type SucursalService struct {
	sucursalRepository port.SucursalRepository
}

func (s SucursalService) RegistrarSucursal(ctx context.Context, request *domain.SucursalRequest) (*int, error) {
	return s.sucursalRepository.RegistrarSucursal(ctx, request)
}

func (s SucursalService) ModificarSucursal(ctx context.Context, id *int, request *domain.SucursalRequest) error {
	return s.sucursalRepository.ModificarSucursal(ctx, id, request)
}

func (s SucursalService) ObtenerSucursalById(ctx context.Context, id *int) (*domain.SucursalDetail, error) {
	return s.sucursalRepository.ObtenerSucursalById(ctx, id)
}

func (s SucursalService) ObtenerListaSucursales(ctx context.Context, filtros map[string]string) (*[]domain.SucursalInfo, error) {
	return s.sucursalRepository.ObtenerListaSucursales(ctx, filtros)
}

func (s SucursalService) HabilitarSucursal(ctx context.Context, id *int) error {
	return s.sucursalRepository.HabilitarSucursal(ctx, id)
}

func (s SucursalService) DeshabilitarSucursal(ctx context.Context, id *int) error {
	return s.sucursalRepository.DeshabilitarSucursal(ctx, id)
}

func NewSucursalService(sucursalRepository port.SucursalRepository) *SucursalService {
	return &SucursalService{sucursalRepository: sucursalRepository}
}

var _ port.SucursalService = (*SucursalService)(nil)
