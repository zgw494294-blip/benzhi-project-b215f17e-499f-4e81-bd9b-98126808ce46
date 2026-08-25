package transport

import (
	"deepdeploy/internal/domain"
	"errors"
	"net/http"
)

func statusFor(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, domain.ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, domain.ErrConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, domain.ErrInvalidState) || errors.Is(err, domain.ErrFrozen) {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}
