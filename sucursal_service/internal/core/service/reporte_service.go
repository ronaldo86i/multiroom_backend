package service

import (
	"context"
	"fmt"
	"multiroom/sucursal-service/internal/core/port"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type ReporteService struct {
	ventaRepository port.VentaRepository
}

func (r ReporteService) ComprobantePDFVentaById(ctx context.Context, ventaId *int) (core.Document, error) {
	// 1. Obtener Datos de la Venta
	venta, err := r.ventaRepository.ObtenerVenta(ctx, ventaId)
	if err != nil {
		return nil, err
	}

	// 2. Configuración para Ticket Térmico (80mm)
	gridSum := 20
	cfg := config.NewBuilder().
		WithTopMargin(5).
		WithLeftMargin(2).
		WithRightMargin(2).
		WithBottomMargin(5).
		WithDisableAutoPageBreak(false).
		WithDimensions(80, 200).
		WithMaxGridSize(gridSum).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)

	// --- Helpers ---
	fMoney := func(val float64) string {
		return fmt.Sprintf("%.2f", val)
	}
	separatorDouble := "=================================="
	separatorDashed := "----------------------------------"

	// ==========================================
	// 1. CABECERA
	// ==========================================
	m.AddRow(5,
		text.NewCol(gridSum, "ESCONDITE MULTIROOM", props.Text{
			Style: fontstyle.Bold,
			Align: align.Center,
			Size:  11,
		}),
	)

	nombreSucursal := "Central"
	if venta.Sucursal != nil {
		nombreSucursal = venta.Sucursal.Nombre
	}
	m.AddRow(4,
		text.NewCol(gridSum, nombreSucursal, props.Text{Align: align.Center, Size: 9}),
	)
	m.AddRow(4,
		text.NewCol(gridSum, "Tel. 76328248", props.Text{Align: align.Center, Size: 8}),
	)

	m.AddRow(3, text.NewCol(gridSum, separatorDouble, props.Text{Align: align.Center, Size: 8}))

	// ==========================================
	// 2. DATOS DE LA ORDEN
	// ==========================================
	m.AddRow(4,
		text.NewCol(gridSum, fmt.Sprintf("ORDEN #: %06d", venta.CodigoVenta), props.Text{
			Style: fontstyle.Bold,
			Size:  9,
			Align: align.Center,
		}),
	)

	labelStyle := props.Text{Align: align.Left, Size: 8}
	valStyle := props.Text{Align: align.Left, Size: 8}

	m.AddRow(4,
		text.NewCol(5, "Fecha:", labelStyle),
		text.NewCol(15, venta.CreadoEn.Format("02/01/2006"), valStyle),
	)
	m.AddRow(4,
		text.NewCol(5, "Hora:", labelStyle),
		text.NewCol(15, venta.CreadoEn.Format("15:04"), valStyle),
	)

	m.AddRow(4,
		text.NewCol(5, "Atendió:", labelStyle),
		text.NewCol(15, venta.Usuario.Username, valStyle),
	)

	nombreSala := "-"
	if venta.Sala != nil {
		nombreSala = venta.Sala.Nombre
	}
	m.AddRow(4,
		text.NewCol(5, "Sala:", labelStyle),
		text.NewCol(15, nombreSala, valStyle),
	)

	m.AddRow(3, text.NewCol(gridSum, separatorDouble, props.Text{Align: align.Center, Size: 8}))

	// ==========================================
	// 3. TABLA DE DETALLES (Productos + Tiempo Sala)
	// ==========================================
	m.AddRow(4,
		text.NewCol(11, "PRODUCTO", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Left}),
		text.NewCol(3, "CANT", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Center}),
		text.NewCol(6, "SUB", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Right}),
	)

	m.AddRow(2, text.NewCol(gridSum, separatorDashed, props.Text{Align: align.Center, Size: 8}))

	var subTotalAcumulado float64 = 0

	// 3.1 Listar Productos
	for _, d := range venta.Detalles {
		totalLinea := float64(d.Cantidad) * d.PrecioVenta
		subTotalAcumulado += totalLinea

		m.AddRow(4,
			text.NewCol(11, d.Producto.Nombre, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(3, fmt.Sprintf("%d", d.Cantidad), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(6, fMoney(totalLinea), props.Text{Size: 8, Align: align.Right}),
		)
	}

	// 3.2 Listar Tiempo de Sala (Como si fuera un producto más)
	if venta.CostoTiempoVenta > 0 {
		// Agregamos el costo del tiempo al subtotal para que cuadre la suma visual
		subTotalAcumulado += venta.CostoTiempoVenta

		nombreItemSala := "ALQUILER SALA"
		if venta.Sala != nil {
			nombreItemSala = fmt.Sprintf("USO SALA %s", venta.Sala.Nombre)
		}

		m.AddRow(4,
			text.NewCol(11, nombreItemSala, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(3, "1", props.Text{Size: 8, Align: align.Center}), // Cantidad 1 servicio
			text.NewCol(6, fMoney(venta.CostoTiempoVenta), props.Text{Size: 8, Align: align.Right}),
		)
	}

	m.AddRow(2, text.NewCol(gridSum, separatorDashed, props.Text{Align: align.Center, Size: 8}))

	// ==========================================
	// 4. TOTALES
	// ==========================================

	// Subtotal (Ahora incluye productos + tiempo sala)
	m.AddRow(4,
		text.NewCol(12, "Subtotal:", props.Text{Align: align.Right, Size: 8}),
		text.NewCol(8, fMoney(subTotalAcumulado), props.Text{Align: align.Right, Size: 8}),
	)

	// Descuento General
	if venta.DescuentoGeneral > 0 {
		m.AddRow(4,
			text.NewCol(12, "Descuento Gral:", props.Text{Align: align.Right, Size: 8}),
			text.NewCol(8, "-"+fMoney(venta.DescuentoGeneral), props.Text{Align: align.Right, Size: 8}),
		)
	}

	// NOTA: Eliminé la sección "Servicio Sala" de aquí abajo porque ya lo agregamos
	// arriba en el detalle para que se vea como un ítem más.

	// TOTAL FINAL
	m.AddRow(6,
		text.NewCol(10, "TOTAL:", props.Text{Align: align.Right, Style: fontstyle.Bold, Size: 11}),
		text.NewCol(10, fmt.Sprintf("Bs.%s", fMoney(venta.Total)), props.Text{Align: align.Right, Style: fontstyle.Bold, Size: 11}),
	)

	// ==========================================
	// 5. DETALLE DE PAGOS
	// ==========================================

	if venta.Pagos != nil && len(*venta.Pagos) > 0 {
		m.AddRow(4,
			text.NewCol(gridSum, "PAGADO CON:", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Left}),
		)

		for _, pago := range *venta.Pagos {
			m.AddRow(4,
				text.NewCol(10, pago.MetodoPago.Nombre, props.Text{Size: 8, Align: align.Left}),
				text.NewCol(10, fMoney(pago.Monto), props.Text{Size: 8, Align: align.Right}),
			)
		}

		var totalPagado float64
		for _, p := range *venta.Pagos {
			totalPagado += p.Monto
		}

		if totalPagado > venta.Total {
			cambio := totalPagado - venta.Total
			m.AddRow(4,
				text.NewCol(10, "CAMBIO:", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Left}),
				text.NewCol(10, fMoney(cambio), props.Text{Size: 8, Align: align.Right}),
			)
		}

	} else {
		m.AddRow(4,
			text.NewCol(6, "ESTADO:", props.Text{Size: 8, Align: align.Left}),
			text.NewCol(14, "Pendiente de Pago", props.Text{Size: 8, Align: align.Right}),
		)
	}

	m.AddRow(2, text.NewCol(gridSum, separatorDashed, props.Text{Align: align.Center, Size: 8}))

	// ==========================================
	// 6. PIE DE PÁGINA
	// ==========================================
	m.AddRow(4,
		text.NewCol(gridSum, "¡Gracias por su compra!", props.Text{Align: align.Center, Size: 9}),
	)
	m.AddRow(4,
		text.NewCol(gridSum, "Vuelva pronto", props.Text{Align: align.Center, Size: 9}),
	)

	m.AddRow(3, text.NewCol(gridSum, separatorDouble, props.Text{Align: align.Center, Size: 8}))

	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document, nil
}
func NewReporteService(ventaRepository port.VentaRepository) *ReporteService {
	return &ReporteService{ventaRepository: ventaRepository}
}

var _ port.ReporteService = (*ReporteService)(nil)
