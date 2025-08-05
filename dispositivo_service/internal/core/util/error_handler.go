package util

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"multiroom/dispositivo-service/internal/core/domain/datatype"
	"net/http"
)

func HandleServiceError(err error) *fiber.Error {
	var errorResponse *datatype.ErrorResponse
	if errors.As(err, &errorResponse) {
		return fiber.NewError(errorResponse.Code, errorResponse.Message)
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
