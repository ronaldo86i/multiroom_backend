package service

import (
	"context"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/port"
)

type RolService struct {
	rolRepository port.RolRepository
}

func (r RolService) RegistrarRol(ctx context.Context, request *domain.RolRequest) (*int, error) {
	return r.rolRepository.RegistrarRol(ctx, request)
}

func (r RolService) ModificarRolById(ctx context.Context, id *int, request *domain.RolRequest) error {
	return r.rolRepository.ModificarRolById(ctx, id, request)
}

func (r RolService) ListarRoles(ctx context.Context, filtros map[string]string) (*[]domain.RolInfo, error) {
	return r.rolRepository.ListarRoles(ctx, filtros)
}

func (r RolService) ObtenerRolById(ctx context.Context, id *int) (*domain.Rol, error) {
	return r.rolRepository.ObtenerRolById(ctx, id)
}

func NewRolService(rolRepository port.RolRepository) *RolService {
	return &RolService{rolRepository: rolRepository}
}

var _ port.RolService = (*RolService)(nil)
