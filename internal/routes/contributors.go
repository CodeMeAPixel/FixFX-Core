package routes

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

// RegisterContributorsRoutes registers all contributors-related routes
func RegisterContributorsRoutes(api fiber.Router) {
	contributors := api.Group("/contributors")

	// GET /api/contributors - Get list of contributors
	contributors.Get("", handlers.GetContributorsHandler)

	// GET /api/contributors/stats - Get contributor statistics
	contributors.Get("/stats", handlers.GetContributorsStatsHandler)
}
