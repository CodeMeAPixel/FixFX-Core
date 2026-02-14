package handlers

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/services"
	"github.com/gofiber/fiber/v2"
)

var validatorService *services.ValidatorService

// InitValidatorHandler initializes the validator handler
func InitValidatorHandler() {
	validatorService = services.NewValidatorService()
}

// ValidateJSON handles POST /api/validator/validate
// @Summary Validate JSON
// @Description Validate JSON with optional txAdmin embed/config schema validation
// @Tags Validator
// @Accept json
// @Produce json
// @Param body body ValidateRequest true "JSON to validate"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /validator/validate [post]
func ValidateJSON(c *fiber.Ctx) error {
	var req struct {
		JSON string `json:"json"`
		Type string `json:"type"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.JSON == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "JSON input is required",
		})
	}

	// Determine validation type
	validationType := services.ValidationGeneric
	switch req.Type {
	case "txadmin-embed":
		validationType = services.ValidationTxEmbed
	case "txadmin-embed-config":
		validationType = services.ValidationTxEmbedConf
	case "generic", "":
		validationType = services.ValidationGeneric
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid validation type. Valid types: generic, txadmin-embed, txadmin-embed-config",
		})
	}

	result := validatorService.Validate(req.JSON, validationType)

	return c.JSON(fiber.Map{
		"result": result,
	})
}

// GetValidatorInfo handles GET /api/validator/info
// @Summary Get validator information
// @Description Get available validation types, txAdmin placeholders, and schema info
// @Tags Validator
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /validator/info [get]
func GetValidatorInfo(c *fiber.Ctx) error {
	placeholders := []map[string]string{
		{"name": "serverCfxId", "description": "The Cfx.re id of your server, tied to sv_licenseKey"},
		{"name": "serverJoinUrl", "description": "The direct join URL (e.g., https://cfx.re/join/xxxxxx)"},
		{"name": "serverBrowserUrl", "description": "The FiveM Server browser URL"},
		{"name": "serverClients", "description": "Number of players online"},
		{"name": "serverMaxClients", "description": "The sv_maxclients value"},
		{"name": "serverName", "description": "The txAdmin-given server name"},
		{"name": "statusColor", "description": "Hex-encoded color from Config JSON"},
		{"name": "statusString", "description": "Status text from Config JSON"},
		{"name": "uptime", "description": "How long the server has been online"},
		{"name": "nextScheduledRestart", "description": "When the next scheduled restart is"},
	}

	types := []map[string]string{
		{"value": "generic", "label": "Generic JSON", "description": "Validates JSON syntax only"},
		{"value": "txadmin-embed", "label": "txAdmin Embed JSON", "description": "Validates Discord embed JSON for txAdmin status embed"},
		{"value": "txadmin-embed-config", "label": "txAdmin Embed Config", "description": "Validates txAdmin embed configuration (colors, buttons)"},
	}

	return c.JSON(fiber.Map{
		"types":        types,
		"placeholders": placeholders,
		"limits": fiber.Map{
			"embed": fiber.Map{
				"title":       256,
				"description": 4096,
				"fields":      25,
				"fieldName":   256,
				"fieldValue":  1024,
			},
			"config": fiber.Map{
				"buttons":     5,
				"buttonLabel": 80,
			},
		},
	})
}
