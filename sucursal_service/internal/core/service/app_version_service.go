package service

import (
	"context"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type AppVersionService struct {
	appVersionRepository port.AppVersionRepository
}

func (a AppVersionService) ObtenerListaVersiones(ctx context.Context, filtros map[string]string) (*[]domain.AppVersion, error) {
	return a.appVersionRepository.ObtenerListaVersiones(ctx, filtros)
}

func (a AppVersionService) ObtenerUltimaVersion(ctx context.Context, q *domain.AppLastVersionQuery) (*domain.AppVersion, error) {
	return a.appVersionRepository.ObtenerUltimaVersion(ctx, q)
}

func (a AppVersionService) RegistrarApp(ctx context.Context, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) (*int, error) {
	return a.appVersionRepository.RegistrarApp(ctx, request, fileHeader)
}

func (a AppVersionService) ModificarVersion(ctx context.Context, id *int, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) error {
	return a.appVersionRepository.ModificarVersion(ctx, id, request, fileHeader)
}

func (a AppVersionService) ObtenerVersion(ctx context.Context, id *int) (*domain.AppVersion, error) {
	return a.appVersionRepository.ObtenerVersion(ctx, id)
}

func NewAppVersionService(appVersionRepository port.AppVersionRepository) *AppVersionService {
	return &AppVersionService{appVersionRepository: appVersionRepository}
}

var _ port.AppVersionService = (*AppVersionService)(nil)
