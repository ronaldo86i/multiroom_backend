package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log"
	"mime/multipart"
	"multiroom/sucursal-service/internal/core/domain"
	"multiroom/sucursal-service/internal/core/domain/datatype"
	"multiroom/sucursal-service/internal/core/port"
	"multiroom/sucursal-service/internal/core/util"
	"os"
	"path/filepath"
)

type AppVersionRepository struct {
	pool *pgxpool.Pool
}

func (a AppVersionRepository) ObtenerListaVersiones(ctx context.Context, _ map[string]string) (*[]domain.AppVersion, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/dl/")
	query := `
SELECT av.id,
       av.version,
       av.archivo,
       av.fecha_lanzamiento,
       av.estado,
       av.tipo,
       av.arquitectura,
       av.size,
       av.sha256,
       av.os,
       ($1::text || av.archivo) AS url
FROM app_version av
ORDER BY av.archivo DESC ,av.tipo DESC,av.fecha_lanzamiento DESC`
	rows, err := a.pool.Query(ctx, query, fullHostname)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	defer rows.Close()
	list := make([]domain.AppVersion, 0)
	for rows.Next() {
		var item domain.AppVersion
		err = rows.Scan(&item.Id, &item.Version, &item.Archivo, &item.FechaLanzamiento,
			&item.Estado, &item.Tipo, &item.Arquitectura, &item.Size, &item.Sha256, &item.Os, &item.Url)
		if err != nil {
			log.Println("Error al escanear app_version:", err)
			return nil, datatype.NewInternalServerErrorGeneric()
		}
		list = append(list, item)
	}
	return &list, nil
}

