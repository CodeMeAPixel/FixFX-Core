package handlers

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/services"
	"github.com/gofiber/fiber/v2"
)

var sourceService *services.SourceService

// InitSourceHandler initializes the source handler with a service instance
func InitSourceHandler(basePath string, allowedPaths []string) {
	sourceService = services.NewSourceService(basePath, allowedPaths)
}

// ReadSourceFile handles GET /api/source
// @Summary Read source file
// @Description Securely read and return source file contents with syntax highlighting
// @Tags Source
// @Produce json
// @Param path query string true "File path (relative to project root)"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{} "Path not provided or invalid"
// @Failure 403 {object} map[string]interface{} "Path not allowed (not in whitelist)"
// @Failure 404 {object} map[string]interface{} "File not found"
// @Router /source [get]
func ReadSourceFile(c *fiber.Ctx) error {
	filePath := c.Query("path")
	if filePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "path parameter is required",
		})
	}

	response, err := sourceService.ReadFile(filePath)
	if err != nil {
		// Determine status code based on error message
		statusCode := fiber.StatusInternalServerError

		switch err.Error() {
		case "no path provided":
			statusCode = fiber.StatusBadRequest
		case "invalid path", "path is outside base path":
			statusCode = fiber.StatusBadRequest
		case "path not allowed":
			statusCode = fiber.StatusForbidden
		case "file not found":
			statusCode = fiber.StatusNotFound
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
			"path":  filePath,
		})
	}

	// Determine file extension for syntax highlighting
	fileExt := getFileExtension(filePath)
	language := getLanguageFromExtension(fileExt)

	return c.JSON(fiber.Map{
		"success":  true,
		"code":     response.Code,
		"path":     response.Path,
		"size":     response.Size,
		"language": language,
	})
}

// Helper functions

func getFileExtension(filePath string) string {
	// Find the last dot
	lastDot := -1
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			lastDot = i
			break
		}
		// Stop at path separator
		if filePath[i] == '/' || filePath[i] == '\\' {
			break
		}
	}

	if lastDot == -1 {
		return ""
	}

	return filePath[lastDot+1:]
}

func getLanguageFromExtension(ext string) string {
	extensionMap := map[string]string{
		// TypeScript/JavaScript
		"ts":  "typescript",
		"tsx": "typescript",
		"js":  "javascript",
		"jsx": "javascript",
		"mjs": "javascript",
		"cjs": "javascript",

		// Go
		"go": "go",

		// Python
		"py":  "python",
		"pyx": "python",
		"pyi": "python",
		"pyc": "python",

		// JSON/YAML
		"json": "json",
		"yaml": "yaml",
		"yml":  "yaml",
		"toml": "toml",

		// Markdown
		"md":       "markdown",
		"mdx":      "markdown",
		"markdown": "markdown",

		// Shell
		"sh":   "shell",
		"bash": "shell",
		"zsh":  "shell",

		// SQL
		"sql": "sql",

		// CSS/SCSS
		"css":  "css",
		"scss": "scss",
		"sass": "sass",

		// HTML
		"html": "html",
		"htm":  "html",

		// XML
		"xml": "xml",

		// C/C++
		"c":   "c",
		"cc":  "cpp",
		"cpp": "cpp",
		"cxx": "cpp",
		"h":   "cpp",
		"hpp": "cpp",

		// Java
		"java": "java",

		// Rust
		"rs": "rust",

		// PHP
		"php": "php",

		// Ruby
		"rb": "ruby",

		// Lua
		"lua": "lua",

		// R
		"r": "r",

		// Dockerfile
		"dockerfile": "dockerfile",

		// Plain text
		"txt": "plaintext",
	}

	if lang, exists := extensionMap[ext]; exists {
		return lang
	}

	return "plaintext"
}
