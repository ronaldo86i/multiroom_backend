package util

import (
	"multiroom/sucursal-service/internal/core/domain/datatype"

	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// file es una estructura que proporciona métodos para manejar operaciones de archivo.
type file struct{}

// File es una instancia de la estructura file, que se puede utilizar para acceder a sus métodos.
var File file

// MakeDir crea un directorio en la ruta especificada si no existe.
func (file) MakeDir(ruta string) error {
	err := os.MkdirAll(ruta, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error creando el directorio: %v", err)
	}
	return nil
}

// DeleteAllFiles elimina todos los archivos en un directorio específico.
// Si hay un error al eliminar un archivo, retorna el error.
func (file) DeleteAllFiles(dir string) error {
	// Abrir el directorio
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func(d *os.File) {
		err := d.Close()
		if err != nil {

		}
	}(d)

	// Leer todos los nombres de archivo en el directorio
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	// Iterar sobre todos los nombres de archivo y eliminarlos
	for _, name := range names {
		err = os.Remove(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func (file) DeleteAllFilesWithException(dir, fileName string) error {
	// Abrir el directorio
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func(d *os.File) {
		err := d.Close()
		if err != nil {

		}
	}(d)

	// Leer todos los nombres de archivo en el directorio
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	// Iterar sobre todos los nombres de archivo y eliminarlos
	for _, name := range names {
		if name == fileName {
			continue
		}
		err = os.Remove(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func (file) Copy(src, dst string) error {
	// Obtener el directorio de destino
	destDir := filepath.Dir(dst)

	// Crear la ruta de destino si no existe
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creando directorio %s: %w", destDir, err)
	}

	// Verificar que el archivo origen existe y es regular
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error obteniendo información del archivo origen: %w", err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s no es un archivo regular", src)
	}

	// Abrir el archivo origen
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error abriendo archivo origen: %w", err)
	}
	defer func(source *os.File) {
		_ = source.Close()

	}(source)

	// Crear el archivo destino
	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creando archivo destino: %w", err)
	}
	defer func(destination *os.File) {
		_ = destination.Close()

	}(destination)

	// Copiar contenido
	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("error copiando archivo: %w", err)
	}

	return nil
}

func (file) SaveFile(destino string, nombreArchivo string, file multipart.File) error {
	// Construir la ruta completa del archivo
	filePath := filepath.Join(destino, nombreArchivo)
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// Verificar si el archivo ya existe
	if _, err := os.Stat(filePath); err == nil {
		// Puedes manejar el caso en que el archivo ya existe si quieres
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error verificando el archivo: %v", err)
	}

	// Crear el archivo de destino
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creando el archivo: %v", err)
	}
	defer func(dst *os.File) {
		if err := dst.Close(); err != nil {
			fmt.Printf("error cerrando el archivo: %v\n", err)
		}
	}(dst)

	// Copiar el contenido al archivo de destino
	if _, err := io.Copy(dst, file); err != nil {
		return fmt.Errorf("error guardando el contenido del archivo: %v", err)
	}

	return nil
}

func (file) ValidarTipoArchivo(fileName string, typesValid ...string) bool {
	extFile := strings.ToLower(filepath.Ext(fileName)) // Obtiene la extensión del archivo y la convierte a minúsculas.
	for _, ext := range typesValid {
		if extFile == strings.ToLower(ext) { // Compara insensiblemente a mayúsculas/minúsculas.
			return true
		}
	}
	return false
}

// VerificarArchivoExistente Verifica si el archivo existe en la ruta especificada
func VerificarArchivoExistente(route string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(route)
	if err != nil {
		// Comprobar si el error es porque el archivo no existe
		if os.IsNotExist(err) {
			return nil, &datatype.ErrorResponse{
				Code:    fiber.StatusNotFound,
				Message: "Archivo no encontrado",
			}
		}
		// Otros errores al acceder al archivo
		return nil, &datatype.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Error al acceder al archivo",
		}
	}
	return fileInfo, nil
}

// AbrirArchivo Abre el archivo y devuelve un puntero al mismo
func AbrirArchivo(route string) (*os.File, error) {
	file, err := os.Open(route)
	if err != nil {
		// Comprobar si el error es porque el archivo no existe
		if os.IsNotExist(err) {
			return nil, &datatype.ErrorResponse{
				Code:    fiber.StatusNotFound,
				Message: "Archivo no encontrado",
			}
		}
		// Otros errores al abrir el archivo
		return nil, &datatype.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Error al abrir el archivo",
		}
	}
	return file, nil
}

// CerrarArchivo Cierra el archivo de forma segura
func CerrarArchivo(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Println("Error al cerrar el archivo:", err)
	}
}

// ObtenerTipoContenido Obtiene el tipo de contenido del archivo
func ObtenerTipoContenido(file *os.File) (string, error) {
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}
	// Volver al inicio del archivo
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}

