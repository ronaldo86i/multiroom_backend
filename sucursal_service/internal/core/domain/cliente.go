package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type ClienteInfo struct {
	Id              int64       `json:"id"`
	Nombres         string      `json:"nombres"`
	Apellidos       string      `json:"apellidos"`
	FechaNacimiento pgtype.Date `json:"fechaNacimiento"`
	CodigoPais      string      `json:"codigoPais"`
	Celular         string      `json:"celular"`
	Estado          string      `json:"estado"`
	CreadoEn        time.Time   `json:"creadoEn"`
}
