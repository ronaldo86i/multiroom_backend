package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type ProveedorService struct {
	proveedorRepository port.ProveedorRepository
}

func (p ProveedorService) RegistrarProveedor(ctx context.Context, request *domain.ProveedorRequest) (*int, error) {
	return p.proveedorRepository.RegistrarProveedor(ctx, request)
}

func (p ProveedorService) ModificarProveedor(ctx context.Context, id *int, request *domain.ProveedorRequest) error {
	return p.proveedorRepository.ModificarProveedor(ctx, id, request)
}

func (p ProveedorService) ListarProveedores(ctx context.Context, filtros map[string]string) (*[]domain.Proveedor, error) {
	return p.proveedorRepository.ListarProveedores(ctx, filtros)
}

func (p ProveedorService) ObtenerProveedorById(ctx context.Context, id *int) (*domain.Proveedor, error) {
	return p.proveedorRepository.ObtenerProveedorById(ctx, id)
}

func NewProveedorService(proveedorRepository port.ProveedorRepository) *ProveedorService {
	return &ProveedorService{proveedorRepository: proveedorRepository}
}

var _ port.ProveedorService = (*ProveedorService)(nil)
