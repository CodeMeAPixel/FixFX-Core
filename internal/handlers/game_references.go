package handlers

import (
	"math"
	"strconv"
	"strings"

	"github.com/CodeMeAPixel/FixFX-Core/internal/services"
	"github.com/gofiber/fiber/v2"
)

var gameRefService *services.GameReferencesService

// InitGameReferencesHandler initialises the game references handler
func InitGameReferencesHandler() {
	gameRefService = services.NewGameReferencesService()
}

// ────────────────────────────────────────────────
// Pagination helpers
// ────────────────────────────────────────────────

func getIntQuery(c *fiber.Ctx, key string, def, min, max int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil || v < min {
		return def
	}
	if max > 0 && v > max {
		return max
	}
	return v
}

func searchLower(s, q string) bool {
	if q == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(q))
}

func clampPage(total, limit, offset int) (page []int, hasMore bool, safeMeta services.RefMetadata) {
	if offset >= total {
		offset = int(math.Max(0, float64(total-limit)))
	}
	end := offset + limit
	if end > total {
		end = total
	}
	hasMore = end < total
	safeMeta = services.RefMetadata{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
	indices := make([]int, 0, end-offset)
	for i := offset; i < end; i++ {
		indices = append(indices, i)
	}
	return indices, hasMore, safeMeta
}

// ────────────────────────────────────────────────
// GET /api/game-references/blips
// ────────────────────────────────────────────────

// GetBlips handles GET /api/game-references/blips
// @Summary Get map blips
// @Description Fetch all minimap blip sprites and the blip colour palette with optional search and pagination
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by blip name"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/blips [get]
func GetBlips(c *fiber.Ctx) error {
	blips, colors, err := gameRefService.GetBlips()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := blips[:0:0]
	for _, b := range blips {
		if searchLower(b.Name, search) {
			filtered = append(filtered, b)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}
	page := filtered[offset:end]

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(page),
		"data": fiber.Map{
			"blips":  page,
			"colors": colors,
		},
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/checkpoints
// ────────────────────────────────────────────────

// GetCheckpoints handles GET /api/game-references/checkpoints
// @Summary Get checkpoint types
// @Description Fetch all CREATE_CHECKPOINT type IDs with labels, optionally filtered by section
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by checkpoint ID or label"
// @Param section query string false "Filter by section (standard/type-44-46)"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/checkpoints [get]
func GetCheckpoints(c *fiber.Ctx) error {
	checkpoints, err := gameRefService.GetCheckpoints()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	section := c.Query("section") // "standard" | "type-44-46"
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := checkpoints[:0:0]
	for _, cp := range checkpoints {
		if section != "" && cp.Section != section {
			continue
		}
		if searchLower(cp.ID, search) || searchLower(cp.Label, search) {
			filtered = append(filtered, cp)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/data-files
// ────────────────────────────────────────────────

// GetDataFiles handles GET /api/game-references/data-files
// @Summary Get resource manifest data file types
// @Description Fetch all data_file keys used in resource manifests with file type, root element, and mounter details
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by key, file type, or mounter"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/data-files [get]
func GetDataFiles(c *fiber.Ctx) error {
	files, err := gameRefService.GetDataFiles()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := files[:0:0]
	for _, f := range files {
		if searchLower(f.Key, search) || searchLower(f.FileType, search) || searchLower(f.Mounter, search) {
			filtered = append(filtered, f)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/game-events
// ────────────────────────────────────────────────

// GetGameEvents handles GET /api/game-references/game-events
// @Summary Get client-side game events
// @Description Fetch all client-side game events available for resource scripting with descriptions
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by event name or description"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/game-events [get]
func GetGameEvents(c *fiber.Ctx) error {
	events, err := gameRefService.GetGameEvents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := events[:0:0]
	for _, e := range events {
		if searchLower(e.Name, search) || searchLower(e.Description, search) {
			filtered = append(filtered, e)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/gamer-tags
// ────────────────────────────────────────────────

// GetGamerTags handles GET /api/game-references/gamer-tags
// @Summary Get gamer tag components
// @Description Fetch all head display component IDs for SET_MULTIPLAYER_HANGER_COLOUR and related natives
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by component name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/gamer-tags [get]
func GetGamerTags(c *fiber.Ctx) error {
	components, err := gameRefService.GetGamerTagComponents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	filtered := components[:0:0]
	for _, comp := range components {
		if searchLower(comp.Name, search) {
			filtered = append(filtered, comp)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered),
		"data":    filtered,
		"metadata": services.RefMetadata{
			Total:   len(filtered),
			Limit:   len(filtered),
			Offset:  0,
			HasMore: false,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/hud-colors
// ────────────────────────────────────────────────

// GetHUDColors handles GET /api/game-references/hud-colors
// @Summary Get HUD colors
// @Description Fetch all HUD colour indices with RGBA values and hex codes for use with HUD_COLOUR_* constants
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by color name or hex value"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/hud-colors [get]
func GetHUDColors(c *fiber.Ctx) error {
	colors, err := gameRefService.GetHUDColors()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := colors[:0:0]
	for _, col := range colors {
		if searchLower(col.Name, search) || searchLower(col.Hex, search) {
			filtered = append(filtered, col)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/markers
// ────────────────────────────────────────────────

// GetMarkers handles GET /api/game-references/markers
// @Summary Get DRAW_MARKER types
// @Description Fetch all DRAW_MARKER type IDs with name labels
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by marker name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/markers [get]
func GetMarkers(c *fiber.Ctx) error {
	markers, err := gameRefService.GetMarkers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	filtered := markers[:0:0]
	for _, m := range markers {
		if searchLower(m.Name, search) {
			filtered = append(filtered, m)
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered),
		"data":    filtered,
		"metadata": services.RefMetadata{
			Total:   len(filtered),
			Limit:   len(filtered),
			Offset:  0,
			HasMore: false,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/net-game-events
// ────────────────────────────────────────────────

// GetNetGameEvents handles GET /api/game-references/net-game-events
// @Summary Get net game events
// @Description Fetch all GTA_EVENT_IDS enum entries with sequential IDs for network event handling
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by event name"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/net-game-events [get]
func GetNetGameEvents(c *fiber.Ctx) error {
	events, err := gameRefService.GetNetGameEvents()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := events[:0:0]
	for _, e := range events {
		if searchLower(e.Name, search) {
			filtered = append(filtered, e)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/ped-models
// ────────────────────────────────────────────────

// GetPedModels handles GET /api/game-references/ped-models
// @Summary Get pedestrian models
// @Description Fetch all pedestrian model names grouped by category, suitable for use with REQUEST_MODEL and GET_HASH_KEY
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by model name or category"
// @Param category query string false "Filter by category name"
// @Param limit query int false "Results per page" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/ped-models [get]
func GetPedModels(c *fiber.Ctx) error {
	peds, err := gameRefService.GetPedModels()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	category := c.Query("category")
	limit := getIntQuery(c, "limit", 50, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := peds[:0:0]
	for _, p := range peds {
		if category != "" && !strings.EqualFold(p.Category, category) {
			continue
		}
		if searchLower(p.Name, search) || searchLower(p.Category, search) {
			filtered = append(filtered, p)
		}
	}

	// Collect unique categories for metadata
	catSet := make(map[string]struct{})
	for _, p := range peds {
		catSet[p.Category] = struct{}{}
	}
	cats := make([]string, 0, len(catSet))
	for cat := range catSet {
		cats = append(cats, cat)
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": fiber.Map{
			"total":      total,
			"limit":      limit,
			"offset":     offset,
			"hasMore":    end < total,
			"search":     search,
			"categories": cats,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/pickup-hashes
// ────────────────────────────────────────────────

// GetPickupHashes handles GET /api/game-references/pickup-hashes
// @Summary Get pickup hashes
// @Description Fetch all ePickupHashes enum entries with numeric hash values for use with CREATE_PICKUP_ROTATE and related natives
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by pickup name or hash"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/pickup-hashes [get]
func GetPickupHashes(c *fiber.Ctx) error {
	pickups, err := gameRefService.GetPickupHashes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := pickups[:0:0]
	for _, p := range pickups {
		if searchLower(p.Name, search) || searchLower(p.Hash, search) {
			filtered = append(filtered, p)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/weapon-models
// ────────────────────────────────────────────────

// GetWeaponModels handles GET /api/game-references/weapon-models
// @Summary Get weapon models
// @Description Fetch all weapon model names grouped by type with hash keys, DLC info, components, and tints
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by weapon name, hash, or group"
// @Param group query string false "Filter by weapon group"
// @Param limit query int false "Results per page" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/weapon-models [get]
func GetWeaponModels(c *fiber.Ctx) error {
	weapons, err := gameRefService.GetWeaponModels()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	group := c.Query("group")
	limit := getIntQuery(c, "limit", 50, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := weapons[:0:0]
	for _, w := range weapons {
		if group != "" && !strings.EqualFold(w.Group, group) {
			continue
		}
		if searchLower(w.Name, search) || searchLower(w.Hash, search) || searchLower(w.Group, search) {
			filtered = append(filtered, w)
		}
	}

	// Collect unique groups for metadata
	groupSet := make(map[string]struct{})
	for _, w := range weapons {
		groupSet[w.Group] = struct{}{}
	}
	groups := make([]string, 0, len(groupSet))
	for g := range groupSet {
		groups = append(groups, g)
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": fiber.Map{
			"total":   total,
			"limit":   limit,
			"offset":  offset,
			"hasMore": end < total,
			"search":  search,
			"groups":  groups,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/zones
// ────────────────────────────────────────────────

// GetZones handles GET /api/game-references/zones
// @Summary Get map zones
// @Description Fetch all 1300+ map zone name IDs with descriptions for area detection scripting
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by zone name, ID, or description"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/zones [get]
func GetZones(c *fiber.Ctx) error {
	zones, err := gameRefService.GetZones()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := zones[:0:0]
	for _, z := range zones {
		if searchLower(z.ZoneName, search) || searchLower(z.ZoneNameID, search) || searchLower(z.Description, search) {
			filtered = append(filtered, z)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// min is a Go 1.21 builtin but kept here for compatibility
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ────────────────────────────────────────────────
// GET /api/game-references/vehicle-models
// ────────────────────────────────────────────────

// GetVehicleModels handles GET /api/game-references/vehicle-models
// @Summary Get vehicle models
// @Description Fetch all GTA V / FiveM vehicle model names and hashes grouped by category, suitable for use with REQUEST_MODEL and GET_HASH_KEY
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by display name, model name, or category"
// @Param category query string false "Filter by vehicle category"
// @Param limit query int false "Results per page" default(50)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/vehicle-models [get]
func GetVehicleModels(c *fiber.Ctx) error {
	vehicles, err := gameRefService.GetVehicleModels()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	category := c.Query("category")
	limit := getIntQuery(c, "limit", 50, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := vehicles[:0:0]
	for _, v := range vehicles {
		if category != "" && !strings.EqualFold(v.Category, category) {
			continue
		}
		if searchLower(v.DisplayName, search) || searchLower(v.ModelName, search) || searchLower(v.Category, search) {
			filtered = append(filtered, v)
		}
	}

	catSet := make(map[string]struct{})
	for _, v := range vehicles {
		catSet[v.Category] = struct{}{}
	}
	cats := make([]string, 0, len(catSet))
	for cat := range catSet {
		cats = append(cats, cat)
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": fiber.Map{
			"total":      total,
			"limit":      limit,
			"offset":     offset,
			"hasMore":    end < total,
			"search":     search,
			"categories": cats,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/vehicle-colours
// ────────────────────────────────────────────────

// GetVehicleColours handles GET /api/game-references/vehicle-colours
// @Summary Get vehicle colours
// @Description Fetch all vehicle paint colour indices grouped by type (metallic, matte, metals, unnamed) for use with SET_VEHICLE_COLOURS and related natives
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by colour name or type"
// @Param type query string false "Filter by colour type (metallic/matte/metals/unnamed)"
// @Param limit query int false "Results per page" default(200)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/vehicle-colours [get]
func GetVehicleColours(c *fiber.Ctx) error {
	colours, err := gameRefService.GetVehicleColours()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	colourType := c.Query("type") // metallic | matte | metals | unnamed
	limit := getIntQuery(c, "limit", 200, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := colours[:0:0]
	for _, col := range colours {
		if colourType != "" && !strings.EqualFold(col.Type, colourType) {
			continue
		}
		if searchLower(col.Name, search) || searchLower(col.Type, search) {
			filtered = append(filtered, col)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}

// ────────────────────────────────────────────────
// GET /api/game-references/vehicle-flags
// ────────────────────────────────────────────────

// GetVehicleFlags handles GET /api/game-references/vehicle-flags
// @Summary Get vehicle flags
// @Description Fetch all vehicle flag definitions with descriptions and the build version they were introduced in
// @Tags Game References
// @Produce json
// @Param search query string false "Filter by flag name or description"
// @Param limit query int false "Results per page" default(100)
// @Param offset query int false "Pagination offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /game-references/vehicle-flags [get]
func GetVehicleFlags(c *fiber.Ctx) error {
	flags, err := gameRefService.GetVehicleFlags()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"status": fiber.StatusInternalServerError,
		})
	}

	search := c.Query("search")
	limit := getIntQuery(c, "limit", 100, 1, 500)
	offset := getIntQuery(c, "offset", 0, 0, 0)

	filtered := flags[:0:0]
	for _, f := range flags {
		if searchLower(f.Name, search) || searchLower(f.Description, search) {
			filtered = append(filtered, f)
		}
	}

	total := len(filtered)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = 0
		end = min(limit, total)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(filtered[offset:end]),
		"data":    filtered[offset:end],
		"metadata": services.RefMetadata{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
			Search:  search,
		},
	})
}
