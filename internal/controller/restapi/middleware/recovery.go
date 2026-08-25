package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/DealUnloker/mockzoo/internal/controller/restapi/v1/response"
	"github.com/DealUnloker/mockzoo/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

const _codeInternal = "internal"

func buildPanicMessage(ctx *fiber.Ctx, err any) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(ctx.OriginalURL())
	result.WriteString(" PANIC DETECTED: ")
	fmt.Fprintf(&result, "%v\n%s\n", err, debug.Stack())

	return result.String()
}

// Recovery catches panics from downstream handlers. It logs the full panic
// value and stack trace server-side, then responds with the standard
// {"error":{"code","message"}} envelope itself — the panic value is never
// echoed back to the client.
func Recovery(l logger.Interface) func(c *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				l.Error(buildPanicMessage(ctx, r))

				err = ctx.Status(http.StatusInternalServerError).JSON(response.Error{
					Error: response.ErrorBody{Code: _codeInternal, Message: "internal server error"},
				})
			}
		}()

		return ctx.Next()
	}
}
