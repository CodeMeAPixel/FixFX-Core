package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ValidationType represents the type of JSON validation to perform
type ValidationType string

const (
	ValidationGeneric     ValidationType = "generic"
	ValidationTxEmbed     ValidationType = "txadmin-embed"
	ValidationTxEmbedConf ValidationType = "txadmin-embed-config"
)

// ValidationSeverity indicates the severity of a validation issue
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// ValidationIssue represents a single validation finding
type ValidationIssue struct {
	Path     string             `json:"path"`
	Message  string             `json:"message"`
	Severity ValidationSeverity `json:"severity"`
}

// ValidationResult holds the complete validation outcome
type ValidationResult struct {
	Valid      bool              `json:"valid"`
	Type       ValidationType    `json:"type"`
	Issues     []ValidationIssue `json:"issues"`
	Formatted  string            `json:"formatted,omitempty"`
	ParseError string            `json:"parseError,omitempty"`
}

// ValidatorService provides JSON validation capabilities
type ValidatorService struct{}

// NewValidatorService creates a new validator service
func NewValidatorService() *ValidatorService {
	return &ValidatorService{}
}

// txAdmin placeholder regex
var txPlaceholderRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

// Valid txAdmin placeholders
var validTxPlaceholders = map[string]bool{
	"serverCfxId":          true,
	"serverJoinUrl":        true,
	"serverBrowserUrl":     true,
	"serverClients":        true,
	"serverMaxClients":     true,
	"serverName":           true,
	"statusColor":          true,
	"statusString":         true,
	"uptime":               true,
	"nextScheduledRestart": true,
}

// Validate performs JSON validation based on the specified type
func (s *ValidatorService) Validate(rawJSON string, validationType ValidationType) *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Type:   validationType,
		Issues: []ValidationIssue{},
	}

	// Trim whitespace
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" {
		result.Valid = false
		result.ParseError = "Empty input: no JSON provided"
		return result
	}

	// Step 1: Parse as generic JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		result.Valid = false
		result.ParseError = formatJSONError(err, rawJSON)
		return result
	}

	// Format the JSON
	formatted, err := json.MarshalIndent(parsed, "", "    ")
	if err == nil {
		result.Formatted = string(formatted)
	}

	// Step 2: Type-specific validation
	switch validationType {
	case ValidationTxEmbed:
		s.validateTxAdminEmbed(parsed, result)
	case ValidationTxEmbedConf:
		s.validateTxAdminEmbedConfig(parsed, result)
	case ValidationGeneric:
		// Generic — just check it parses
	}

	// If any error-level issue exists, mark as invalid
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError {
			result.Valid = false
			break
		}
	}

	return result
}

// validateTxAdminEmbed validates a Discord embed JSON for txAdmin
func (s *ValidatorService) validateTxAdminEmbed(parsed interface{}, result *ValidationResult) {
	obj, ok := parsed.(map[string]interface{})
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$",
			Message:  "Embed JSON must be an object, not an array or primitive",
			Severity: SeverityError,
		})
		return
	}

	// Validate known Discord embed fields
	validTopLevel := map[string]bool{
		"title":       true,
		"description": true,
		"url":         true,
		"timestamp":   true,
		"color":       true,
		"footer":      true,
		"image":       true,
		"thumbnail":   true,
		"video":       true,
		"provider":    true,
		"author":      true,
		"fields":      true,
	}

	for key := range obj {
		if !validTopLevel[key] {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + key,
				Message:  fmt.Sprintf("Unknown embed property \"%s\". Discord may ignore this.", key),
				Severity: SeverityWarning,
			})
		}
	}

	// Info: color and footer overridden by txAdmin
	if _, exists := obj["color"]; exists {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.color",
			Message:  "The \"color\" property is overridden by txAdmin at runtime. Use the Config JSON to set the color.",
			Severity: SeverityInfo,
		})
	}
	if _, exists := obj["footer"]; exists {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.footer",
			Message:  "The \"footer\" property is overridden by txAdmin at runtime and cannot be customized.",
			Severity: SeverityWarning,
		})
	}

	// Validate title
	if title, exists := obj["title"]; exists {
		if str, ok := title.(string); ok {
			if len(str) > 256 {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     "$.title",
					Message:  fmt.Sprintf("Title exceeds Discord's 256 character limit (%d characters)", len(str)),
					Severity: SeverityError,
				})
			}
		}
	}

	// Validate description
	if desc, exists := obj["description"]; exists {
		if str, ok := desc.(string); ok {
			if len(str) > 4096 {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     "$.description",
					Message:  fmt.Sprintf("Description exceeds Discord's 4096 character limit (%d characters)", len(str)),
					Severity: SeverityError,
				})
			}
		}
	}

	// Validate url
	if u, exists := obj["url"]; exists {
		if str, ok := u.(string); ok {
			s.validateURLOrPlaceholder(str, "$.url", result)
		}
	}

	// Validate fields
	if fields, exists := obj["fields"]; exists {
		s.validateEmbedFields(fields, result)
	}

	// Validate image
	if img, exists := obj["image"]; exists {
		s.validateMediaObject(img, "$.image", result)
	}

	// Validate thumbnail
	if thumb, exists := obj["thumbnail"]; exists {
		s.validateMediaObject(thumb, "$.thumbnail", result)
	}

	// Validate author
	if author, exists := obj["author"]; exists {
		s.validateAuthor(author, result)
	}

	// Check for txAdmin placeholders in string values
	s.checkPlaceholders(obj, "$", result)
}

