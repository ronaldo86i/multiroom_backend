package service

import (
	"context"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
)

type ProductoService struct {
	productoRepository port.ProductoRepository
}

func (p ProductoService) RegistrarProducto(ctx context.Context, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) (*int, error) {
	if !util.File.ValidarTipoArchivo(fileHeader.Filename, ".png", ".jpg", ".jpeg") {
		return nil, datatype.NewBadRequestError("Tipo de archivo no válido")
	}
	return p.productoRepository.RegistrarProducto(ctx, request, fileHeader)
}

func (p ProductoService) ModificarProductoById(ctx context.Context, productoId *int, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) error {
	if !util.File.ValidarTipoArchivo(fileHeader.Filename, ".png", ".jpg", ".jpeg") {
		return datatype.NewBadRequestError("Tipo de archivo no válido")
	}
	return p.productoRepository.ModificarProductoById(ctx, productoId, request, fileHeader)
}

func (p ProductoService) ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.Producto, error) {
	return p.productoRepository.ListarProductos(ctx, filtros)
}

func (p ProductoService) ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error) {
	return p.productoRepository.ObtenerProductoById(ctx, productoId)
}

func NewProductoService(productoRepository port.ProductoRepository) *ProductoService {
	return &ProductoService{productoRepository: productoRepository}
}

var _ port.ProductoService = (*ProductoService)(nil)
