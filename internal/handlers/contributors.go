package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Contributor represents a GitHub contributor
type Contributor struct {
	Login         string `json:"login"`
	ID            int    `json:"id"`
	AvatarURL     string `json:"avatar_url"`
	HTMLURL       string `json:"html_url"`
	Contributions int    `json:"contributions"`
	Type          string `json:"type,omitempty"`
}

// ContributorsHandler handles contributor-related requests
type ContributorsHandler struct {
	githubToken string
	cache       []Contributor
	cacheMutex  sync.RWMutex
	cacheTime   time.Time
	cacheTTL    time.Duration
}

var contributorsHandler *ContributorsHandler

// InitContributorsHandler initializes the contributors handler
func InitContributorsHandler(githubToken string) {
	contributorsHandler = &ContributorsHandler{
		githubToken: githubToken,
		cacheTTL:    24 * time.Hour, // Cache for 24 hours
		cache:       []Contributor{},
	}
}

// GetContributorsHandler handles GET /api/contributors
// @Summary Get contributors
// @Description Fetch merged contributors from the FixFX and FixFX-Core GitHub repositories, sorted by total contributions
// @Tags Contributors
// @Produce json
// @Param limit query int false "Maximum number of contributors to return (0 = all)"
// @Success 200 {array} handlers.Contributor
// @Failure 500 {object} map[string]interface{}
// @Router /contributors [get]
func GetContributorsHandler(c *fiber.Ctx) error {
	if contributorsHandler == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Contributors handler not initialized",
		})
	}
	return contributorsHandler.GetContributors(c)
}

// GetContributorsStatsHandler handles GET /api/contributors/stats
// @Summary Get contributor statistics
// @Description Return aggregate statistics about contributors across the FixFX and FixFX-Core repositories
// @Tags Contributors
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /contributors/stats [get]
func GetContributorsStatsHandler(c *fiber.Ctx) error {
	if contributorsHandler == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Contributors handler not initialized",
		})
	}
	return contributorsHandler.GetContributorStats(c)
}

// GetContributors retrieves contributors from both repositories
func (h *ContributorsHandler) GetContributors(c *fiber.Ctx) error {
	// Check cache
	h.cacheMutex.RLock()
	if len(h.cache) > 0 && time.Since(h.cacheTime) < h.cacheTTL {
		defer h.cacheMutex.RUnlock()
		return c.JSON(h.cache)
	}
	h.cacheMutex.RUnlock()

	// Fetch from both repos
	fixfxContribs := h.fetchRepoContributors("CodeMeAPixel", "FixFX")
	coreContribs := h.fetchRepoContributors("CodeMeAPixel", "FixFX-Core")

	// Merge contributors
	merged := h.mergeContributors(fixfxContribs, coreContribs)

	// Sort by contributions (descending)
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Contributions > merged[j].Contributions
	})

	// Apply limit if specified
	limit := c.QueryInt("limit", 0)
	if limit > 0 && limit < len(merged) {
		merged = merged[:limit]
	}

	// Update cache
	h.cacheMutex.Lock()
	h.cache = merged
	h.cacheTime = time.Now()
	h.cacheMutex.Unlock()

	return c.JSON(merged)
}

// GetContributorStats returns statistics about contributors
func (h *ContributorsHandler) GetContributorStats(c *fiber.Ctx) error {
	contributors := h.getOrFetchContributors()

	stats := fiber.Map{
		"total":              len(contributors),
		"topContributor":     nil,
		"totalContributions": 0,
		"byRepository":       fiber.Map{},
	}

	totalContrib := 0
	for _, contrib := range contributors {
		totalContrib += contrib.Contributions
	}
	stats["totalContributions"] = totalContrib

	if len(contributors) > 0 {
		stats["topContributor"] = contributors[0]
	}

	// Break down by repo
	fixfxContribs := h.fetchRepoContributors("CodeMeAPixel", "FixFX")
	coreContribs := h.fetchRepoContributors("CodeMeAPixel", "FixFX-Core")

	stats["byRepository"] = fiber.Map{
		"FixFX":      len(fixfxContribs),
		"FixFX-Core": len(coreContribs),
	}

	return c.JSON(stats)
}

// fetchRepoContributors fetches contributors from a specific GitHub repository
func (h *ContributorsHandler) fetchRepoContributors(owner, repo string) []Contributor {
	var contributors []Contributor
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors?per_page=%d&page=%d", owner, repo, perPage, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		// Add authentication header if token is available
		if h.githubToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.githubToken))
		}

		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "FixFX-API")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			break
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			break
		}

		var pageContribs []Contributor
		if err := json.Unmarshal(body, &pageContribs); err != nil {
			break
		}

		if len(pageContribs) == 0 {
			break
		}

		contributors = append(contributors, pageContribs...)

		if len(pageContribs) < perPage {
			break
		}

		page++
	}

	return contributors
}

// mergeContributors merges contributors from multiple repos, combining their contributions
func (h *ContributorsHandler) mergeContributors(repos ...[]Contributor) []Contributor {
	mergedMap := make(map[string]*Contributor)

	// Merge all contributors
	for _, repo := range repos {
		for _, contrib := range repo {
			if existing, found := mergedMap[strings.ToLower(contrib.Login)]; found {
				existing.Contributions += contrib.Contributions
			} else {
				contrib := contrib // Create a copy
				mergedMap[strings.ToLower(contrib.Login)] = &contrib
			}
		}
	}

	// Convert back to slice
	result := make([]Contributor, 0, len(mergedMap))
	for _, contrib := range mergedMap {
		result = append(result, *contrib)
	}

	return result
}

// getOrFetchContributors gets contributors from cache or fetches them
func (h *ContributorsHandler) getOrFetchContributors() []Contributor {
	h.cacheMutex.RLock()
	if len(h.cache) > 0 && time.Since(h.cacheTime) < h.cacheTTL {
		defer h.cacheMutex.RUnlock()
		return h.cache
	}
	h.cacheMutex.RUnlock()

	fixfxContribs := h.fetchRepoContributors("CodeMeAPixel", "FixFX")
	coreContribs := h.fetchRepoContributors("CodeMeAPixel", "FixFX-Core")
	merged := h.mergeContributors(fixfxContribs, coreContribs)

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Contributions > merged[j].Contributions
	})

	h.cacheMutex.Lock()
	h.cache = merged
	h.cacheTime = time.Now()
	h.cacheMutex.Unlock()

	return merged
}