// validateTxAdminEmbedConfig validates txAdmin embed config JSON
func (s *ValidatorService) validateTxAdminEmbedConfig(parsed interface{}, result *ValidationResult) {
	obj, ok := parsed.(map[string]interface{})
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$",
			Message:  "Config JSON must be an object",
			Severity: SeverityError,
		})
		return
	}

	// Required string fields with their descriptions
	requiredStrings := map[string]string{
		"onlineString":  "Text displayed when the server is online",
		"offlineString": "Text displayed when the server is offline",
	}

	// Required color fields
	requiredColors := map[string]string{
		"onlineColor":  "Color shown when the server is online",
		"offlineColor": "Color shown when the server is offline",
	}

	// Optional string fields
	optionalStrings := map[string]string{
		"partialString": "Text displayed when the server is partially online",
	}
	optionalColors := map[string]string{
		"partialColor": "Color shown when the server is partially online",
	}

	// Known fields
	knownFields := map[string]bool{
		"onlineString":  true,
		"onlineColor":   true,
		"partialString": true,
		"partialColor":  true,
		"offlineString": true,
		"offlineColor":  true,
		"buttons":       true,
	}

	// Check for unknown fields
	for key := range obj {
		if !knownFields[key] {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + key,
				Message:  fmt.Sprintf("Unknown config property \"%s\"", key),
				Severity: SeverityWarning,
			})
		}
	}

	// Validate required strings
	for field, desc := range requiredStrings {
		if val, exists := obj[field]; !exists {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + field,
				Message:  fmt.Sprintf("Missing required field \"%s\" (%s)", field, desc),
				Severity: SeverityError,
			})
		} else if _, ok := val.(string); !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + field,
				Message:  fmt.Sprintf("\"%s\" must be a string", field),
				Severity: SeverityError,
			})
		}
	}

	// Validate required colors
	for field, desc := range requiredColors {
		if val, exists := obj[field]; !exists {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + field,
				Message:  fmt.Sprintf("Missing required field \"%s\" (%s)", field, desc),
				Severity: SeverityError,
			})
		} else if str, ok := val.(string); !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + field,
				Message:  fmt.Sprintf("\"%s\" must be a string", field),
				Severity: SeverityError,
			})
		} else if !isValidHexColor(str) {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$." + field,
				Message:  fmt.Sprintf("Invalid hex color \"%s\" — expected format #RRGGBB", str),
				Severity: SeverityError,
			})
		}
	}

	// Validate optional strings
	for field := range optionalStrings {
		if val, exists := obj[field]; exists {
			if _, ok := val.(string); !ok {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     "$." + field,
					Message:  fmt.Sprintf("\"%s\" must be a string", field),
					Severity: SeverityError,
				})
			}
		}
	}

	// Validate optional colors
	for field := range optionalColors {
		if val, exists := obj[field]; exists {
			if str, ok := val.(string); !ok {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     "$." + field,
					Message:  fmt.Sprintf("\"%s\" must be a string", field),
					Severity: SeverityError,
				})
			} else if !isValidHexColor(str) {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     "$." + field,
					Message:  fmt.Sprintf("Invalid hex color \"%s\" — expected format #RRGGBB", str),
					Severity: SeverityError,
				})
			}
		}
	}

	// Validate buttons
	if buttons, exists := obj["buttons"]; exists {
		s.validateConfigButtons(buttons, result)
	}
}

