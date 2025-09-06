package service

import (
	"context"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/port"
)

type UsuarioSucursalService struct {
	usuarioSucursalRepository port.UsuarioSucursalRepository
}

func (u UsuarioSucursalService) ObtenerUsuarioSucursalByUsername(ctx context.Context, username *string) (*domain.UsuarioSucursal, error) {
	return u.usuarioSucursalRepository.ObtenerUsuarioSucursalByUsername(ctx, username)
}

func (u UsuarioSucursalService) RegistrarUsuarioSucursal(ctx context.Context, request *domain.UsuarioSucursalRequest) (*int, error) {
	return u.usuarioSucursalRepository.RegistrarUsuarioSucursal(ctx, request)
}

func (u UsuarioSucursalService) ModificarUsuarioSucursal(ctx context.Context, id *int, request *domain.UsuarioSucursalRequest) error {
	return u.usuarioSucursalRepository.ModificarUsuarioSucursal(ctx, id, request)
}

func (u UsuarioSucursalService) ObtenerListaUsuariosSucursal(ctx context.Context, filtros map[string]string) (*[]domain.UsuarioSucursalInfo, error) {
	return u.usuarioSucursalRepository.ObtenerListaUsuariosSucursal(ctx, filtros)
}

func (u UsuarioSucursalService) ObtenerUsuarioSucursalById(ctx context.Context, id *int) (*domain.UsuarioSucursal, error) {
	return u.usuarioSucursalRepository.ObtenerUsuarioSucursalById(ctx, id)
}

func (u UsuarioSucursalService) HabilitarUsuarioSucursal(ctx context.Context, id *int) error {
	return u.usuarioSucursalRepository.HabilitarUsuarioSucursal(ctx, id)
}

func (u UsuarioSucursalService) DeshabilitarUsuarioSucursal(ctx context.Context, id *int) error {
	return u.usuarioSucursalRepository.DeshabilitarUsuarioSucursal(ctx, id)
}

func NewUsuarioSucursalService(usuarioSucursalRepository port.UsuarioSucursalRepository) *UsuarioSucursalService {
	return &UsuarioSucursalService{usuarioSucursalRepository: usuarioSucursalRepository}
}

var _ port.UsuarioSucursalService = (*UsuarioSucursalService)(nil)
