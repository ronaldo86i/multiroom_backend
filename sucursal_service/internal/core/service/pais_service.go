package service

import (
	"context"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type PaisService struct {
	paisRepository port.PaisRepository
}

func (p PaisService) HabilitarPaisById(ctx context.Context, id *int) error {
	return p.paisRepository.HabilitarPaisById(ctx, id)
}

func (p PaisService) DeshabilitarPaisById(ctx context.Context, id *int) error {
	return p.paisRepository.DeshabilitarPaisById(ctx, id)
}

func (p PaisService) RegistrarPais(ctx context.Context, request *domain.PaisRequest, fileHeader *multipart.FileHeader) (*int, error) {
	return p.paisRepository.RegistrarPais(ctx, request, fileHeader)
}

func (p PaisService) ModificarPais(ctx context.Context, id *int, request *domain.PaisRequest, fileHeader *multipart.FileHeader) error {
	return p.paisRepository.ModificarPais(ctx, id, request, fileHeader)
}

func (p PaisService) ObtenerPaisById(ctx context.Context, id *int) (*domain.PaisDetail, error) {
	return p.paisRepository.ObtenerPaisById(ctx, id)
}

func (p PaisService) ObtenerListaPaises(ctx context.Context, filtros map[string]string) (*[]domain.PaisInfo, error) {
	return p.paisRepository.ObtenerListaPaises(ctx, filtros)
}

func NewPaisService(paisRepository port.PaisRepository) *PaisService {
	return &PaisService{paisRepository: paisRepository}
}

var _ port.PaisService = (*PaisService)(nil)
