package service

import (
	"context"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/port"
)

type UsuarioAdminService struct {
	usuarioAdminRepository port.UsuarioAdminRepository
}

func (u UsuarioAdminService) ObtenerUsuarioAdminByUsername(ctx context.Context, username *string) (*domain.UsuarioAdmin, error) {
	return u.usuarioAdminRepository.ObtenerUsuarioAdminByUsername(ctx, username)
}

func (u UsuarioAdminService) ObtenerUsuarioAdminById(ctx context.Context, id *int) (*domain.UsuarioAdmin, error) {
	return u.usuarioAdminRepository.ObtenerUsuarioAdminById(ctx, id)
}

func NewUsuarioAdminService(usuarioAdminRepository port.UsuarioAdminRepository) *UsuarioAdminService {
	return &UsuarioAdminService{usuarioAdminRepository: usuarioAdminRepository}
}

var _ port.UsuarioAdminService = (*UsuarioAdminService)(nil)
