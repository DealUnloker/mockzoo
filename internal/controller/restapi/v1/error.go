package v1

import (
	"github.com/DealUnloker/mockzoo/internal/controller/restapi/v1/response"
	"github.com/gofiber/fiber/v2"
)

func errorResponse(ctx *fiber.Ctx, status int, code, msg string) error {
	return ctx.Status(status).JSON(response.Error{
		Error: response.ErrorBody{Code: code, Message: msg},
	})
}