// EnviarArchivoCompleto Envía el archivo completo si no se especificó un rango
func EnviarArchivoCompleto(c *fiber.Ctx, file *os.File, route string, fileSize int64, contentType string) error {
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "inline; filename=\""+filepath.Base(route)+"\"")
	c.Set("Content-Length", fmt.Sprintf("%d", fileSize))

	if _, err := io.Copy(c.Response().BodyWriter(), file); err != nil {
		return &datatype.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Error al enviar el archivo",
		}
	}
	return nil
}

// EnviarArchivoPorRango Envía el archivo en el rango especificado
func EnviarArchivoPorRango(c *fiber.Ctx, file *os.File, fileSize int64, contentType, rangeHeader string) error {
	var rangeStart, rangeEnd int64 = 0, fileSize - 1

	// Parsear el encabezado de rango
	if strings.HasPrefix(rangeHeader, "bytes=") {
		ranges := strings.Split(rangeHeader[6:], "-")
		if start, err := strconv.ParseInt(ranges[0], 10, 64); err == nil {
			rangeStart = start
		}
		if len(ranges) > 1 && ranges[1] != "" {
			if end, err := strconv.ParseInt(ranges[1], 10, 64); err == nil {
				rangeEnd = end
			}
		}
	}

	// Validar límites del rango
	if rangeStart >= fileSize || rangeEnd >= fileSize || rangeStart > rangeEnd {
		c.Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		return &datatype.ErrorResponse{
			Code:    fiber.StatusRequestedRangeNotSatisfiable,
			Message: "Rango solicitado no válido",
		}
	}

	// Ajustar encabezados de respuesta para contenido parcial
	c.Set("Content-Type", contentType)
	c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeStart, rangeEnd, fileSize))
	c.Set("Content-Length", fmt.Sprintf("%d", rangeEnd-rangeStart+1))
	c.Status(fiber.StatusPartialContent)

	// Posicionar el archivo en el inicio del rango y enviar el contenido
	if _, err := file.Seek(rangeStart, io.SeekStart); err != nil {
		return &datatype.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Error al buscar en el archivo",
		}
	}
	if _, err := io.CopyN(c.Response().BodyWriter(), file, rangeEnd-rangeStart+1); err != nil {
		return &datatype.ErrorResponse{
			Code:    fiber.StatusInternalServerError,
			Message: "Error al enviar el rango del archivo",
		}
	}

	return nil
}

func EnviarArchivo(ctx *fiber.Ctx) error {

	path, err := url.PathUnescape(ctx.Path())
	if err != nil {
		log.Println("Archivo no encontrado:", err)
		return ctx.SendStatus(http.StatusNotFound)
	}
	log.Println("Solicitando archivo:", path)

	// Convertir el path al archivo local
	filePath := "./public" + path // Esto asegura que estás accediendo al sistema de archivos local
	log.Println("Solicitando archivo:", filePath)
	// Verificar si el archivo existe
	fileInfo, err := VerificarArchivoExistente(filePath)
	if err != nil {
		log.Println("Archivo no encontrado:", err)
		return ctx.SendStatus(http.StatusNotFound)
	}

	// Abrir el archivo
	file, err := AbrirArchivo(filePath)
	if err != nil {
		log.Println("Error abriendo archivo:", err)
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	defer CerrarArchivo(file)

	// Obtener el tipo de contenido
	contentType, err := ObtenerTipoContenido(file)
	if err != nil {
		log.Println("Error obteniendo tipo de contenido:", err)
		return ctx.SendStatus(http.StatusInternalServerError)
	}
	log.Println("Contenido:", contentType)
	// Forzar tipo MIME correcto si es .apk
	if strings.HasSuffix(filePath, ".apk") {
		contentType = "application/vnd.android.package-archive"
		ctx.Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	}

	// Obtener el rango solicitado por el cliente
	rangeHeader := ctx.Get("Range")
	if rangeHeader == "" {
		// Enviar el archivo completo
		return EnviarArchivoCompleto(ctx, file, filePath, fileInfo.Size(), contentType)
	}

	// Enviar el archivo por rangos
	return EnviarArchivoPorRango(ctx, file, fileInfo.Size(), contentType, rangeHeader)
}

// BackupFiles crea una copia en memoria de los archivos de un directorio
func (file) BackupFiles(dir string) (map[string][]byte, error) {
	backup := make(map[string][]byte)

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return backup, nil
		}
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		backup[file.Name()] = data
	}

	return backup, nil
}

func (file) BackupFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// RestoreFiles borra archivos actuales y restaura desde un backup
func (file) RestoreFiles(backup map[string][]byte, dir string) error {
	// Eliminar archivos actuales
	files, err := os.ReadDir(dir)
	if err == nil {
		for _, file := range files {
			_ = os.Remove(filepath.Join(dir, file.Name()))
		}
	}

	// Restaurar archivos
	for name, data := range backup {
		err := os.WriteFile(filepath.Join(dir, name), data, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteFiles elimina una lista de archivos de un directorio
func (file) DeleteFiles(dir string, filenames []string) {
	for _, name := range filenames {
		_ = os.Remove(filepath.Join(dir, name))
	}
}
