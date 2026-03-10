package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Native types
type NativeParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type Native struct {
	Name               string        `json:"name"`
	Params             []NativeParam `json:"params"`
	Results            string        `json:"results"`
	Description        string        `json:"description"`
	Hash               string        `json:"hash"`
	JHash              string        `json:"jhash,omitempty"`
	NS                 string        `json:"ns"`
	ResultsDescription string        `json:"resultsDescription,omitempty"`
	Environment        string        `json:"environment"`
	APISet             string        `json:"apiset,omitempty"`
	Game               string        `json:"game"`
	IsCfx              bool          `json:"isCfx"`
}

type RawNativeData struct {
	Name               string        `json:"name"`
	Params             []interface{} `json:"params,omitempty"`
	ResultsType        string        `json:"resultsType,omitempty"`
	Results            string        `json:"results,omitempty"`
	Description        string        `json:"description,omitempty"`
	ResultsDescription string        `json:"resultsDescription,omitempty"`
	APISet             string        `json:"apiset,omitempty"`
	JHash              string        `json:"jhash,omitempty"`
}

type RawNativesData map[string]map[string]RawNativeData

type NativeSource string

const (
	GTA5Source NativeSource = "gta5"
	RDR3Source NativeSource = "rdr3"
	CFXSource  NativeSource = "cfx"
)

type GameType string

const (
	GTA5Game GameType = "gta5"
	RDR3Game GameType = "rdr3"
)

type EnvironmentType string

const (
	ClientEnv EnvironmentType = "client"
	ServerEnv EnvironmentType = "server"
	SharedEnv EnvironmentType = "shared"
	AllEnv    EnvironmentType = "all"
)

type NativesQuery struct {
	Game         GameType
	Environment  EnvironmentType
	Namespace    string
	Search       string
	Limit        int
	Offset       int
	IncludeCfx   bool
	FullMetadata bool
}

type NativesMetadata struct {
	Total                  int                            `json:"total"`
	Limit                  int                            `json:"limit"`
	Offset                 int                            `json:"offset"`
	HasMore                bool                           `json:"hasMore"`
	Namespaces             []string                       `json:"namespaces"`
	NamespacesByGameAndEnv map[string]map[string][]string `json:"namespacesByGameAndEnv"`
	HasCfxNamespace        bool                           `json:"hasCfxNamespace"`
	Games                  []string                       `json:"games"`
	EnvironmentStats       map[string]int                 `json:"environmentStats"`
}

type NativesResponse struct {
	Data     []Native        `json:"data"`
	Metadata NativesMetadata `json:"metadata"`
}

type NativesService struct {
	cache      map[NativeSource]*cacheEntry
	httpClient *http.Client
}

const (
	nativesCacheDuration = 3600000 // 1 hour in milliseconds
)

var endpoints = map[NativeSource]string{
	GTA5Source: "https://runtime.fivem.net/doc/natives.json",
	RDR3Source: "https://runtime.fivem.net/doc/natives_rdr3.json",
	CFXSource:  "https://runtime.fivem.net/doc/natives_cfx.json",
}

