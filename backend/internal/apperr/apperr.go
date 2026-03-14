package apperr

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

func New(status int, message string) *Error {
	return &Error{Status: status, Message: message}
}

func NotFound(msg string) *Error      { return New(http.StatusNotFound, msg) }
func Conflict(msg string) *Error      { return New(http.StatusConflict, msg) }
func Forbidden(msg string) *Error     { return New(http.StatusForbidden, msg) }
func Unprocessable(msg string) *Error { return New(http.StatusUnprocessableEntity, msg) }

func HTTPError(err error) (int, string) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Status, ae.Message
	}
	return 0, ""
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
