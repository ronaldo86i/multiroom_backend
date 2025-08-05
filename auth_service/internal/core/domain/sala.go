package domain

import "time"

type Sala struct {
	Id       int       `json:"id"`
	Nombre   string    `json:"nombre"`
	CreadoEn time.Time `json:"creadoEn"`
}