func NewNativesService() *NativesService {
	return &NativesService{
		cache: make(map[NativeSource]*cacheEntry),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchNatives fetches natives data for a specific source
func (s *NativesService) FetchNatives(source NativeSource, useCache bool) (RawNativesData, error) {
	if useCache {
		if cached := s.getCache(string(source)); cached != nil {
			if data, ok := cached.(RawNativesData); ok {
				return data, nil
			}
		}
	}

	endpoint, exists := endpoints[source]
	if !exists {
		return nil, fmt.Errorf("unknown native source: %s", source)
	}

	var rawData RawNativesData
	if err := s.fetchJSON(endpoint, &rawData); err != nil {
		return nil, fmt.Errorf("failed to fetch %s natives: %w", source, err)
	}

	s.setCache(string(source), rawData, nativesCacheDuration)
	return rawData, nil
}

// FetchMultiple fetches multiple native sources in parallel
func (s *NativesService) FetchMultiple(sources []NativeSource) map[NativeSource]RawNativesData {
	results := make(map[NativeSource]RawNativesData)

	for _, source := range sources {
		data, err := s.FetchNatives(source, true)
		if err != nil {
			fmt.Printf("Error fetching %s natives: %v\n", source, err)
			results[source] = nil
		} else {
			results[source] = data
		}
	}

	return results
}

// ProcessRawData converts raw native data into Native objects
func (s *NativesService) ProcessRawData(rawData RawNativesData, source NativeSource) []Native {
	var result []Native

	for namespace, nativesMap := range rawData {
		for hash, nativeData := range nativesMap {
			if nativeData.Name == "" {
				continue
			}

			params := []NativeParam{}
			if nativeData.Params != nil {
				for _, param := range nativeData.Params {
					// Convert param to proper structure
					if paramMap, ok := param.(map[string]interface{}); ok {
						paramName := ""
						if name, ok := paramMap["name"].(string); ok {
							paramName = name
						}
						paramType := "any"
						if paramTypeVal, ok := paramMap["type"].(string); ok {
							paramType = paramTypeVal
						}
						paramDesc := ""
						if desc, ok := paramMap["description"].(string); ok {
							paramDesc = desc
						}
						params = append(params, NativeParam{
							Name:        paramName,
							Type:        paramType,
							Description: paramDesc,
						})
					}
				}
			}

			resultsType := nativeData.ResultsType
			if resultsType == "" {
				resultsType = nativeData.Results
			}
			if resultsType == "" {
				resultsType = "void"
			}

			game := GTA5Game
			if source == RDR3Source {
				game = RDR3Game
			}

			native := Native{
				Name:               nativeData.Name,
				Params:             params,
				Results:            resultsType,
				Description:        nativeData.Description,
				Hash:               hash,
				JHash:              nativeData.JHash,
				NS:                 namespace,
				ResultsDescription: nativeData.ResultsDescription,
				Environment:        s.determineEnvironment(nativeData, namespace),
				APISet:             nativeData.APISet,
				Game:               string(game),
				IsCfx:              source == CFXSource || namespace == "CFX",
			}

			result = append(result, native)
		}
	}

	return result
}

// ProcessMultipleSources processes natives from multiple sources
func (s *NativesService) ProcessMultipleSources(rawDataSources map[NativeSource]RawNativesData) []Native {
	var allNatives []Native

	for source, rawData := range rawDataSources {
		if rawData != nil {
			processed := s.ProcessRawData(rawData, source)
			allNatives = append(allNatives, processed...)
		}
	}

	return allNatives
}

// FilterNatives filters natives based on query
func (s *NativesService) FilterNatives(natives []Native, query NativesQuery) []Native {
	filtered := natives

	// Game filter (CFX natives work for both games)
	filtered = filterNativeSlice(filtered, func(n Native) bool {
		if !n.IsCfx && n.Game != string(query.Game) {
			return false
		}
		return true
	})

	// Environment filter
	if query.Environment != AllEnv && query.Environment != "" {
		filtered = filterNativeSlice(filtered, func(n Native) bool {
			if n.Environment != string(query.Environment) && n.Environment != string(SharedEnv) {
				return false
			}
			return true
		})
	}

	// Namespace filter
	if query.Namespace != "" {
		filtered = filterNativeSlice(filtered, func(n Native) bool {
			return n.NS == query.Namespace
		})
	}

	// Search filter
	if query.Search != "" {
		searchLower := strings.ToLower(strings.TrimSpace(query.Search))
		filtered = filterNativeSlice(filtered, func(n Native) bool {
			searchableText := strings.ToLower(
				n.Name + " " + n.Description + " " + n.ResultsDescription,
			)

			// Include params in search text
			for _, param := range n.Params {
				searchableText += " " + strings.ToLower(param.Name+" "+param.Description)
			}

			return strings.Contains(searchableText, searchLower)
		})
	}

	return filtered
}

// SortByRelevance sorts natives by search relevance
func (s *NativesService) SortByRelevance(natives []Native, searchQuery string) []Native {
	sorted := make([]Native, len(natives))
	copy(sorted, natives)

	if searchQuery == "" {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
		return sorted
	}

	searchLower := strings.ToLower(strings.TrimSpace(searchQuery))

	sort.Slice(sorted, func(i, j int) bool {
		aScore := s.calculateRelevanceScore(sorted[i], searchLower)
		bScore := s.calculateRelevanceScore(sorted[j], searchLower)

		if aScore != bScore {
			return aScore > bScore
		}

		// If equal scores, sort alphabetically
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

// PaginateNatives applies pagination
func (s *NativesService) PaginateNatives(natives []Native, limit, offset int) []Native {
	if offset < 0 || offset >= len(natives) {
		return []Native{}
	}

	end := offset + limit
	if end > len(natives) {
		end = len(natives)
	}

	return natives[offset:end]
}

// GetNatives returns filtered and paginated natives
func (s *NativesService) GetNatives(query NativesQuery) (NativesResponse, error) {
	// Set defaults
	if query.Game == "" {
		query.Game = GTA5Game
	}
	if query.Limit == 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}

	// Determine sources to fetch
	sources := []NativeSource{NativeSource(query.Game)}
	if query.IncludeCfx {
		sources = append(sources, CFXSource)
	}

	// Fetch raw data
	rawDataSources := s.FetchMultiple(sources)

	// Process raw data
	allNatives := s.ProcessMultipleSources(rawDataSources)

	// Apply game+cfx filter only (no env/namespace/search) for metadata:
	// this gives us the full picture of what's available for this game selection.
	metaQuery := NativesQuery{
		Game:        query.Game,
		IncludeCfx:  query.IncludeCfx,
		Environment: AllEnv,
	}
	gameFiltered := s.FilterNatives(allNatives, metaQuery)

	// Apply all filters for the actual paginated response
	filtered := s.FilterNatives(allNatives, query)

	// Sort by relevance
	sorted := s.SortByRelevance(filtered, query.Search)

	// Paginate
	paginated := s.PaginateNatives(sorted, query.Limit, query.Offset)

	// Build metadata from game-filtered set (not search/namespace filtered)
	namespaceSet := make(map[string]bool)
	environmentStats := make(map[string]int)
	hasCfxNamespace := false
	nsForEnv := make(map[string]map[string]bool) // env -> ns -> bool

	for _, n := range gameFiltered {
		namespaceSet[n.NS] = true
		environmentStats[n.Environment]++
		if n.IsCfx || n.NS == "CFX" {
			hasCfxNamespace = true
		}
		env := n.Environment
		if nsForEnv[env] == nil {
			nsForEnv[env] = make(map[string]bool)
		}
		nsForEnv[env][n.NS] = true
		if nsForEnv["all"] == nil {
			nsForEnv["all"] = make(map[string]bool)
		}
		nsForEnv["all"][n.NS] = true
	}
	environmentStats["total"] = len(gameFiltered)

	namespaces := make([]string, 0, len(namespaceSet))
	for ns := range namespaceSet {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	// Build namespacesByGameAndEnv: game -> env -> sorted []namespaces
	namespacesByGameAndEnv := make(map[string]map[string][]string)
	gameKey := string(query.Game)
	namespacesByGameAndEnv[gameKey] = make(map[string][]string)
	for env, nsMap := range nsForEnv {
		nsList := make([]string, 0, len(nsMap))
		for ns := range nsMap {
			nsList = append(nsList, ns)
		}
		sort.Strings(nsList)
		namespacesByGameAndEnv[gameKey][env] = nsList
	}

	metadata := NativesMetadata{
		Total:                  len(sorted),
		Limit:                  query.Limit,
		Offset:                 query.Offset,
		HasMore:                query.Offset+query.Limit < len(sorted),
		Namespaces:             namespaces,
		NamespacesByGameAndEnv: namespacesByGameAndEnv,
		HasCfxNamespace:        hasCfxNamespace,
		EnvironmentStats:       environmentStats,
		Games:                  []string{string(query.Game)},
	}

	return NativesResponse{
		Data:     paginated,
		Metadata: metadata,
	}, nil
}

// Helper methods

func (s *NativesService) determineEnvironment(nativeData RawNativeData, namespace string) string {
	// Use explicit apiset if provided
	if nativeData.APISet != "" {
		switch strings.ToLower(nativeData.APISet) {
		case "server":
			return "server"
		case "shared":
			return "shared"
		default:
			return "client"
		}
	}

	// Server-side namespaces
	serverNamespaces := []string{"NETWORK", "PLAYER_SV", "SERVER"}
	for _, srvNs := range serverNamespaces {
		if strings.Contains(namespace, srvNs) {
			return "server"
		}
	}

	// For CFX namespace, analyze function name patterns
	if namespace == "CFX" {
		nameUpper := strings.ToUpper(nativeData.Name)

		// Client-side patterns
		clientPatterns := []string{"NUI_", "SCREEN", "MINIMAP", "CAMERA", "AUDIO", "TEXTURE", "DRAW", "STREAMING"}
		for _, pattern := range clientPatterns {
			if strings.Contains(nameUpper, pattern) {
				return "client"
			}
		}

		// Server-side patterns
		serverPatterns := []string{"PLAYER", "RESOURCE", "CONVAR", "EVENT", "ENTITY"}
		for _, pattern := range serverPatterns {
			if strings.Contains(nameUpper, pattern) && !strings.Contains(nameUpper, "LOCAL") {
				return "server"
			}
		}
	}

	return "client" // Default
}

func (s *NativesService) calculateRelevanceScore(native Native, searchLower string) int {
	score := 0

	// Exact name match - highest priority
	if strings.ToLower(native.Name) == searchLower {
		score += 4
	}

	// Name contains search term
	if strings.Contains(strings.ToLower(native.Name), searchLower) {
		score += 3
	}

	// Description contains search term
	if strings.Contains(strings.ToLower(native.Description), searchLower) {
		score += 2
	}

	// Parameters contain search term
	for _, param := range native.Params {
		if strings.Contains(strings.ToLower(param.Name), searchLower) ||
			strings.Contains(strings.ToLower(param.Description), searchLower) {
			score += 1
			break
		}
	}

	return score
}

func (s *NativesService) getCache(key string) interface{} {
	if entry, exists := s.cache[NativeSource(key)]; exists {
		if time.Now().UnixMilli()-entry.timestamp < nativesCacheDuration {
			return entry.data
		}
		delete(s.cache, NativeSource(key))
	}
	return nil
}

func (s *NativesService) setCache(key string, data interface{}, duration int64) {
	s.cache[NativeSource(key)] = &cacheEntry{
		data:      data,
		timestamp: time.Now().UnixMilli(),
	}
}

func (s *NativesService) fetchJSON(url string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "FixFX-Core/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// Generic filter helper
func filterNativeSlice(slice []Native, predicate func(Native) bool) []Native {
	var filtered []Native
	for _, item := range slice {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
