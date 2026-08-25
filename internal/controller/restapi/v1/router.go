package v1

import (
	"github.com/DealUnloker/mockzoo/docs"
	"github.com/DealUnloker/mockzoo/internal/usecase"
	"github.com/DealUnloker/mockzoo/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// NewRoutes -.
func NewRoutes(apiV1Group fiber.Router, p usecase.Pet, l logger.Interface) {
	r := &V1{p: p, l: l, v: validator.New(validator.WithRequiredStructEnabled())}

	apiV1Group.Get("/openapi.json", func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		return ctx.Send(docs.OpenAPIJSON)
	})

	petGroup := apiV1Group.Group("/pets")
	{
		petGroup.Get("/", r.listPets)
		petGroup.Post("/", r.createPet)
		petGroup.Get("/:petId", r.getPetByID)
		petGroup.Patch("/:petId", r.updatePet)
		petGroup.Delete("/:petId", r.deletePet)
	}

	adminGroup := apiV1Group.Group("/admin")
	{
		adminGroup.Post("/reset", r.resetSandbox)
	}
}
