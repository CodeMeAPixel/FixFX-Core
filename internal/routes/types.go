package routes

// Type definitions for Swagger documentation

// ArtifactsResponse represents the artifacts fetch response
// @Description Response containing paginated artifacts with metadata
type ArtifactsResponse struct {
	Data     []Artifact        `json:"data"`
	Metadata ArtifactsMetadata `json:"metadata"`
}

// Artifact represents a single FiveM/RedM artifact
type Artifact struct {
	Version       string `json:"version"`
	Platform      string `json:"platform"`
	Hash          string `json:"hash"`
	URL           string `json:"url"`
	Size          int64  `json:"size"`
	Date          string `json:"date"`
	SupportStatus string `json:"supportStatus"`
}

// ArtifactsMetadata contains pagination and filter information
type ArtifactsMetadata struct {
	Total           int      `json:"total"`
	Limit           int      `json:"limit"`
	Offset          int      `json:"offset"`
	HasMore         bool     `json:"hasMore"`
	Platforms       []string `json:"platforms"`
	SupportStatuses []string `json:"supportStatuses"`
}

// ArtifactStatus represents the status of a specific artifact
type ArtifactStatus struct {
	Version       string `json:"version"`
	Status        string `json:"status"`
	Available     bool   `json:"available"`
	SupportStatus string `json:"supportStatus"`
}

// ChangelogEntry represents a changelog entry
type ChangelogEntry struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Changes string `json:"changes"`
}

// NativesResponse represents the natives fetch response
type NativesResponse struct {
	Data     []Native        `json:"data"`
	Metadata NativesMetadata `json:"metadata"`
}

// Native represents a game native function
type Native struct {
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	Game        string `json:"game"`
	Environment string `json:"environment"`
	Deprecated  bool   `json:"deprecated"`
}

// NativesMetadata contains information about natives
type NativesMetadata struct {
	Total        int      `json:"total"`
	Games        []string `json:"games"`
	Namespaces   []string `json:"namespaces"`
	Environments []string `json:"environments"`
}

// NativesStats represents statistics about natives
type NativesStats struct {
	TotalNatives    int            `json:"totalNatives"`
	ByGame          map[string]int `json:"byGame"`
	ByEnvironment   map[string]int `json:"byEnvironment"`
	DeprecatedCount int            `json:"deprecatedCount"`
}

// SourceResponse represents the source file read response
type SourceResponse struct {
	Code     string `json:"code"`
	Path     string `json:"path"`
	Language string `json:"language"`
	Size     int    `json:"size"`
}

// SearchResponse represents the search results
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Query   string         `json:"query"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	Score       float64 `json:"score"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
	Path   string `json:"path,omitempty"`
}

// ValidateRequest represents the validator input
type ValidateRequest struct {
	JSON string `json:"json"`
	Type string `json:"type"`
}

// ValidationIssueResponse represents a single validation issue
type ValidationIssueResponse struct {
	Path     string `json:"path"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// ValidationResponse represents the validation result
type ValidationResponse struct {
	Valid      bool                      `json:"valid"`
	Type       string                    `json:"type"`
	Issues     []ValidationIssueResponse `json:"issues"`
	Formatted  string                    `json:"formatted,omitempty"`
	ParseError string                    `json:"parseError,omitempty"`
}

// ValidatorInfoResponse represents validator metadata
type ValidatorInfoResponse struct {
	Types        []map[string]string `json:"types"`
	Placeholders []map[string]string `json:"placeholders"`
	Limits       map[string]any      `json:"limits"`
}
