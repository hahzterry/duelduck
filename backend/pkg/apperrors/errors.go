package apperrors

import "net/http"

func BadRequest(message string, err ...error) error {
	return newError(http.StatusBadRequest, message, err...)
}

func Unauthorized(message string, err ...error) error {
	return newError(http.StatusUnauthorized, message, err...)
}

func PaymentRequired(message string, err ...error) error {
	return newError(http.StatusPaymentRequired, message, err...)
}

const (
	StatusLoginTimeout = 440
)

func Forbidden(message string, err ...error) error {
	return newError(http.StatusForbidden, message, err...)
}

func NotFound(message string, err ...error) error {
	return newError(http.StatusNotFound, message, err...)
}

func Gone(message string, err ...error) error {
	return newError(http.StatusGone, message, err...)
}

func AlreadyExist(message string, err ...error) error {
	return newError(http.StatusConflict, message, err...)
}

func TooManyRequests(message string, err ...error) error {
	return newError(http.StatusTooManyRequests, message, err...)
}

func LoginTimeout(message string, err ...error) error {
	return newError(StatusLoginTimeout, message, err...)
}

func Internal(message string, err ...error) error {
	return newError(http.StatusInternalServerError, message, err...)
}

func ServiceUnavailable(message string, err ...error) error {
	return newError(http.StatusServiceUnavailable, message, err...)
}

func IsPaymentRequiredError(err error) bool {
	appErr, ok := IsAppError(err)
	if !ok {
		return false
	}

	return appErr.Status == http.StatusPaymentRequired
}

func IsForbiddenError(err error) bool {
	appErr, ok := IsAppError(err)
	if !ok {
		return false
	}
	return appErr.Status == http.StatusForbidden
}
