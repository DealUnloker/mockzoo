package restapi

import (
	"net/http"

	"github.com/DealUnloker/mockzoo/docs"
	"github.com/DealUnloker/mockzoo/internal/controller/restapi/middleware"
	v1 "github.com/DealUnloker/mockzoo/internal/controller/restapi/v1"
	"github.com/DealUnloker/mockzoo/internal/usecase"
	"github.com/DealUnloker/mockzoo/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// NewRouter wires the sandbox pet API.
//
// The OpenAPI 3.1 specification served at /v1/openapi.json is hand-maintained at
// docs/openapi.json (embedded via docs.OpenAPIJSON) and is the single source of
// truth for this surface: it must be kept in sync with the handlers by hand,
// there is no codegen step here.
func NewRouter(app *fiber.App, p usecase.Pet, l logger.Interface) {
	// Options
	app.Use(middleware.Logger(l))
	app.Use(middleware.Recovery(l))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// API reference UI (Scalar, loaded from CDN, pointed at /v1/openapi.json)
	app.Get("/docs", func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)

		return ctx.Send(docs.ReferenceHTML)
	})

	// Routers
	apiV1Group := app.Group("/v1")
	{
		v1.NewRoutes(apiV1Group, p, l)
	}
}
