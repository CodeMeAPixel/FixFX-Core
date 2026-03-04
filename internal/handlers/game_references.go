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

// GetBlips returns paginated blip data
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

// GetCheckpoints returns paginated checkpoint data
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

// GetDataFiles returns paginated data file entries
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

// GetGameEvents returns paginated game event entries
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

// GetGamerTags returns gamer tag component entries
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

// GetHUDColors returns HUD color entries
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

// GetMarkers returns marker type data
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

// GetNetGameEvents returns net game event entries
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

// GetPedModels returns paginated ped model entries
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

// GetPickupHashes returns paginated pickup hash entries
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

// GetWeaponModels returns paginated weapon model entries
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

// GetZones returns paginated zone entries
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
