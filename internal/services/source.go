package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SourceService handles secure file serving
type SourceService struct {
	allowedPaths []string
	basePath     string
}

// SourceResponse represents the response for a source file
type SourceResponse struct {
	Code string `json:"code"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// NewSourceService creates a new source service
func NewSourceService(basePath string, allowedPaths []string) *SourceService {
	if basePath == "" {
		var err error
		basePath, err = os.Getwd()
		if err != nil {
			basePath = "."
		}
	}

	if len(allowedPaths) == 0 {
		// Default allowed paths
		allowedPaths = []string{
			"lib/",
			"packages/",
			"internal/",
		}
	}

	return &SourceService{
		allowedPaths: allowedPaths,
		basePath:     basePath,
	}
}

// ReadFile safely reads a file with security validation
func (s *SourceService) ReadFile(filePath string) (*SourceResponse, error) {
	if filePath == "" {
		return nil, fmt.Errorf("no path provided")
	}

	// Security: Prevent path traversal
	if strings.Contains(filePath, "..") {
		return nil, fmt.Errorf("invalid path")
	}

	// Security: Only allow whitelisted paths
	if !s.isPathAllowed(filePath) {
		return nil, fmt.Errorf("path not allowed")
	}

	// Construct absolute path
	absolutePath := filepath.Join(s.basePath, filePath)

	// Verify the resolved path is still within allowed boundaries
	absPath, err := filepath.Abs(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path")
	}

	basePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("invalid base path")
	}

	// Ensure the resolved path is within the base path
	if !strings.HasPrefix(absPath, basePath) {
		return nil, fmt.Errorf("path not allowed")
	}

	// Read the file
	data, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return &SourceResponse{
		Code: string(data),
		Path: filePath,
		Size: int64(len(data)),
	}, nil
}

// isPathAllowed checks if a path is in the whitelist
func (s *SourceService) isPathAllowed(filePath string) bool {
	cleanPath := filepath.Clean(filePath)

	for _, allowed := range s.allowedPaths {
		allowedClean := filepath.Clean(allowed)
		// Check if path starts with an allowed path
		if strings.HasPrefix(cleanPath, allowedClean) {
			return true
		}
		// Also check with forward slashes for consistency
		if strings.HasPrefix(strings.ReplaceAll(cleanPath, "\\", "/"), strings.ReplaceAll(allowedClean, "\\", "/")) {
			return true
		}
	}

	return false
}

// AddAllowedPath adds a path to the whitelist
func (s *SourceService) AddAllowedPath(path string) {
	s.allowedPaths = append(s.allowedPaths, path)
}

// GetAllowedPaths returns the list of allowed paths
func (s *SourceService) GetAllowedPaths() []string {
	return s.allowedPaths
}
