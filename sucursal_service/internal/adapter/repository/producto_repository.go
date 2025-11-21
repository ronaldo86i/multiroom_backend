package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductoRepository struct {
	pool *pgxpool.Pool
}

func (p ProductoRepository) RegistrarProducto(ctx context.Context, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) (*int, error) {
	nombreArchivo := strings.ToLower(fileHeader.Filename)
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	var productoId int
	query := `INSERT INTO producto(nombre, estado, foto, precio) VALUES ($1,$2,$3,$4) RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, request.Estado, nombreArchivo, request.Precio).Scan(&productoId)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	routeDir := fmt.Sprintf("./public/uploads/productos/%d", productoId)
	defer func() {
		if !committed {
			_ = util.File.DeleteAllFiles(routeDir)
		}
	}()
	// Crear directorio
	err = util.File.MakeDir(routeDir)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Abrir archivo
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error al abrir archivo")
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Guardar archivo
	err = util.File.SaveFile(routeDir, nombreArchivo, file)
	if err != nil {
		log.Println("Error al guardar imagen:", err)
		return nil, datatype.NewInternalServerError("Error al guardar imagen")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return &productoId, nil
}

func (p ProductoRepository) ModificarProductoById(ctx context.Context, productoId *int, request *domain.ProductoRequest, fileHeader *multipart.FileHeader) error {
	nombreArchivo := strings.ToLower(fileHeader.Filename)
	// Iniciar transacción
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	var committed bool
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	query := `UPDATE producto SET nombre=$1,foto=$2,estado=$3,actualizado_en=now(),precio=$4 WHERE id=$5`
	ct, err := tx.Exec(ctx, query, request.Nombre, nombreArchivo, request.Estado, request.Precio, *productoId)
	if err != nil {
		log.Println("Error al actualizar producto")
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Producto no encontrado")
	}

	// Respaldar archivos existentes
	route := fmt.Sprintf("./public/uploads/productos/%d", *productoId)
	backupFiles, err := util.File.BackupFiles(route)
	if err != nil {
		log.Println(err)
		return datatype.NewInternalServerErrorGeneric()
	}
	defer func() {
		if !committed {
			_ = util.File.RestoreFiles(backupFiles, route)
		}
	}()
	// Eliminar archivos
	err = util.File.DeleteAllFiles(route)
	if err != nil {
		log.Println("Error al eliminar archivos:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	// Crear directorio
	routeDir := fmt.Sprintf("./public/uploads/productos/%d", *productoId)
	err = util.File.MakeDir(routeDir)
	if err != nil {
		log.Println("Error al crear directorio:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	// Abrir archivo
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error al abrir archivo")
		return datatype.NewInternalServerErrorGeneric()
	}

	// Guardar archivo
	err = util.File.SaveFile(routeDir, nombreArchivo, file)
	if err != nil {
		log.Println("Error al guardar imagen:", err)
		return datatype.NewInternalServerError("Error al guardar imagen")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return nil
}

func (p ProductoRepository) ListarProductos(ctx context.Context, filtros map[string]string) (*[]domain.Producto, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `SELECT p.id,p.nombre,p.estado,($1::text || p.id::text || '/' || p.foto) AS url_foto,p.precio,p.creado_en,p.actualizado_en,p.eliminado_en FROM producto p`
	rows, err := p.pool.Query(ctx, query, fullHostname)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.Producto, 0)
	for rows.Next() {
		var item domain.Producto
		err = rows.Scan(&item.Id, &item.Nombre, &item.Estado, &item.UrlFoto, &item.Precio, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn)
		if err != nil {
			log.Println("Error al escanear producto:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (p ProductoRepository) ObtenerProductoById(ctx context.Context, productoId *int) (*domain.Producto, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/productos/")
	query := `SELECT p.id,p.nombre,p.estado,($1::text || p.id::text || '/' || p.foto) AS url_foto,p.precio,p.creado_en,p.actualizado_en,p.eliminado_en FROM producto p WHERE p.id = $2 LIMIT 1`
	var item domain.Producto
	err := p.pool.QueryRow(ctx, query, fullHostname, *productoId).
		Scan(&item.Id, &item.Nombre, &item.Estado, &item.UrlFoto, &item.Precio, &item.CreadoEn, &item.ActualizadoEn, &item.EliminadoEn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Producto no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}
func NewProductoRepository(pool *pgxpool.Pool) *ProductoRepository {
	return &ProductoRepository{pool: pool}
}

var _ port.ProductoRepository = (*ProductoRepository)(nil)
