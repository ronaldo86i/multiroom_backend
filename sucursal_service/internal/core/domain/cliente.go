package domain

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type ClienteInfo struct {
	Id              int         `json:"id"`
	Nombres         string      `json:"nombres"`
	Apellidos       string      `json:"apellidos"`
	FechaNacimiento pgtype.Date `json:"fechaNacimiento"`
	CodigoPais      string      `json:"codigoPais"`
	Celular         string      `json:"celular"`
	Estado          string      `json:"estado"`
	CreadoEn        time.Time   `json:"creadoEn"`
}
