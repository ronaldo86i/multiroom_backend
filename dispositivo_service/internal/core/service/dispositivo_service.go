package service

import (
	"context"
	"multiroom/dispositivo-service/internal/core/domain"
	"multiroom/dispositivo-service/internal/core/port"
)

type DispositivoService struct {
	dispositivoRepository port.DispositivoRepository
}

func (d DispositivoService) ObtenerDispositivoByDispositivoId(ctx context.Context, dispositivoId *string) (*domain.DispositivoInfo, error) {
	return d.dispositivoRepository.ObtenerDispositivoByDispositivoId(ctx, dispositivoId)
}

func (d DispositivoService) ObtenerDispositivoById(ctx context.Context, id *int) (*domain.DispositivoInfo, error) {
	return d.dispositivoRepository.ObtenerDispositivoById(ctx, id)
}

func (d DispositivoService) DeshabilitarDispositivo(ctx context.Context, id *int) error {
	err := d.dispositivoRepository.DeshabilitarDispositivo(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (d DispositivoService) HabilitarDispositivo(ctx context.Context, id *int) error {
	err := d.dispositivoRepository.HabilitarDispositivo(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (d DispositivoService) ObtenerListaDispositivosByUsuarioId(ctx context.Context, usuarioId *int) (*[]domain.DispositivoInfo, error) {
	return d.dispositivoRepository.ObtenerListaDispositivosByUsuarioId(ctx, usuarioId)
}

func (d DispositivoService) RegistrarDispositivo(ctx context.Context, request *domain.DispositivoRequest) error {
	return d.dispositivoRepository.RegistrarDispositivo(ctx, request)
}

func (d DispositivoService) ObtenerListaDispositivos(ctx context.Context, filtros map[string]string) (*[]domain.DispositivoInfo, error) {
	return d.dispositivoRepository.ObtenerListaDispositivos(ctx, filtros)
}

func NewDispositivoService(dispositivoRepository port.DispositivoRepository) *DispositivoService {
	return &DispositivoService{dispositivoRepository: dispositivoRepository}
}

var _ port.DispositivoService = (*DispositivoService)(nil)
