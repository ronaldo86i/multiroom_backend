package util

import (
	"strings"
)

// NewMessage crea un mensaje concatenando todas las cadenas proporcionadas.
func NewMessage(message string, messages ...string) map[string]interface{} {
	var builder strings.Builder
	builder.WriteString(message)
	for _, s := range messages {
		builder.WriteString(s)
	}
	return map[string]interface{}{
		"message": builder.String(),
	}
}

// NewMessageData crea un mensaje concatenado y asigna datos genéricos (solo structs).
func NewMessageData[T any](data T, message string, messages ...string) map[string]interface{} {
	var builder strings.Builder
	builder.WriteString(message)
	for _, s := range messages {
		builder.WriteString(s)
	}

	return map[string]interface{}{
		"message": builder.String(),
		"data":    data,
	}
}
