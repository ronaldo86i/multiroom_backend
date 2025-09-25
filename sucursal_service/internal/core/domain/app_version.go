package domain

import "time"

type EstadoApp string

const (
	EstadoActivo   EstadoApp = "activo"
	EstadoInactivo EstadoApp = "inactivo"
)

type TipoApp string

const (
	TipoAPK    TipoApp = "apk"
	TipoBundle TipoApp = "bundle"
)

type ArquitecturaApp string

const (
	ArqArm64   ArquitecturaApp = "arm64-v8a"
	ArqArmeabi ArquitecturaApp = "armeabi-v7a"
	ArqX86     ArquitecturaApp = "x86"
	ArqX86_64  ArquitecturaApp = "x86_64"
)

type OsApp string

const (
	OsAndroid OsApp = "android"
	OsWindows OsApp = "windows"
	OsMacOS   OsApp = "macos"
	OsLinux   OsApp = "linux"
)

type AppVersion struct {
	Id               int             `json:"id"`
	Version          string          `json:"version"`
	Archivo          string          `json:"-"`
	FechaLanzamiento time.Time       `json:"fechaLanzamiento"`
	Estado           EstadoApp       `json:"estado"`
	Tipo             TipoApp         `json:"tipo"`
	Arquitectura     ArquitecturaApp `json:"arquitectura"`
	Size             int64           `json:"size"`
	Sha256           string          `json:"sha256"`
	Url              *string         `json:"url,omitempty"`
	Os               OsApp           `json:"os"`
}

type AppVersionRequest struct {
	NombreArchivo    string          `json:"nombreArchivo"`
	Version          string          `json:"version"`
	FechaLanzamiento time.Time       `json:"fechaLanzamiento"`
	Tipo             TipoApp         `json:"tipo"`
	Arquitectura     ArquitecturaApp `json:"arquitectura"`
	Os               OsApp           `json:"os"`
}

type AppLastVersionQuery struct {
	Tipo         TipoApp         `json:"tipo"`
	Arquitectura ArquitecturaApp `json:"arquitectura"`
	Os           OsApp           `json:"os"`
}

type AppVersionId struct {
	Id int `json:"id"`
}
