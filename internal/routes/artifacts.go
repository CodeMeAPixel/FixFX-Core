package routes

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func RegisterArtifactsRoutes(api fiber.Router) {
	artifacts := api.Group("/artifacts")

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
	// @Success 200 {object} ArtifactsResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /artifacts/fetch [get]
	artifacts.Get("/fetch", handlers.GetArtifacts)

	// @Summary Get artifact by version
	// @Description Fetch a specific artifact version
	// @Tags Artifacts
	// @Produce json
	// @Param version path string true "Artifact version"
	// @Param platform query string false "Platform filter"
	// @Success 200 {array} Artifact
	// @Failure 404 {object} ErrorResponse
	// @Router /artifacts/version/{version} [get]
	artifacts.Get("/version/:version", handlers.GetArtifactByVersion)

	// @Summary Check artifact status
	// @Description Check if an artifact version is available and its status
	// @Tags Artifacts
	// @Produce json
	// @Param version query string true "Version to check"
	// @Success 200 {object} ArtifactStatus
	// @Failure 404 {object} ErrorResponse
	// @Router /artifacts/check [get]
	artifacts.Get("/check", handlers.CheckArtifact)

	// @Summary Get artifact changes
	// @Description Get changelog/changes for artifacts
	// @Tags Artifacts
	// @Produce json
	// @Param limit query int false "Number of entries" default(20)
	// @Success 200 {array} ChangelogEntry
	// @Router /artifacts/changes [get]
	artifacts.Get("/changes", handlers.GetArtifactChanges)
}
