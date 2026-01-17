package service

import (
	"context"
	"multiroom/auth-service/internal/core/domain"
	"multiroom/auth-service/internal/core/domain/datatype"
	"multiroom/auth-service/internal/core/port"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type PermisoService struct {
	permisoRepository port.PermisoRepository
}

var validate = validator.New()

func (p PermisoService) ListarPermisos(ctx context.Context, filtros map[string]string) (*[]domain.Permiso, error) {
	return p.permisoRepository.ListarPermisos(ctx, filtros)
}

func (p PermisoService) ObtenerPermisoById(ctx context.Context, id *int) (*domain.Permiso, error) {
	return p.permisoRepository.ObtenerPermisoById(ctx, id)
}

func (p PermisoService) RegistrarPermiso(ctx context.Context, request *domain.PermisoRequest) (*int, error) {
	var validPermisoRegex = regexp.MustCompile(`^[a-z0-9_]+:[a-z0-9_]+$`)

	request.Nombre = strings.ToLower(strings.TrimSpace(request.Nombre))
	request.Descripcion = strings.TrimSpace(request.Descripcion)
	if err := validate.Struct(request); err != nil {
		return nil, datatype.NewBadRequestError("Datos inválidos: " + err.Error())
	}
	if !validPermisoRegex.MatchString(request.Nombre) {
		return nil, datatype.NewBadRequestError("Formato inválido. Use 'recurso:accion' (ej: venta:crear).")
	}
	return p.permisoRepository.RegistrarPermiso(ctx, request)
}

func (p PermisoService) ModificarPermisoById(ctx context.Context, id *int, request *domain.PermisoRequest) error {
	request.Descripcion = strings.TrimSpace(request.Descripcion)
	return p.permisoRepository.ModificarPermisoById(ctx, id, request)
}

func NewPermisoService(permisoRepository port.PermisoRepository) *PermisoService {
	return &PermisoService{permisoRepository: permisoRepository}
}

var _ port.PermisoService = (*PermisoService)(nil)
