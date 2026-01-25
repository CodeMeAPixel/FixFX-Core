package handlers

import (
	"strconv"

	"github.com/CodeMeAPixel/FixFX-Core/internal/services"
	"github.com/gofiber/fiber/v2"
)

var nativesService *services.NativesService

// InitNativesHandler initializes the natives handler with a service instance
func InitNativesHandler() {
	nativesService = services.NewNativesService()
}

// GetNatives handles GET /api/natives
// @Summary Get game natives
// @Description Fetch game natives with filtering and search
// @Tags Natives
// @Produce json
// @Param game query string false "Game (gta5/rdr3)" default(gta5)
// @Param environment query string false "Environment (client/server/shared/all)" default(all)
// @Param namespace query string false "Namespace filter"
// @Param search query string false "Search term"
// @Param limit query int false "Results per page" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Param includeCfx query boolean false "Include CFX natives" default(true)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /natives [get]
func GetNatives(c *fiber.Ctx) error {
	gameStr := c.Query("game", "gta5")
	game := services.GameType(gameStr)

	envStr := c.Query("environment", "all")
	env := services.EnvironmentType(envStr)

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := services.NativesQuery{
		Game:         game,
		Environment:  env,
		Namespace:    c.Query("namespace", ""),
		Search:       c.Query("search", ""),
		Limit:        limit,
		Offset:       offset,
		IncludeCfx:   c.QueryBool("includeCfx", true),
		FullMetadata: c.QueryBool("fullMetadata", false),
	}

	response, err := nativesService.GetNatives(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"count":    len(response.Data),
		"data":     response.Data,
		"metadata": response.Metadata,
	})
}

// SearchNatives handles GET /api/natives/search
// @Summary Search natives
// @Description Search across all native functions
// @Tags Natives
// @Produce json
// @Param q query string true "Search query"
// @Param game query string false "Game (gta5/rdr3)" default(gta5)
// @Param environment query string false "Environment filter"
// @Param limit query int false "Results per page" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /natives/search [get]
func SearchNatives(c *fiber.Ctx) error {
	searchQuery := c.Query("q", "")
	if searchQuery == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query (q) parameter is required",
		})
	}

	gameStr := c.Query("game", "gta5")
	game := services.GameType(gameStr)

	envStr := c.Query("environment", "all")
	env := services.EnvironmentType(envStr)

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	query := services.NativesQuery{
		Game:        game,
		Environment: env,
		Search:      searchQuery,
		Limit:       limit,
		Offset:      offset,
		IncludeCfx:  c.QueryBool("includeCfx", true),
	}

	response, err := nativesService.GetNatives(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"count":    len(response.Data),
		"query":    searchQuery,
		"data":     response.Data,
		"metadata": response.Metadata,
	})
}

// GetNativeByHash handles GET /api/natives/:hash
// @Summary Get native by hash
// @Description Fetch a specific native by its hash
// @Tags Natives
// @Produce json
// @Param hash path string true "Native hash"
// @Param game query string false "Game (gta5/rdr3)" default(gta5)
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /natives/{hash} [get]
func GetNativeByHash(c *fiber.Ctx) error {
	hash := c.Params("hash")
	if hash == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Hash parameter is required",
		})
	}

	gameStr := c.Query("game", "gta5")
	game := services.GameType(gameStr)

	query := services.NativesQuery{
		Game:       game,
		IncludeCfx: c.QueryBool("includeCfx", true),
	}

	// Fetch all natives
	response, err := nativesService.GetNatives(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Find by hash
	for _, native := range response.Data {
		if native.Hash == hash {
			return c.JSON(fiber.Map{
				"success": true,
				"data":    native,
			})
		}
	}

	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error": "Native hash not found",
		"hash":  hash,
	})
}

// GetNativesStats handles GET /api/natives/stats
// @Summary Get natives statistics
// @Description Get statistics about available natives
// @Tags Natives
// @Produce json
// @Param game query string false "Game (gta5/rdr3)" default(gta5)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /natives/stats [get]
func GetNativesStats(c *fiber.Ctx) error {
	gameStr := c.Query("game", "gta5")
	game := services.GameType(gameStr)

	query := services.NativesQuery{
		Game:       game,
		IncludeCfx: c.QueryBool("includeCfx", true),
		Limit:      10000, // Get all for stats
	}

	response, err := nativesService.GetNatives(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Calculate statistics
	stats := fiber.Map{
		"total":          response.Metadata.Total,
		"environments":   response.Metadata.EnvironmentStats,
		"namespaces":     len(response.Metadata.Namespaces),
		"games":          response.Metadata.Games,
	}

	// Count by namespace
	namespaceCounts := make(map[string]int)
	for _, native := range response.Data {
		namespaceCounts[native.NS]++
	}

	stats["byNamespace"] = namespaceCounts

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}
