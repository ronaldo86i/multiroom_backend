package service

import (
	"context"
	"fmt"
	"log"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"strconv"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontfamily"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type ReporteService struct {
	ventaRepository    port.VentaRepository
	sucursalRepository port.SucursalRepository
	productoRepository port.ProductoRepository
}

func (r ReporteService) ReportePDFProductosVendidos(ctx context.Context, filtros map[string]string) (core.Document, error) {
	// 1. Obtener Datos
	stats, err := r.ventaRepository.ListarProductosVentas(ctx, filtros)
	if err != nil {
		return nil, err
	}

	// --- PALETA DE COLORES FORMAL ---
	colorHeaderBg := &props.Color{Red: 230, Green: 230, Blue: 230}
	colorZebraEven := &props.Color{Red: 250, Green: 250, Blue: 250}
	colorZebraOdd := &props.Color{Red: 255, Green: 255, Blue: 255}
	colorLine := &props.Color{Red: 100, Green: 100, Blue: 100}

	// 2. Configurar PDF (Vertical es mejor para rankings)
	gridSum := 12
	pageNumber := props.PageNumber{
		Pattern: "Página {current} de {total}",
		Place:   props.RightBottom,
		Family:  fontfamily.Arial,
		Style:   fontstyle.Italic,
		Size:    8,
		Color:   &props.Color{Red: 100, Green: 100, Blue: 100},
	}

	cfg := config.NewBuilder().
		WithPageSize(pagesize.Letter).
		WithOrientation(orientation.Vertical). // Vertical
		WithTopMargin(10).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithBottomMargin(10).
		WithMaxGridSize(gridSum).
		WithPageNumber(pageNumber).
		Build()

	m := maroto.New(cfg)

	// --- Estilos ---
	titleStyle := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 14}
	subTitleStyle := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 10, Color: colorLine}

	// Estilos de celda
	headerCellStyle := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 9, Top: 1.5}
	rowTextStyle := props.Text{Align: align.Left, Size: 8, Top: 1}
	rowNumStyle := props.Text{Align: align.Center, Size: 8, Top: 1}
	rowMoneyStyle := props.Text{Align: align.Right, Size: 8, Top: 1}

	now := time.Now().Format("02/01/2006 15:04")

	// ==========================================
	// 1. CABECERA
	// ==========================================

	// Fila 1: Título Empresa
	r1 := row.New(12).Add(
		text.NewCol(8, "ESCONDITE MULTIROOM", titleStyle),
		text.NewCol(4, fmt.Sprintf("Generado: %s", now), props.Text{Align: align.Right, Size: 8, Style: fontstyle.Italic}),
	)

	// Fila 2: Filtros Dinámicos
	var partesFiltro []string

	// A. Sucursal
	if val, ok := filtros["sucursalId"]; ok && val != "" {
		sucursalId, _ := strconv.Atoi(val)
		// Validamos existencia si tienes el repo inyectado
		if r.sucursalRepository != nil {
			sucursal, _ := r.sucursalRepository.ObtenerSucursalById(ctx, &sucursalId)
			if sucursal != nil {
				partesFiltro = append(partesFiltro, fmt.Sprintf("Sucursal: %s", sucursal.Nombre))
			} else {
				partesFiltro = append(partesFiltro, fmt.Sprintf("Sucursal ID: %s", val))
			}
		}
	} else {
		partesFiltro = append(partesFiltro, "Sucursal: TODAS")
	}

	// B. Fechas (Muy importante en reporte de ventas)
	if val, ok := filtros["fechaInicio"]; ok && val != "" {
		partesFiltro = append(partesFiltro, fmt.Sprintf("Desde: %s", val))
	}
	if val, ok := filtros["fechaFin"]; ok && val != "" {
		partesFiltro = append(partesFiltro, fmt.Sprintf("Hasta: %s", val))
	}

	// C. Categoría
	if val, ok := filtros["categoriaId"]; ok && val != "" {
		partesFiltro = append(partesFiltro, fmt.Sprintf("Cat ID: %s", val))
	}

	infoFiltros := strings.Join(partesFiltro, " | ")

	r2 := row.New(10).Add(
		text.NewCol(5, "RANKING DE PRODUCTOS VENDIDOS", subTitleStyle),
		text.NewCol(7, infoFiltros, props.Text{Align: align.Right, Size: 9}),
	)

	r3 := row.New(4) // Espacio

	// Fila 4: Encabezados de Tabla
	// Distribución (12): Producto(7), Cantidad(2), Total(3)
	r4 := row.New(8).Add(
		text.NewCol(7, "PRODUCTO", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 9, Top: 1.5}),
		text.NewCol(2, "CANT. VENDIDA", headerCellStyle),
		text.NewCol(3, "SUB TOTAL", props.Text{Style: fontstyle.Bold, Align: align.Right, Size: 9, Top: 1.5}),
	).WithStyle(&props.Cell{BackgroundColor: colorHeaderBg})

	// Línea separadora
	r5 := row.New(1).Add(
		line.NewCol(gridSum, props.Line{Color: colorLine}),
	)

	if err := m.RegisterHeader(r1, r2, r3, r4, r5); err != nil {
		return nil, err
	}

	// ==========================================
	// 2. CUERPO
	// ==========================================
	var granTotalDinero float64 = 0
	var granTotalCantidad int = 0
	cantidadItems := 0

	if stats != nil {
		for i, s := range *stats {
			// Acumuladores
			granTotalDinero += s.TotalVentas
			granTotalCantidad += s.CantidadVentas
			cantidadItems++

			// Zebra Striping
			currentRowColor := colorZebraOdd
			if i%2 == 0 {
				currentRowColor = colorZebraEven
			}

			// Truncar nombre largo
			nombreProd := s.Producto.Nombre
			if len(nombreProd) > 55 {
				nombreProd = nombreProd[:52] + "..."
			}

			m.AddRow(6,
				text.NewCol(7, nombreProd, rowTextStyle),
				text.NewCol(2, fmt.Sprintf("%d", s.CantidadVentas), rowNumStyle),
				text.NewCol(3, fmt.Sprintf("%.2f", s.TotalVentas), rowMoneyStyle),
			).WithStyle(&props.Cell{BackgroundColor: currentRowColor})
		}
	}

	// ==========================================
	// 3. PIE DE PÁGINA (TOTALES)
	// ==========================================

	// Línea final
	m.AddRow(2, line.NewCol(gridSum, props.Line{Color: colorLine}))
	m.AddRow(2)

	// Resumen Ejecutivo
	m.AddRow(12,
		text.NewCol(5, fmt.Sprintf("Items Listados: %d", cantidadItems), props.Text{Align: align.Left, Size: 9, Top: 2}),

		// Totales alineados con sus columnas
		text.NewCol(4, fmt.Sprintf("Cant. Total: %d", granTotalCantidad), props.Text{Align: align.Right, Style: fontstyle.Bold, Size: 9, Top: 2}),

		text.NewCol(3, fmt.Sprintf("Bs %.2f", granTotalDinero), props.Text{Align: align.Right, Style: fontstyle.Bold, Size: 11, Top: 1}),
	)

	// Generar
	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (r ReporteService) ReportePDFVentas(ctx context.Context, filtros map[string]string) (core.Document, error) {
	// 1. Obtener Datos
	ventas, err := r.ventaRepository.ListarVentas(ctx, filtros)
	if err != nil {
		return nil, err
	}

	// --- PALETA DE COLORES FORMAL ---
	colorHeaderBg := &props.Color{Red: 230, Green: 230, Blue: 230}  // Gris Claro para cabecera tabla
	colorZebraEven := &props.Color{Red: 250, Green: 250, Blue: 250} // Blanco Humo para filas pares
	colorZebraOdd := &props.Color{Red: 255, Green: 255, Blue: 255}  // Blanco puro
	colorLine := &props.Color{Red: 100, Green: 100, Blue: 100}      // Gris oscuro para líneas

	// 2. Configurar PDF
	gridSum := 24
	pageNumber := props.PageNumber{
		Pattern: "Página {current} de {total}",
		Place:   props.RightBottom,
		Family:  fontfamily.Arial,
		Style:   fontstyle.Italic,
		Size:    8,
		Color:   &props.Color{Red: 100, Green: 100, Blue: 100},
	}

	cfg := config.NewBuilder().
		WithPageSize(pagesize.Letter).
		WithOrientation(orientation.Horizontal).
		WithTopMargin(10).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithBottomMargin(10).
		WithMaxGridSize(gridSum).
		WithPageNumber(pageNumber).
		Build()

	m := maroto.New(cfg)

	// --- Estilos de Texto ---
	titleStyle := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 14}
	subTitleStyle := props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 10, Color: colorLine}
	tableHeaderStyle := props.Text{Style: fontstyle.Bold, Align: align.Center, Size: 9, Top: 1.5} // Top para centrar verticalmente
	rowTextStyle := props.Text{Align: align.Left, Size: 8, Top: 1}
	rowMoneyStyle := props.Text{Align: align.Right, Size: 8, Top: 1}

	now := time.Now().Format("02/01/2006 15:04")

	// ==========================================
	// 1. CABECERA (HEADER)
	// ==========================================

	// Fila 1: Nombre Empresa y Fecha
	r1 := row.New(12).Add(
		text.NewCol(16, "ESCONDITE MULTIROOM", titleStyle),
		text.NewCol(8, fmt.Sprintf("Impreso: %s", now), props.Text{Align: align.Right, Size: 8, Style: fontstyle.Italic}),
	)

	// Fila 2: Subtítulo y Filtros Dinámicos
	var partesFiltro []string
	if val, ok := filtros["sucursalId"]; ok && val != "" {
		sucursalId, err := strconv.Atoi(val)
		if err != nil {
			log.Println("Error al convertir sucursalId a int:", err)
			return nil, datatype.NewBadRequestError("El valor de sucursalId no es válido")
		}
		sucursal, err := r.sucursalRepository.ObtenerSucursalById(ctx, &sucursalId)
		if err != nil {
			return nil, err
		}
		partesFiltro = append(partesFiltro, fmt.Sprintf("Sucursal: %s, País: %s", sucursal.Nombre, sucursal.Pais.Nombre))
	}

	if val, ok := filtros["fechaInicio"]; ok && val != "" {
		partesFiltro = append(partesFiltro, fmt.Sprintf("Desde: %s", val))
	}
	if val, ok := filtros["fechaFin"]; ok && val != "" {
		partesFiltro = append(partesFiltro, fmt.Sprintf("Hasta: %s", val))
	}

	infoFiltros := "Filtros: General"
	if len(partesFiltro) > 0 {
		infoFiltros = strings.Join(partesFiltro, " | ")
	}

	r2 := row.New(10).Add(
		text.NewCol(12, "REPORTE GENERAL DE VENTAS", subTitleStyle),
		text.NewCol(12, infoFiltros, props.Text{Align: align.Right, Size: 9}),
	)

	// Fila 3: Espaciado
	r3 := row.New(4)

	// Fila 4: Encabezados de Tabla (Fondo Gris)
	// Usamos .WithStyle(&props.Cell{BackgroundColor: ...}) para dar color a toda la franja
	r4 := row.New(7).Add(
		text.NewCol(4, "FECHA", tableHeaderStyle),
		text.NewCol(4, "CÓDIGO", tableHeaderStyle),
		text.NewCol(4, "SUCURSAL", tableHeaderStyle),
		text.NewCol(5, "VENDEDOR", props.Text{Style: fontstyle.Bold, Align: align.Left, Size: 9, Top: 1.5}),
		text.NewCol(3, "ESTADO", tableHeaderStyle),
		text.NewCol(4, "TOTAL", props.Text{Style: fontstyle.Bold, Align: align.Right, Size: 9, Top: 1.5}),
	).WithStyle(&props.Cell{BackgroundColor: colorHeaderBg})

	// Fila 5: Línea sutil debajo de la cabecera
	r5 := row.New(1).Add(
		line.NewCol(gridSum, props.Line{Color: colorLine}),
	)

	if err := m.RegisterHeader(r1, r2, r3, r4, r5); err != nil {
		return nil, err
	}

	// ==========================================
	// 2. CUERPO (DATOS)
	// ==========================================
	var totalGeneral float64 = 0
	var cantidadVentas = 0

	if ventas != nil {
		for i, v := range *ventas {
			// Lógica de datos
			nombreCliente := "S/N"
			if v.Cliente != nil {
				nombreCliente = fmt.Sprintf("%s %s", v.Cliente.Nombres, v.Cliente.Apellidos)
				// Truncar nombre si es muy largo para mantener formalidad
				if len(nombreCliente) > 25 {
					nombreCliente = nombreCliente[:22] + "..."
				}
			}
			nombreSucursal := "-"
			if v.Sucursal != nil {
				nombreSucursal = v.Sucursal.Nombre
			}

			if v.Estado == "Completado" {
				totalGeneral += v.Total
				cantidadVentas++
			}

			// Alternar colores (Zebra Striping)
			currentRowColor := colorZebraOdd
			if i%2 == 0 {
				currentRowColor = colorZebraEven
			}

			// Agregamos la fila con color de fondo alternado
			m.AddRow(6,
				text.NewCol(4, v.CreadoEn.Format("02/01/06 15:04"), rowTextStyle),
				text.NewCol(4, fmt.Sprintf("%07d", v.CodigoVenta), props.Text{Align: align.Center, Size: 8, Top: 1}),
				text.NewCol(4, nombreSucursal, rowTextStyle),
				text.NewCol(5, v.Usuario.Username, rowTextStyle),
				text.NewCol(3, v.Estado, props.Text{Align: align.Center, Size: 8, Top: 1}),
				text.NewCol(4, fmt.Sprintf("%.2f", v.Total), rowMoneyStyle),
			).WithStyle(&props.Cell{BackgroundColor: currentRowColor})
		}
	}

	// ==========================================
	// 3. TOTALES FINALES
	// ==========================================

	// Línea separadora final
	m.AddRow(2, line.NewCol(gridSum, props.Line{Color: colorLine}))

	m.AddRow(2) // Espacio pequeño

	// Caja de Totales
	m.AddRow(10,
		text.NewCol(15, "RESUMEN EJECUTIVO:", props.Text{Align: align.Right, Style: fontstyle.Bold, Size: 9, Top: 2}),
		text.NewCol(6, fmt.Sprintf("Transacciones: %d", cantidadVentas), props.Text{Align: align.Right, Size: 9, Top: 2}),
		text.NewCol(3, fmt.Sprintf("Bs %.2f", totalGeneral), props.Text{Align: align.Right, Style: fontstyle.Bold, Size: 10, Top: 1}),
	)

	// Generar
	document, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return document, nil
}

