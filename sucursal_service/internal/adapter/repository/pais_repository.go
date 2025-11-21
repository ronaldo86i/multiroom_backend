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

type PaisRepository struct {
	pool *pgxpool.Pool
}

func (p PaisRepository) HabilitarPaisById(ctx context.Context, id *int) error {
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

	query := `UPDATE pais SET estado= 'Activo',actualizado_en=now(),eliminado_en=NULL WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("País no encontrado")
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

func (p PaisRepository) DeshabilitarPaisById(ctx context.Context, id *int) error {
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

	query := `UPDATE pais SET estado= 'Inactivo',actualizado_en=now(),eliminado_en=now() WHERE id = $1`
	ct, err := tx.Exec(ctx, query, *id)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("País no encontrado")
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

func (p PaisRepository) RegistrarPais(ctx context.Context, request *domain.PaisRequest, fileHeader *multipart.FileHeader) (*int, error) {
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

	var paisId int
	query := `INSERT INTO pais(nombre,archivo,estado) VALUES ($1,$2,'Activo') RETURNING id`
	err = tx.QueryRow(ctx, query, request.Nombre, nombreArchivo).Scan(&paisId)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	routeDir := fmt.Sprintf("./public/uploads/paises/%d", paisId)
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
	return &paisId, nil
}

func (p PaisRepository) ModificarPais(ctx context.Context, id *int, request *domain.PaisRequest, fileHeader *multipart.FileHeader) error {
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

	query := `UPDATE pais SET nombre=$1,archivo=$2 WHERE id=$3`
	ct, err := tx.Exec(ctx, query, request.Nombre, nombreArchivo, *id)
	if err != nil {
		log.Println("Error al actualizar país")
		return datatype.NewInternalServerErrorGeneric()
	}
	if ct.RowsAffected() == 0 {
		return datatype.NewNotFoundError("Pais no encontrado")
	}

	// Respaldar archivos existentes
	route := fmt.Sprintf("./public/uploads/paises/%d", *id)
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
	routeDir := fmt.Sprintf("./public/uploads/paises/%d", *id)
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

func (p PaisRepository) ObtenerPaisById(ctx context.Context, id *int) (*domain.PaisDetail, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/paises/")

	query := `SELECT p.id,p.nombre,p.estado,($1::text || p.id::text || '/' || p.archivo) AS url_foto,p.creado_en,p.actualizado_en,p.eliminado_en FROM pais p WHERE p.id=$2 LIMIT 1`
	var pais domain.PaisDetail
	err := p.pool.QueryRow(ctx, query, fullHostname, *id).Scan(&pais.Id, &pais.Nombre, &pais.Estado, &pais.UrlFoto, &pais.CreadoEn, &pais.ActualizadoEn, &pais.EliminadoEn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("País no encontrado")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &pais, nil
}

func (p PaisRepository) ObtenerListaPaises(ctx context.Context, _ map[string]string) (*[]domain.PaisInfo, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/paises/")

	query := `SELECT p.id,p.nombre,p.estado,($1::text || p.id::text || '/' || p.archivo) AS url_foto,p.creado_en FROM pais p`
	rows, err := p.pool.Query(ctx, query, fullHostname)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.PaisInfo, 0)
	for rows.Next() {
		var item domain.PaisInfo
		err = rows.Scan(&item.Id, &item.Nombre, &item.Estado, &item.UrlFoto, &item.CreadoEn)
		if err != nil {
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func NewPaisRepository(pool *pgxpool.Pool) *PaisRepository {
	return &PaisRepository{pool: pool}
}

var _ port.PaisRepository = (*PaisRepository)(nil)
