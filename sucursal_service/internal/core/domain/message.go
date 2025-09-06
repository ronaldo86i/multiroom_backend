package domain

type MessageData[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message"`
}
