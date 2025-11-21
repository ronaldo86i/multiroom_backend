package service

import (
	"context"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/port"
)

type VentaService struct {
	ventaService port.VentaService
}

func (v VentaService) RegistrarVenta(ctx context.Context, request *domain.VentaRequest) (*int, error) {
	return v.ventaService.RegistrarVenta(ctx, request)
}

func (v VentaService) AnularVentaById(ctx context.Context, id *int) error {
	return v.ventaService.AnularVentaById(ctx, id)
}

func (v VentaService) RegistrarPagoVenta(ctx context.Context, ventaId *int, request *domain.RegistrarPagosRequest) (*[]int, error) {
	return v.ventaService.RegistrarPagoVenta(ctx, ventaId, request)
}

func (v VentaService) ObtenerVenta(ctx context.Context, id *int) (*domain.Venta, error) {
	return v.ventaService.ObtenerVenta(ctx, id)
}

func (v VentaService) ListarVentas(ctx context.Context, filtros map[string]string) (*[]domain.VentaInfo, error) {
	return v.ventaService.ListarVentas(ctx, filtros)
}

func NewVentaService(ventaService port.VentaService) *VentaService {
	return &VentaService{ventaService: ventaService}
}

var _ port.VentaService = (*VentaService)(nil)
