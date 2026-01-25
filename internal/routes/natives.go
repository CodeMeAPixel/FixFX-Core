package routes

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func RegisterNativesRoutes(api fiber.Router) {
	natives := api.Group("/natives")

	// @Summary Get game natives
	// @Description Fetch game natives with filtering and search
	// @Tags Natives
	// @Produce json
	// @Param game query string false "Game type (gta5/rdr3/all)" default(gta5)
	// @Param environment query string false "Environment (server/client/all)" default(all)
	// @Param ns query string false "Namespace filter"
	// @Param search query string false "Search in native names"
	// @Param limit query int false "Results per page" default(50)
	// @Param offset query int false "Pagination offset" default(0)
	// @Param cfx query boolean false "Include CFX natives" default(true)
	// @Param full query boolean false "Include full metadata" default(false)
	// @Success 200 {object} NativesResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /natives [get]
	natives.Get("", handlers.GetNatives)

	// @Summary Search natives
	// @Description Full-text search for game natives
	// @Tags Natives
	// @Produce json
	// @Param query query string true "Search query"
	// @Param limit query int false "Results per page" default(20)
	// @Success 200 {array} Native
	// @Router /natives/search [get]
	natives.Get("/search", handlers.SearchNatives)

	// @Summary Get native by hash
	// @Description Fetch a specific native by its hash
	// @Tags Natives
	// @Produce json
	// @Param hash path string true "Native hash"
	// @Success 200 {object} Native
	// @Failure 404 {object} ErrorResponse
	// @Router /natives/{hash} [get]
	natives.Get("/:hash", handlers.GetNativeByHash)

	// @Summary Get native statistics
	// @Description Get statistics about game natives
	// @Tags Natives
	// @Produce json
	// @Success 200 {object} NativesStats
	// @Router /natives/stats [get]
	natives.Get("/stats", handlers.GetNativesStats)
}