func (a AppVersionRepository) ObtenerUltimaVersion(ctx context.Context, q *domain.AppLastVersionQuery) (*domain.AppVersion, error) {
	fullHostname, ok := ctx.Value("fullHostname").(string)
	if !ok || fullHostname == "" {
		log.Println("No se encontró fullHostname en el contexto")
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/dl/")

	query := `
	SELECT av.id,
	       av.version,
	       av.archivo,
	       av.fecha_lanzamiento,
	       av.estado,
	       av.tipo,
	       av.arquitectura,
	       av.size,
	       av.sha256,
	       av.os,
	       ($1::text || av.archivo) AS url
	FROM app_version av
	WHERE av.os = $2
	  AND av.tipo = $3
	  AND av.arquitectura = $4
	  AND av.estado = 'activo'
	ORDER BY av.fecha_lanzamiento DESC
	LIMIT 1`

	var item domain.AppVersion
	err := a.pool.QueryRow(ctx, query, fullHostname, q.Os, q.Tipo, q.Arquitectura).Scan(&item.Id, &item.Version, &item.Archivo, &item.FechaLanzamiento, &item.Estado, &item.Tipo, &item.Arquitectura, &item.Size, &item.Sha256, &item.Os, &item.Url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Versión no encontrada")
		}
		log.Println("Error al obtener última versión:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	return &item, nil
}

func (a AppVersionRepository) RegistrarApp(ctx context.Context, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) (*int, error) {
	// Abrir archivo
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error al abrir archivo")
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	// Creamos el hasher
	hasher := sha256.New()
	// Copiamos el contenido del archivo al hasher
	_, err = io.Copy(hasher, file)
	if err != nil {
		log.Println("Error al leer archivo:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	newSha256 := hex.EncodeToString(hasher.Sum(nil))

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		log.Println("Error al iniciar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	var committed bool
	defer func() {
		_ = file.Close()
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	var appVersionId int
	query := `INSERT INTO app_version(archivo, version, fecha_lanzamiento, estado, tipo, arquitectura, size, sha256) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`
	err = tx.QueryRow(ctx, query, request.NombreArchivo, request.Version, request.FechaLanzamiento, "Activo", request.Tipo, request.Arquitectura, fileHeader.Size, newSha256).Scan(&appVersionId)
	if err != nil {
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	routeDir := fmt.Sprintf("./public/uploads/dl")
	// Guardar archivo
	err = util.File.SaveFile(routeDir, fileHeader.Filename, file)
	if err != nil {
		log.Println("Error al guardar archivo:", err)
		return nil, datatype.NewInternalServerError("Error al guardar archivo, inténtelo más tarde")
	}
	// Confirmar transacción
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("Error al confirmar transacción:", err)
		return nil, datatype.NewInternalServerErrorGeneric()
	}

	committed = true
	return &appVersionId, nil
}

func (a AppVersionRepository) ModificarVersion(ctx context.Context, id *int, request *domain.AppVersionRequest, fileHeader *multipart.FileHeader) error {
	// Abrir archivo
	file, err := fileHeader.Open()
	if err != nil {
		log.Println("Error al abrir archivo")
		return datatype.NewInternalServerErrorGeneric()
	}
	defer func() {
		_ = file.Close()

	}()

	// Crear SHA256 del nuevo archivo
	hasher := sha256.New()
	if _, err = io.Copy(hasher, file); err != nil {
		log.Println("Error al leer archivo:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	newSha256 := hex.EncodeToString(hasher.Sum(nil))

	// Obtener versión actual (para luego borrar el archivo viejo)
	actualVersion, err := a.ObtenerVersion(ctx, id)
	if err != nil {
		return err
	}

	// Reiniciar el reader porque ya lo usamos para el hash
	_, _ = file.Seek(0, io.SeekStart)

	// Iniciar transacción
	tx, err := a.pool.Begin(ctx)
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

	query := `
		UPDATE app_version
		SET archivo = $1,
		    tipo = $2,
		    os = $3,
		    size = $4,
		    sha256 = $5,
		    arquitectura = $6,
		    fecha_lanzamiento = $7,
		    version = $8
		WHERE id = $9`
	_, err = tx.Exec(ctx, query,
		fileHeader.Filename,
		request.Tipo,
		request.Os,
		fileHeader.Size,
		newSha256,
		request.Arquitectura,
		request.FechaLanzamiento,
		request.Version,
		*id,
	)
	if err != nil {
		return datatype.NewInternalServerErrorGeneric()
	}

	// Guardar nuevo archivo
	routeDir := "./public/uploads/dl"
	err = util.File.SaveFile(routeDir, fileHeader.Filename, file)
	if err != nil {
		return err
	}

	// Confirmar transacción
	if err = tx.Commit(ctx); err != nil {
		log.Println("Error al confirmar transacción:", err)
		return datatype.NewInternalServerErrorGeneric()
	}
	committed = true

	// Borrar archivo viejo (si existía y no es el mismo que el nuevo)
	if actualVersion.Archivo != "" && actualVersion.Archivo != fileHeader.Filename {
		oldPath := filepath.Join(routeDir, actualVersion.Archivo)
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			log.Println("⚠️ Error al borrar archivo viejo:", err)
		}
	}

	return nil
}

func (a AppVersionRepository) ObtenerVersion(ctx context.Context, id *int) (*domain.AppVersion, error) {
	fullHostname := ctx.Value("fullHostname").(string)
	fullHostname = fmt.Sprintf("%s%s", fullHostname, "/uploads/dl/")
	query := `
SELECT av.id,
       av.version,
       av.archivo,
       av.fecha_lanzamiento,
       av.estado,
       av.tipo,
       av.arquitectura,
       av.size,
       av.sha256,
       av.os,
       ($1::text || av.archivo) AS url
FROM app_version av WHERE $2`
	var item domain.AppVersion
	err := a.pool.QueryRow(ctx, query, fullHostname, *id).Scan(&item.Id, &item.Version, &item.Archivo, &item.FechaLanzamiento,
		&item.Estado, &item.Tipo, &item.Arquitectura, &item.Size, &item.Sha256, &item.Os, &item.Url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, datatype.NewNotFoundError("Versión no encontrada")
		}
		return nil, datatype.NewInternalServerErrorGeneric()
	}
	return &item, nil
}

func NewAppVersionRepository(pool *pgxpool.Pool) *AppVersionRepository {
	return &AppVersionRepository{pool: pool}
}

var _ port.AppVersionRepository = (*AppVersionRepository)(nil)
