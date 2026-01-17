package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type ProductoCategoriaService struct {
	productoCategoriaRepository port.ProductoCategoriaRepository
}

func (p ProductoCategoriaService) RegistrarCategoria(ctx context.Context, request *domain.ProductoCategoriaRequest) (*int, error) {
	return p.productoCategoriaRepository.RegistrarCategoria(ctx, request)
}

func (p ProductoCategoriaService) ModificarCategoriaById(ctx context.Context, id *int, request *domain.ProductoCategoriaRequest) error {
	return p.productoCategoriaRepository.ModificarCategoriaById(ctx, id, request)
}

func (p ProductoCategoriaService) ListarCategorias(ctx context.Context, filtros map[string]string) (*[]domain.ProductoCategoria, error) {
	return p.productoCategoriaRepository.ListarCategorias(ctx, filtros)
}

func (p ProductoCategoriaService) ObtenerCategoriaById(ctx context.Context, id *int) (*domain.ProductoCategoria, error) {
	return p.productoCategoriaRepository.ObtenerCategoriaById(ctx, id)
}

func NewProductoCategoriaService(productoCategoriaRepository port.ProductoCategoriaRepository) *ProductoCategoriaService {
	return &ProductoCategoriaService{productoCategoriaRepository: productoCategoriaRepository}
}

var _ port.ProductoCategoriaService = (*ProductoCategoriaService)(nil)