// validateEmbedFields validates the "fields" array in a Discord embed
func (s *ValidatorService) validateEmbedFields(fields interface{}, result *ValidationResult) {
	arr, ok := fields.([]interface{})
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.fields",
			Message:  "\"fields\" must be an array",
			Severity: SeverityError,
		})
		return
	}

	if len(arr) > 25 {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.fields",
			Message:  fmt.Sprintf("Discord allows a maximum of 25 fields, found %d", len(arr)),
			Severity: SeverityError,
		})
	}

	for i, field := range arr {
		path := fmt.Sprintf("$.fields[%d]", i)
		fieldObj, ok := field.(map[string]interface{})
		if !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path,
				Message:  "Each field must be an object with name and value",
				Severity: SeverityError,
			})
			continue
		}

		// name is required
		if name, exists := fieldObj["name"]; !exists {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".name",
				Message:  "Field is missing required \"name\" property",
				Severity: SeverityError,
			})
		} else if str, ok := name.(string); ok && len(str) > 256 {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".name",
				Message:  fmt.Sprintf("Field name exceeds 256 character limit (%d characters)", len(str)),
				Severity: SeverityError,
			})
		}

		// value is required
		if value, exists := fieldObj["value"]; !exists {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".value",
				Message:  "Field is missing required \"value\" property",
				Severity: SeverityError,
			})
		} else if str, ok := value.(string); ok && len(str) > 1024 {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".value",
				Message:  fmt.Sprintf("Field value exceeds 1024 character limit (%d characters)", len(str)),
				Severity: SeverityError,
			})
		}

		// inline is optional, must be boolean
		if inline, exists := fieldObj["inline"]; exists {
			if _, ok := inline.(bool); !ok {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     path + ".inline",
					Message:  "\"inline\" must be a boolean (true or false)",
					Severity: SeverityError,
				})
			}
		}

		// Check for unknown field properties
		validFieldProps := map[string]bool{"name": true, "value": true, "inline": true}
		for key := range fieldObj {
			if !validFieldProps[key] {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     path + "." + key,
					Message:  fmt.Sprintf("Unknown field property \"%s\"", key),
					Severity: SeverityWarning,
				})
			}
		}
	}
}

// validateMediaObject validates image/thumbnail objects
func (s *ValidatorService) validateMediaObject(media interface{}, path string, result *ValidationResult) {
	obj, ok := media.(map[string]interface{})
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     path,
			Message:  fmt.Sprintf("\"%s\" must be an object with a \"url\" property", path),
			Severity: SeverityError,
		})
		return
	}

	if u, exists := obj["url"]; !exists {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     path + ".url",
			Message:  "Missing required \"url\" property",
			Severity: SeverityError,
		})
	} else if str, ok := u.(string); ok {
		s.validateURLOrPlaceholder(str, path+".url", result)
	}
}

// validateAuthor validates the author object
func (s *ValidatorService) validateAuthor(author interface{}, result *ValidationResult) {
	obj, ok := author.(map[string]interface{})
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.author",
			Message:  "\"author\" must be an object",
			Severity: SeverityError,
		})
		return
	}

	if name, exists := obj["name"]; exists {
		if str, ok := name.(string); ok && len(str) > 256 {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     "$.author.name",
				Message:  fmt.Sprintf("Author name exceeds 256 character limit (%d characters)", len(str)),
				Severity: SeverityError,
			})
		}
	} else {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.author.name",
			Message:  "Author is missing required \"name\" property",
			Severity: SeverityError,
		})
	}

	if u, exists := obj["url"]; exists {
		if str, ok := u.(string); ok {
			s.validateURLOrPlaceholder(str, "$.author.url", result)
		}
	}

	if iconURL, exists := obj["icon_url"]; exists {
		if str, ok := iconURL.(string); ok {
			s.validateURLOrPlaceholder(str, "$.author.icon_url", result)
		}
	}
}

