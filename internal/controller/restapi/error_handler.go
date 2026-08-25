package restapi

import (
	"errors"
	"net/http"

	"github.com/DealUnloker/mockzoo/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

const (
	_codeNotFound         = "not_found"
	_codeMethodNotAllowed = "method_not_allowed"
	_codeInternal         = "internal"
)

// NewErrorHandler renders every error that escapes route handlers as the
// standard {"error":{"code","message"}} envelope, instead of Fiber's
// text/plain defaults ("Cannot GET ..."). It mainly covers unmatched routes
// and disallowed methods: application errors are already turned into the
// envelope by v1.handleError, and panics are turned into the envelope by the
// recovery middleware before they ever reach here.
func NewErrorHandler() fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			switch fiberErr.Code {
			case fiber.StatusNotFound:
				return errorEnvelope(ctx, http.StatusNotFound, _codeNotFound, "route not found")
			case fiber.StatusMethodNotAllowed:
				return errorEnvelope(ctx, http.StatusMethodNotAllowed, _codeMethodNotAllowed, "method not allowed")
			}
		}

		return errorEnvelope(ctx, http.StatusInternalServerError, _codeInternal, "internal server error")
	}
}

func errorEnvelope(ctx *fiber.Ctx, status int, code, msg string) error {
	return ctx.Status(status).JSON(response.Error{
		Error: response.ErrorBody{Code: code, Message: msg},
	})
}
