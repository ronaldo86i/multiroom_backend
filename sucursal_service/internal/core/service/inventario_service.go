package service

import (
	"context"
	"fmt"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
)

type InventarioService struct {
	inventarioRepository port.InventarioRepository
}

func (i InventarioService) ObtenerTransferenciaById(ctx context.Context, id *int) (*domain.TransferenciaInventario, error) {
	return i.inventarioRepository.ObtenerTransferenciaById(ctx, id)
}

func (i InventarioService) ObtenerAjusteById(ctx context.Context, id *int) (*domain.AjusteInventario, error) {
	return i.inventarioRepository.ObtenerAjusteById(ctx, id)
}

func (i InventarioService) ListarTransferencias(ctx context.Context, filtros map[string]string) (*[]domain.TransferenciaInventarioInfo, error) {
	return i.inventarioRepository.ListarTransferencias(ctx, filtros)
}

func (i InventarioService) ListarAjustes(ctx context.Context, filtros map[string]string) (*[]domain.AjusteInventarioInfo, error) {
	return i.inventarioRepository.ListarAjustes(ctx, filtros)
}

func (i InventarioService) RegistrarAjusteConDetalle(ctx context.Context, request *domain.AjusteInventarioRequest) (*int, error) {

	// Validación de Vacío
	if len(request.Detalles) == 0 {
		return nil, datatype.NewBadRequestError("El ajuste debe contener al menos un detalle.")
	}

	// Validación del Tipo de Ajuste (La nueva lógica)
	direction, ok := domain.ValidAjusteTypes[request.TipoAjuste]
	if !ok {
		// El TipoAjuste enviado no existe en nuestro mapa de reglas
		return nil, datatype.NewBadRequestError(fmt.Sprintf("El tipo de ajuste '%s' no es válido.", request.TipoAjuste))
	}

	// Validación de Cantidades (La nueva lógica)
	for _, detalle := range request.Detalles {

		// La BD ya valida esto con un CHECK, pero es bueno validarlo aquí primero
		if detalle.Cantidad == 0 {
			return nil, datatype.NewBadRequestError(fmt.Sprintf("La cantidad para el producto %d no puede ser cero.", detalle.ProductoId))
		}

		switch direction {
		case domain.AjusteSalida:
			// Si el tipo es de Salida, la cantidad NO PUEDE ser positiva
			if detalle.Cantidad > 0 {
				return nil, datatype.NewBadRequestError(fmt.Sprintf(
					"El tipo de ajuste '%s' solo permite cantidades negativas (error en producto %d).",
					request.TipoAjuste, detalle.ProductoId,
				))
			}
		case domain.AjusteEntrada:
			// Si el tipo es de Entrada, la cantidad NO PUEDE ser negativa
			if detalle.Cantidad < 0 {
				return nil, datatype.NewBadRequestError(fmt.Sprintf(
					"El tipo de ajuste '%s' solo permite cantidades positivas (error en producto %d).",
					request.TipoAjuste, detalle.ProductoId,
				))
			}
		case domain.AjusteMixto:
			// Tipo "ERROR_CONTEO": Permite tanto positivos como negativos. No hacemos nada.
		}
	}

	return i.inventarioRepository.RegistrarAjusteConDetalle(ctx, request)
}

func (i InventarioService) RegistrarTransferencia(ctx context.Context, request *domain.TransferenciaRequest) (*int, error) {
	// Validación de Vacío
	if len(request.Detalles) == 0 {
		return nil, datatype.NewBadRequestError("La transferencia debe contener al menos un detalle.")
	}
	return i.inventarioRepository.RegistrarTransferencia(ctx, request)
}

func (i InventarioService) ListarInventario(ctx context.Context, filtros map[string]string) (*[]domain.Inventario, error) {
	return i.inventarioRepository.ListarInventario(ctx, filtros)
}

func NewInventarioService(inventarioRepository port.InventarioRepository) *InventarioService {
	return &InventarioService{inventarioRepository: inventarioRepository}
}

var _ port.InventarioService = (*InventarioService)(nil)
