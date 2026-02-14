package routes

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func RegisterValidatorRoutes(api fiber.Router) {
	validator := api.Group("/validator")

	// @Summary Validate JSON
	// @Description Validate JSON with optional txAdmin embed/config schema validation
	// @Tags Validator
	// @Accept json
	// @Produce json
	// @Param body body ValidateRequest true "JSON to validate"
	// @Success 200 {object} ValidationResponse
	// @Failure 400 {object} ErrorResponse
	// @Router /validator/validate [post]
	validator.Post("/validate", handlers.ValidateJSON)

	// @Summary Get validator info
	// @Description Get available validation types, placeholders, and schema info
	// @Tags Validator
	// @Produce json
	// @Success 200 {object} ValidatorInfoResponse
	// @Router /validator/info [get]
	validator.Get("/info", handlers.GetValidatorInfo)
}