// validateConfigButtons validates the buttons array in embed config
func (s *ValidatorService) validateConfigButtons(buttons interface{}, result *ValidationResult) {
	arr, ok := buttons.([]interface{})
	if !ok {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.buttons",
			Message:  "\"buttons\" must be an array",
			Severity: SeverityError,
		})
		return
	}

	if len(arr) > 5 {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     "$.buttons",
			Message:  fmt.Sprintf("txAdmin allows a maximum of 5 buttons, found %d", len(arr)),
			Severity: SeverityError,
		})
	}

	for i, btn := range arr {
		path := fmt.Sprintf("$.buttons[%d]", i)
		btnObj, ok := btn.(map[string]interface{})
		if !ok {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path,
				Message:  "Each button must be an object",
				Severity: SeverityError,
			})
			continue
		}

		// label is required
		if label, exists := btnObj["label"]; !exists {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".label",
				Message:  "Button is missing required \"label\" property",
				Severity: SeverityError,
			})
		} else if str, ok := label.(string); ok && len(str) > 80 {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".label",
				Message:  fmt.Sprintf("Button label exceeds Discord's 80 character limit (%d characters)", len(str)),
				Severity: SeverityError,
			})
		}

		// url is required
		if u, exists := btnObj["url"]; !exists {
			result.Issues = append(result.Issues, ValidationIssue{
				Path:     path + ".url",
				Message:  "Button is missing required \"url\" property",
				Severity: SeverityError,
			})
		} else if str, ok := u.(string); ok {
			s.validateURLOrPlaceholder(str, path+".url", result)
		}

		// emoji is optional
		if emoji, exists := btnObj["emoji"]; exists {
			if _, ok := emoji.(string); !ok {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     path + ".emoji",
					Message:  "\"emoji\" must be a string (unicode emoji or emoji ID)",
					Severity: SeverityError,
				})
			}
		}

		// Check for unknown button properties
		validBtnProps := map[string]bool{"label": true, "url": true, "emoji": true}
		for key := range btnObj {
			if !validBtnProps[key] {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     path + "." + key,
					Message:  fmt.Sprintf("Unknown button property \"%s\"", key),
					Severity: SeverityWarning,
				})
			}
		}
	}
}

// validateURLOrPlaceholder checks if a string is a valid URL or contains txAdmin placeholders
func (s *ValidatorService) validateURLOrPlaceholder(str, path string, result *ValidationResult) {
	// Replace placeholders with dummy values for URL validation
	cleaned := txPlaceholderRegex.ReplaceAllString(str, "https://example.com")

	if !strings.HasPrefix(cleaned, "http://") && !strings.HasPrefix(cleaned, "https://") {
		// Could still be a placeholder-only value
		if txPlaceholderRegex.MatchString(str) {
			return // Pure placeholder is fine
		}
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     path,
			Message:  "URL must start with http:// or https://",
			Severity: SeverityWarning,
		})
		return
	}

	_, err := url.ParseRequestURI(cleaned)
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Path:     path,
			Message:  fmt.Sprintf("Invalid URL: %s", err.Error()),
			Severity: SeverityWarning,
		})
	}
}

// checkPlaceholders recursively validates txAdmin placeholders in strings
func (s *ValidatorService) checkPlaceholders(obj interface{}, path string, result *ValidationResult) {
	switch v := obj.(type) {
	case map[string]interface{}:
		for key, val := range v {
			s.checkPlaceholders(val, path+"."+key, result)
		}
	case []interface{}:
		for i, val := range v {
			s.checkPlaceholders(val, fmt.Sprintf("%s[%d]", path, i), result)
		}
	case string:
		matches := txPlaceholderRegex.FindAllStringSubmatch(v, -1)
		for _, match := range matches {
			placeholder := match[1]
			if !validTxPlaceholders[placeholder] {
				result.Issues = append(result.Issues, ValidationIssue{
					Path:     path,
					Message:  fmt.Sprintf("Unknown txAdmin placeholder \"{{%s}}\"", placeholder),
					Severity: SeverityWarning,
				})
			}
		}
	}
}

// Helper functions

func isValidHexColor(s string) bool {
	if !strings.HasPrefix(s, "#") {
		return false
	}
	hex := s[1:]
	if len(hex) != 3 && len(hex) != 6 {
		return false
	}
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func formatJSONError(err error, rawJSON string) string {
	errStr := err.Error()

	// Try to get a more user-friendly message
	if syntaxErr, ok := err.(*json.SyntaxError); ok {
		offset := syntaxErr.Offset
		line := 1
		col := int(offset)
		for i, ch := range rawJSON {
			if int64(i) >= offset {
				break
			}
			if ch == '\n' {
				line++
				col = int(offset) - i
			}
		}
		return fmt.Sprintf("JSON syntax error at line %d, column %d: %s", line, col, errStr)
	}

	return fmt.Sprintf("Invalid JSON: %s", errStr)
}
