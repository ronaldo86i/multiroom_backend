package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type UbicacionService struct {
	ubicacionService port.UbicacionService
}

func (u UbicacionService) HabilitarUbicacion(ctx context.Context, id *int) error {
	return u.ubicacionService.HabilitarUbicacion(ctx, id)
}

func (u UbicacionService) DeshabilitarUbicacion(ctx context.Context, id *int) error {
	return u.ubicacionService.DeshabilitarUbicacion(ctx, id)
}

func (u UbicacionService) RegistrarUbicacion(ctx context.Context, request *domain.UbicacionRequest) (*int, error) {
	return u.ubicacionService.RegistrarUbicacion(ctx, request)
}

func (u UbicacionService) ModificarUbicacionById(ctx context.Context, id *int, request *domain.UbicacionRequest) error {
	return u.ubicacionService.ModificarUbicacionById(ctx, id, request)
}

func (u UbicacionService) ListarUbicaciones(ctx context.Context, filtros map[string]string) (*[]domain.Ubicacion, error) {
	return u.ubicacionService.ListarUbicaciones(ctx, filtros)
}

func (u UbicacionService) ObtenerUbicacionById(ctx context.Context, id *int) (*domain.Ubicacion, error) {
	return u.ubicacionService.ObtenerUbicacionById(ctx, id)
}

func NewUbicacionService(ubicacionService port.UbicacionService) *UbicacionService {
	return &UbicacionService{ubicacionService: ubicacionService}
}

var _ port.UbicacionService = (*UbicacionService)(nil)