func (r ReporteService) ComprobantePDFVentaById(ctx context.Context, ventaId *int) (core.Document, error) {
	// 1. Obtener Datos de la Venta
	venta, err := r.ventaRepository.ObtenerVenta(ctx, ventaId)
	if err != nil {
		return nil, err
	}

	// 2. Configuración para Ticket Térmico (80mm)
	gridSum := 24
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
		text.NewCol(9, "PROD.", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Left}),
		text.NewCol(3, "PRE.", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Left}),
		text.NewCol(3, "CANT.", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Center}),
		text.NewCol(3, "DESC.", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Center}),
		text.NewCol(6, "SUB.", props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Right}),
	)

	m.AddRow(2, text.NewCol(gridSum, separatorDashed, props.Text{Align: align.Center, Size: 8}))

	var subTotalAcumulado float64 = 0

	// 3.1 Listar Productos
	for _, d := range venta.Detalles {
		totalLinea := float64(d.Cantidad)*d.PrecioVenta - d.Descuento
		subTotalAcumulado += totalLinea

		m.AddRow(4,
			text.NewCol(9, d.Producto.Nombre, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(3, fMoney(d.PrecioVenta), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(3, fmt.Sprintf("%d", d.Cantidad), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(3, fMoney(d.Descuento), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(6, fMoney(totalLinea), props.Text{Size: 8, Align: align.Right}),
		)
	}

	// 3.2 Listar Tiempo de Sala (Como si fuera un producto más)
	if venta.CostoTiempoVenta > 0 {
		// Agregamos el costo del tiempo al subtotal para que cuadre la suma visual
		subTotalAcumulado += venta.CostoTiempoVenta

		nombreItemSala := "USO SALA"

		m.AddRow(4,
			text.NewCol(9, nombreItemSala, props.Text{Size: 8, Align: align.Left}),
			text.NewCol(3, fMoney(venta.CostoTiempoVenta), props.Text{Size: 8, Align: align.Center}),
			text.NewCol(3, "1", props.Text{Size: 8, Align: align.Center}),          // Cantidad
			text.NewCol(3, fMoney(0.00), props.Text{Size: 8, Align: align.Center}), // Descuento
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
			text.NewCol(12, "Descuento Adicional:", props.Text{Align: align.Right, Size: 8}),
			text.NewCol(8, "-"+fMoney(venta.DescuentoGeneral), props.Text{Align: align.Right, Size: 8}),
		)
	}

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

func NewReporteService(ventaRepository port.VentaRepository, sucursalRepository port.SucursalRepository, productoRepository port.ProductoRepository) *ReporteService {
	return &ReporteService{ventaRepository: ventaRepository, sucursalRepository: sucursalRepository, productoRepository: productoRepository}
}

var _ port.ReporteService = (*ReporteService)(nil)
