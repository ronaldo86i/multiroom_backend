package service

import (
	"context"
	"fmt"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
)

type CompraService struct {
	compraRepository port.CompraRepository
}

func (c CompraService) ListarCompras(ctx context.Context, filtros map[string]string) (*[]domain.CompraInfo, error) {
	return c.compraRepository.ListarCompras(ctx, filtros)
}

func (c CompraService) ObtenerCompraById(ctx context.Context, id *int) (*domain.Compra, error) {
	return c.compraRepository.ObtenerCompraById(ctx, id)
}

func (c CompraService) RegistrarOrdenCompra(ctx context.Context, request *domain.CompraRequest) (*int, error) {

	// Usamos un map para guardar el primer precio que vemos de cada producto
	type precios struct {
		Compra float64
		Venta  float64
	}
	preciosVistos := make(map[int]precios) // Key: productoId

	for _, detalle := range request.Detalles {

		if p, ok := preciosVistos[detalle.ProductoId]; ok {
			// Ya hemos visto este producto. Comparamos los precios.
			if p.Compra != detalle.PrecioCompra || p.Venta != detalle.PrecioVenta {
				// Los precios no coinciden
				errMsg := fmt.Sprintf(
					"El productoId %d tiene precios diferentes en la misma orden. Unifique los precios.",
					detalle.ProductoId,
				)
				return nil, datatype.NewBadRequestError(errMsg)
			}
		} else {
			// Es la primera vez que vemos este producto. Guardamos sus precios.
			preciosVistos[detalle.ProductoId] = precios{
				Compra: detalle.PrecioCompra,
				Venta:  detalle.PrecioVenta,
			}
		}
	}

	// Definimos una clave para nuestro map que representa la combinación (producto + ubicación)
	type detalleKey struct {
		ProductoId  int
		UbicacionId int
	}

	// Creamos un map para rastrear las combinaciones que ya hemos visto.
	// map[Clave] -> Valor (usamos 'struct{}' porque no nos importa el valor, solo la clave)
	vistos := make(map[detalleKey]struct{})

	for _, detalle := range request.Detalles {
		key := detalleKey{
			ProductoId:  detalle.ProductoId,
			UbicacionId: detalle.UbicacionId,
		}

		// ¿Ya hemos visto esta combinación?
		if _, ok := vistos[key]; ok {
			// ¡Sí! Es un duplicado.
			errMsg := fmt.Sprintf(
				"Producto duplicado (productoId: %d) en la misma ubicación (ubicacionId: %d). Agrupe las cantidades en una sola línea.",
				detalle.ProductoId,
				detalle.UbicacionId,
			)
			return nil, datatype.NewBadRequestError(errMsg)
		}

		// No es un duplicado, lo marcamos como visto para la próxima iteración.
		vistos[key] = struct{}{}
	}

	// Si el bucle termina, significa que no hay duplicados.
	// Ahora podemos llamar al repositorio de forma segura.
	return c.compraRepository.RegistrarOrdenCompra(ctx, request)
}

func (c CompraService) ModificarOrdenCompra(ctx context.Context, id *int, request *domain.CompraRequest) error {
	// Usamos un map para guardar el primer precio que vemos de cada producto
	type precios struct {
		Compra float64
		Venta  float64
	}
	preciosVistos := make(map[int]precios) // Key: productoId

	for _, detalle := range request.Detalles {

		if p, ok := preciosVistos[detalle.ProductoId]; ok {
			// Ya hemos visto este producto. Comparamos los precios.
			if p.Compra != detalle.PrecioCompra || p.Venta != detalle.PrecioVenta {
				// Los precios no coinciden
				errMsg := fmt.Sprintf(
					"El productoId %d tiene precios diferentes en la misma orden. Unifique los precios.",
					detalle.ProductoId,
				)
				return datatype.NewBadRequestError(errMsg)
			}
		} else {
			// Es la primera vez que vemos este producto. Guardamos sus precios.
			preciosVistos[detalle.ProductoId] = precios{
				Compra: detalle.PrecioCompra,
				Venta:  detalle.PrecioVenta,
			}
		}
	}

	// Definimos una clave para nuestro map que representa la combinación (producto + ubicación)
	type detalleKey struct {
		ProductoId  int
		UbicacionId int
	}

	// Creamos un map para rastrear las combinaciones que ya hemos visto.
	// map[Clave] -> Valor (usamos 'struct{}' porque no nos importa el valor, solo la clave)
	vistos := make(map[detalleKey]struct{})

	for _, detalle := range request.Detalles {
		key := detalleKey{
			ProductoId:  detalle.ProductoId,
			UbicacionId: detalle.UbicacionId,
		}

		// ¿Ya hemos visto esta combinación?
		if _, ok := vistos[key]; ok {
			// ¡Sí! Es un duplicado.
			errMsg := fmt.Sprintf(
				"Producto duplicado (productoId: %d) en la misma ubicación (ubicacionId: %d). Agrupe las cantidades en una sola línea.",
				detalle.ProductoId,
				detalle.UbicacionId,
			)
			return datatype.NewBadRequestError(errMsg)
		}

		// No es un duplicado, lo marcamos como visto para la próxima iteración.
		vistos[key] = struct{}{}
	}

	// Si el bucle termina, significa que no hay duplicados.
	// Ahora podemos llamar al repositorio de forma segura.
	return c.compraRepository.ModificarOrdenCompra(ctx, id, request)
}

func (c CompraService) ConfirmarRecepcionCompra(ctx context.Context, id *int) error {
	return c.compraRepository.ConfirmarRecepcionCompra(ctx, id)
}

func NewCompraService(compraRepository port.CompraRepository) *CompraService {
	return &CompraService{compraRepository: compraRepository}
}

var _ port.CompraService = (*CompraService)(nil)
