package service

import (
	"context"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/port"
)

type UsuarioService struct {
	usuarioRepository port.UsuarioRepository
}

func (u UsuarioService) DeshabilitarUsuario(ctx context.Context, id *int) error {
	return u.usuarioRepository.DeshabilitarUsuario(ctx, id)
}

func (u UsuarioService) HabilitarUsuario(ctx context.Context, id *int) error {
	return u.usuarioRepository.HabilitarUsuario(ctx, id)
}

func (u UsuarioService) RegistrarUsuario(ctx context.Context, request *domain.UsuarioRequest) (*int, error) {
	return u.usuarioRepository.RegistrarUsuario(ctx, request)
}

func (u UsuarioService) ObtenerUsuarioById(ctx context.Context, id *int) (*domain.Usuario, error) {
	return u.usuarioRepository.ObtenerUsuarioById(ctx, id)
}

func (u UsuarioService) ObtenerListaUsuarios(ctx context.Context) (*[]domain.UsuarioInfo, error) {
	return u.usuarioRepository.ObtenerListaUsuarios(ctx)
}

func NewUsuarioService(usuarioRepository port.UsuarioRepository) *UsuarioService {
	return &UsuarioService{usuarioRepository: usuarioRepository}
}

var _ port.UsuarioService = (*UsuarioService)(nil)
