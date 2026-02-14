package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	_ "github.com/CodeMeAPixel/FixFX-Core/docs"
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/CodeMeAPixel/FixFX-Core/internal/routes"
)

// Version information
const (
	Version   = "0.1.0"
	BuildTime = "2026-01-25"
)

// @title FixFX API
// @version 0.1.0
// @description Backend API for FixFX documentation and tools

// @contact.name Support
// @contact.url https://fixfx.wiki/discord

// @host localhost:3001
// @BasePath /api
// @schemes http https

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Log startup information
	log.Printf("╔═══════════════════════════════════════════════════════════╗")
	log.Printf("║          FixFX Backend API - Starting Up                  ║")
	log.Printf("╠═══════════════════════════════════════════════════════════╣")
	log.Printf("║ Version:     v%s                                         ║", Version)
	log.Printf("║ Build Time:  %s                                    ║", BuildTime)
	log.Printf("║ Environment: %s                                         ║", os.Getenv("ENVIRONMENT"))
	log.Printf("╚═══════════════════════════════════════════════════════════╝")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "FixFX API",
		ServerHeader: "FixFX/" + Version,
		ErrorHandler: globalErrorHandler,
	})

	// Middleware
	app.Use(corsMiddleware())
	app.Use(loggingMiddleware())

	// Initialize handlers
	githubToken := os.Getenv("GITHUB_TOKEN")
	handlers.InitArtifactsHandler(githubToken)
	handlers.InitNativesHandler()
	handlers.InitSourceHandler(".", nil) // Use current directory as base path
	handlers.InitContributorsHandler(githubToken)
	handlers.InitValidatorHandler()

	// Health check
	app.Get("/health", healthCheck)

	// Swagger JSON endpoint for ReDoc (must come first)
	app.Get("/docs/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	// ReDoc UI - Custom styled documentation
	app.Get("/docs", docsHandler)
	app.Get("/docs/", docsHandler)
	app.Get("/docs/index.html", docsHandler)

	// API Routes
	api := app.Group("/api")

	// Register route groups
	routes.RegisterArtifactsRoutes(api)
	routes.RegisterNativesRoutes(api)
	routes.RegisterSourceRoutes(api)
	routes.RegisterSearchRoutes(api)
	routes.RegisterContributorsRoutes(api)
	routes.RegisterValidatorRoutes(api)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Starting FixFX API on port %s", port)

	// Check for version updates in background
	go checkVersionUpdates()

	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

// @Summary Health Check
// @Description Check if the API is running and get version information
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"version":   Version,
		"service":   "FixFX API",
		"timestamp": c.Get("Date"),
	})
}

func corsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(200)
		}

		return c.Next()
	}
}

func loggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Printf("%s %s", c.Method(), c.Path())
		return c.Next()
	}
}

func globalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":  err.Error(),
		"status": code,
	})
}

var docsHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8"/>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>FixFX API Documentation</title>
	<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		
		html {
			scroll-behavior: smooth;
		}
		
		body {
			font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
			background: linear-gradient(135deg, #0f0f0f 0%, #1a1a1a 100%);
			color: #e4e4e7;
			line-height: 1.6;
		}
		
		redoc {
			display: block;
		}
		
		/* ReDoc customization */
		:root {
			--primary-color: #5865F2;
			--secondary-color: #9D84B7;
			--background-color: #09090b;
			--text-color: #e4e4e7;
			--border-color: #27272a;
			--code-background: #18181b;
		}
		
		/* ReDoc Components - Dark Theme */
		.redoc-container {
			background: #09090b !important;
			color: #e4e4e7 !important;
		}
		
		[data-testid="redoc-container"] {
			background: #09090b !important;
		}
		
		/* Sidebar styling */
		.menu-content, [class*="sidebar"], nav {
			background: #0f0f0f !important;
		}
		
		.menu-item, [class*="menu-item"] {
			color: #e4e4e7 !important;
		}
		
		.menu-item.active, [class*="menu-item"][class*="active"] {
			color: #5865F2 !important;
			background: rgba(88, 101, 242, 0.1) !important;
		}
		
		.menu-item:hover {
			background: rgba(88, 101, 242, 0.15) !important;
			color: #7c7cff !important;
		}
		
		/* Header and panels */
		.api-info, [class*="api-info"] {
			background: rgba(88, 101, 242, 0.05) !important;
			border-bottom: 1px solid #27272a !important;
		}
		
		[class*="header"], .header {
			background: linear-gradient(135deg, rgba(88, 101, 242, 0.05) 0%, rgba(157, 132, 183, 0.03) 100%) !important;
			border-bottom: 1px solid #27272a !important;
		}
		
		/* Content area */
		[class*="right-panel"], .right-panel {
			background: #09090b !important;
		}
		
		/* Syntax highlighting */
		code {
			font-family: 'JetBrains Mono', monospace;
			background: #18181b !important;
			color: #e4e4e7 !important;
			padding: 2px 6px;
			border-radius: 4px;
		}
		
		pre {
			background: #18181b !important;
			border: 1px solid #27272a !important;
			color: #e4e4e7 !important;
		}
		
		pre code {
			background: transparent !important;
			color: inherit !important;
		}
		
		/* Links */
		a, [class*="link"] {
			color: #5865F2 !important;
			text-decoration: none;
		}
		
		a:hover {
			color: #7c7cff !important;
			text-decoration: underline;
		}
		
		/* Scrollbar styling */
		::-webkit-scrollbar {
			width: 10px;
		}
		
		::-webkit-scrollbar-track {
			background: #09090b;
		}
		
		::-webkit-scrollbar-thumb {
			background: #27272a;
			border-radius: 5px;
		}
		
		::-webkit-scrollbar-thumb:hover {
			background: #3f3f46;
		}
		
		/* Tables */
		table {
			border-collapse: collapse;
		}
		
		th, [class*="th"] {
			background: rgba(88, 101, 242, 0.1) !important;
			color: #5865F2 !important;
			border: 1px solid #27272a !important;
			padding: 12px !important;
			text-align: left;
			font-weight: 600;
		}
		
		td, [class*="td"] {
			border: 1px solid #27272a !important;
			padding: 12px !important;
			background: #09090b !important;
			color: #e4e4e7 !important;
		}
		
		tr:hover {
			background: rgba(88, 101, 242, 0.05) !important;
		}
		
		/* Buttons */
		button, [role="button"] {
			background: #5865F2 !important;
			color: white !important;
			border: none !important;
			padding: 8px 16px !important;
			border-radius: 6px !important;
			cursor: pointer;
			font-weight: 500;
		}
		
		button:hover, [role="button"]:hover {
			background: #7c7cff !important;
		}
		
		/* Response codes */
		.response-code, [class*="response"] {
			background: #18181b !important;
			color: #e4e4e7 !important;
			border: 1px solid #27272a !important;
		}
		
		/* Try it out button */
		[class*="try-it"] {
			background: #5865F2 !important;
			color: white !important;
		}
		
		/* Text and general styling */
		p, span, div {
			color: #e4e4e7 !important;
		}
		
		/* Headings */
		h1, h2, h3, h4, h5, h6 {
			color: #e4e4e7 !important;
		}
		
		/* Loading animation */
		@keyframes pulse {
			0%, 100% {
				opacity: 1;
			}
			50% {
				opacity: 0.5;
			}
		}
		
		.loading {
			animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
		}
	</style>
</head>
<body>
	<redoc 
		spec-url="/docs/doc.json"
		expand-single-schema="true"
		sort-props-alphabetically="true"
		show-extensions="true"
		native-scrollbars="true"
		path-in-middle-panel="true"
		theme='{
			"colors": {
				"primary": {
					"main": "#5865F2",
					"light": "#7c7cff",
					"dark": "#4752c4"
				},
				"success": {
					"main": "#10b981"
				},
				"error": {
					"main": "#ef4444"
				},
				"warning": {
					"main": "#f59e0b"
				},
				"info": {
					"main": "#3b82f6"
				},
				"http": {
					"get": "#10b981",
					"post": "#3b82f6",
					"put": "#f59e0b",
					"delete": "#ef4444",
					"patch": "#8b5cf6",
					"head": "#6b7280",
					"options": "#9333ea"
				},
				"text": {
					"primary": "#e4e4e7",
					"secondary": "#a1a1a6"
				},
				"border": {
					"dark": "#27272a",
					"light": "#3f3f46"
				},
				"background": {
					"primary": "#09090b",
					"secondary": "#18181b"
				}
			},
			"schema": {
				"nestedBackground": "#18181b",
				"linesColor": "#27272a"
			},
			"typography": {
				"fontFamily": "\"Inter\", -apple-system, BlinkMacSystemFont, \"Segoe UI\", sans-serif",
				"fontSize": {
					"base": "14px",
					"big": "18px",
					"code": "13px",
					"small": "12px",
					"title": "24px"
				},
				"lineHeight": "1.6",
				"fontWeightRegular": "400",
				"fontWeightBold": "600",
				"fontWeightLight": "300",
				"smooth": "true"
			},
			"sidebar": {
				"width": "250px",
				"backgroundColor": "#0f0f0f",
				"textColor": "#e4e4e7"
			},
			"rightPanel": {
				"backgroundColor": "#09090b"
			},
			"codeBlock": {
				"backgroundColor": "#18181b",
				"textColor": "#e4e4e7"
			}
		}'>
	</redoc>
	
	<script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
</body>
</html>
`

func docsHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(docsHTML)
}

// checkVersionUpdates checks GitHub releases for newer versions
// This runs as a goroutine to not block startup
func checkVersionUpdates() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Version check failed: %v", r)
		}
	}()

	// Wait for Fiber to print its startup message first
	time.Sleep(100 * time.Millisecond)

	// For now, just log the current version
	// In a future release, we can add GitHub API checking
	log.Printf("")
	log.Printf("✓ FixFX API v%s is running", Version)
	log.Printf("→ Documentation available at http://localhost:3001/docs")
	log.Printf("→ Health check at http://localhost:3001/health")
	log.Printf("")

	// TODO: Add GitHub API check for new versions
	// Example implementation:
	// resp, err := http.Get("https://api.github.com/repos/CodeMeAPixel/FixFX-Core/releases/latest")
	// if err != nil {
	// 	log.Printf("Could not check for updates: %v", err)
	// 	return
	// }
	// var latestRelease struct {
	// 	TagName string `json:"tag_name"`
	// }
	// json.NewDecoder(resp.Body).Decode(&latestRelease)
	// if latestRelease.TagName > "v"+Version {
	// 	log.Printf("⚠️  New version available: %s (current: v%s)", latestRelease.TagName, Version)
	// }
}
