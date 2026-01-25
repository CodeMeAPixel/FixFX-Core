package routes

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func RegisterSourceRoutes(api fiber.Router) {
	source := api.Group("/source")

	// @Summary Read source file
	// @Description Securely read and return source file contents with syntax highlighting
	// @Tags Source
	// @Produce json
	// @Param path query string true "File path (relative to project root)"
	// @Security ApiKeyAuth
	// @Success 200 {object} SourceResponse
	// @Failure 400 {object} ErrorResponse "Path not provided or invalid"
	// @Failure 403 {object} ErrorResponse "Path not allowed (not in whitelist)"
	// @Failure 404 {object} ErrorResponse "File not found"
	// @Router /source [get]
	source.Get("", handlers.ReadSourceFile)
}

func RegisterSearchRoutes(api fiber.Router) {
	search := api.Group("/search")

	// @Summary Global search
	// @Description Search across documentation and resources
	// @Tags Search
	// @Produce json
	// @Param q query string true "Search query"
	// @Param searchType query string false "Search type (docs/artifacts/natives/all)" default(all)
	// @Param limit query int false "Results per page" default(20)
	// @Success 200 {object} map[string]interface{}
	// @Router /search [get]
	search.Get("", GlobalSearch)
}

// Placeholder handler for global search - will be implemented
func GlobalSearch(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Global search - coming soon"})
}
