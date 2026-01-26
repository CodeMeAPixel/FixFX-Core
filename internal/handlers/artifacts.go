package handlers

import (
	"strconv"

	"github.com/CodeMeAPixel/FixFX-Core/internal/services"
	"github.com/gofiber/fiber/v2"
)

var artifactsService *services.ArtifactsService

// InitArtifactsHandler initializes the artifacts handler with a service instance
func InitArtifactsHandler(githubToken string) {
	artifactsService = services.NewArtifactsService(githubToken)
}

// GetArtifacts handles GET /api/artifacts/fetch
// @Summary Get artifacts
// @Description Fetch FiveM/RedM server artifacts with filtering and pagination
// @Tags Artifacts
// @Produce json
// @Param platform query string false "Platform (windows/linux/all)" default(all)
// @Param version query string false "Specific version to fetch"
// @Param status query string false "Filter by status (recommended/latest/active/deprecated/eol)"
// @Param sortBy query string false "Sort by (version/date)" default(version)
// @Param sortOrder query string false "Sort order (asc/desc)" default(desc)
// @Param limit query int false "Results per page" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Param includeEol query boolean false "Include end-of-life artifacts" default(true)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /artifacts/fetch [get]
func GetArtifacts(c *fiber.Ctx) error {
	query := services.ArtifactsQuery{
		Platform:   c.Query("platform", "all"),
		Version:    c.Query("version", ""),
		Status:     services.SupportStatus(c.Query("status", "")),
		SortBy:     c.Query("sortBy", "version"),
		SortOrder:  c.Query("sortOrder", "desc"),
		IncludeEOL: c.QueryBool("includeEol", true),
	}

	// Parse limit and offset
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query.Limit = limit
	query.Offset = offset

	result, err := artifactsService.GetArtifacts(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	artifacts := result.Data
	totalFiltered := result.Total

	// Get unique platforms and support statuses for metadata
	platformsMap := make(map[string]bool)
	statusesMap := make(map[string]bool)
	for _, artifact := range artifacts {
		platformsMap[string(artifact.Platform)] = true
		statusesMap[string(artifact.SupportStatus)] = true
	}

	platforms := make([]string, 0)
	statuses := make([]string, 0)
	for p := range platformsMap {
		platforms = append(platforms, p)
	}
	for s := range statusesMap {
		statuses = append(statuses, s)
	}

	return c.JSON(fiber.Map{
		"data": artifacts,
		"metadata": fiber.Map{
			"total":           totalFiltered,
			"limit":           limit,
			"offset":          offset,
			"hasMore":         offset+len(artifacts) < totalFiltered,
			"platforms":       platforms,
			"supportStatuses": statuses,
			"stats": fiber.Map{
				"total":       result.Stats.Total,
				"recommended": result.Stats.Recommended,
				"latest":      result.Stats.Latest,
				"active":      result.Stats.Active,
				"deprecated":  result.Stats.Deprecated,
				"eol":         result.Stats.EOL,
			},
			"query": fiber.Map{
				"platform":   query.Platform,
				"version":    query.Version,
				"status":     string(query.Status),
				"limit":      limit,
				"offset":     offset,
				"sortBy":     query.SortBy,
				"sortOrder":  query.SortOrder,
				"includeEol": query.IncludeEOL,
			},
		},
	})
}

// GetArtifactByVersion handles GET /api/artifacts/version/:version
// @Summary Get artifact by version
// @Description Fetch a specific artifact version
// @Tags Artifacts
// @Produce json
// @Param version path string true "Artifact version"
// @Param platform query string false "Platform filter"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /artifacts/version/{version} [get]
func GetArtifactByVersion(c *fiber.Ctx) error {
	version := c.Params("version")
	if version == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Version parameter required",
		})
	}

	query := services.ArtifactsQuery{
		Version:    version,
		Platform:   c.Query("platform", "all"),
		IncludeEOL: true,
	}

	result, err := artifactsService.GetArtifacts(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(result.Data) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Artifact version not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(result.Data),
		"data":    result.Data,
	})
}

// CheckArtifact handles GET /api/artifacts/check
// @Summary Check artifact status
// @Description Check the status and availability of an artifact
// @Tags Artifacts
// @Produce json
// @Param version query string true "Version to check"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /artifacts/check [get]
func CheckArtifact(c *fiber.Ctx) error {
	version := c.Query("version")
	if version == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Version query parameter required",
		})
	}

	query := services.ArtifactsQuery{
		Version:    version,
		IncludeEOL: true,
	}

	result, err := artifactsService.GetArtifacts(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(result.Data) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Artifact version not found",
			"version": version,
		})
	}

	return c.JSON(fiber.Map{
		"success":       true,
		"version":       version,
		"available":     true,
		"platformCount": len(result.Data),
		"platforms":     result.Data,
	})
}

// GetArtifactChanges handles GET /api/artifacts/changes
// @Summary Get artifact changelog
// @Description Get changelog/commits between two artifact versions
// @Tags Artifacts
// @Produce json
// @Param base query string false "Base version for comparison"
// @Param head query string false "Head version for comparison"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /artifacts/changes [get]
func GetArtifactChanges(c *fiber.Ctx) error {
	base := c.Query("base", "")
	head := c.Query("head", "")

	if base == "" || head == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Both 'base' and 'head' version parameters are required",
		})
	}

	// Get both artifacts to validate they exist
	baseQuery := services.ArtifactsQuery{Version: base, IncludeEOL: true}
	headQuery := services.ArtifactsQuery{Version: head, IncludeEOL: true}

	baseResult, err := artifactsService.GetArtifacts(baseQuery)
	if err != nil || len(baseResult.Data) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Base version not found",
		})
	}

	headResult, err := artifactsService.GetArtifacts(headQuery)
	if err != nil || len(headResult.Data) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Head version not found",
		})
	}

	// Return changelog information
	return c.JSON(fiber.Map{
		"success": true,
		"base":    base,
		"head":    head,
		"comparison": map[string]interface{}{
			"base_hash": baseResult.Data[0].Hash,
			"head_hash": headResult.Data[0].Hash,
			"base_date": baseResult.Data[0].Date,
			"head_date": headResult.Data[0].Date,
			"platforms": len(headResult.Data),
		},
		"message": "Use GitHub API to fetch detailed commit history",
	})
}
