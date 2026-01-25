# FixFX Backend API

A high performance Go backend for the FixFX documentation and tools platform, built with Fiber and featuring comprehensive API documentation with ReDoc.

## Features

- **Fast & Lightweight** - Built with Fiber v2 web framework for high performance
- **Artifacts API** - Fetch and filter FiveM/RedM server artifacts from GitHub
- **Natives API** - Query game natives database with multi-source support (GTA5, RDR3, CFX)
- **Source API** - Securely serve source code files with syntax highlighting
- **Search API** - Full-text search across documentation (Next.js backend, will be migrated)
- **API Documentation** - Auto-generated with Swagger and beautiful ReDoc UI
- **Version Tracking** - Automatic version logging on startup with update checks
- **Security** - Path whitelisting, input validation, CORS support, proper error handling

## Version

Current version: **0.1.0** (January 25, 2026)

See [CHANGELOG.md](./CHANGELOG.md) for full release notes and planned features.

## Prerequisites

- Go 1.22 or higher
- Make (for convenient commands)
- Optional: GitHub API token for higher rate limits

## Installation

1. **Clone the repository**
```bash
git clone https://github.com/CodeMeAPixel/FixFX-Core.git
cd FixFX-Core
```

2. **Install dependencies**
```bash
make install
# or manually: go mod download
```

3. **Create environment file**
```bash
cp .env.example .env
```

4. **Generate/Update Swagger docs**
```bash
make swagger
# or manually: swag init
```

5. **Run the server**
```bash
make dev          # Development with hot-reload
# or
make run          # Production binary
# or
go run main.go    # Direct run
```

The API will be available at `http://localhost:3001`

## API Endpoints

### Health Check
- `GET /health` - Check API status and version

### Artifacts
- `GET /api/artifacts/fetch` - Get FiveM/RedM artifacts with filtering and pagination
- `GET /api/artifacts/version/:version` - Get specific artifact version
- `GET /api/artifacts/check` - Check artifact availability
- `GET /api/artifacts/changes` - Get artifact changelog between versions

### Natives
- `GET /api/natives` - Get game natives with filtering and pagination
- `GET /api/natives/search` - Full-text search across natives database
- `GET /api/natives/:hash` - Get specific native by hash
- `GET /api/natives/stats` - Get natives statistics

### Source
- `GET /api/source` - Read source file (requires `path` parameter with whitelisted paths)

### Search
- `GET /api/search` - Global search across documentation (currently on Next.js)

## Documentation

### Interactive API Documentation
- **ReDoc UI** - `http://localhost:3001/docs` - Beautiful, interactive API explorer
- **Swagger JSON** - `http://localhost:3001/docs/doc.json` - Machine-readable specification
- **Swagger UI** (fallback) - Available through ReDoc

### Startup Output
On startup, the server displays:
```
╔═══════════════════════════════════════════════════════════╗
║          FixFX Backend API - Starting Up                  ║
╠═══════════════════════════════════════════════════════════╣
║ Version:     v0.1.0                                       ║
║ Build Time:  2026-01-25                                   ║
║ Environment: development/production                       ║
╚═══════════════════════════════════════════════════════════╝

FixFX API v0.1.0 is running
Documentation available at http://localhost:3001/docs
Health check at http://localhost:3001/health
```

## Development

### Project Structure
```
backend/
├── main.go                    # Application entry point with version tracking
├── go.mod                     # Go modules
├── Makefile                   # Build automation
├── CHANGELOG.md               # Release notes
├── internal/
│   ├── routes/                # API route definitions with Swagger comments
│   │   ├── artifacts.go
│   │   ├── natives.go
│   │   ├── source.go
│   │   ├── search.go
│   │   └── types.go           # Type definitions for Swagger
│   ├── services/              # Business logic layer
│   │   ├── artifacts.go       # GitHub API integration, caching, filtering
│   │   ├── natives.go         # Multi-source native fetching, search
│   │   └── source.go          # Secure file serving with whitelist
│   └── handlers/              # HTTP request/response handling
│       ├── artifacts.go
│       ├── natives.go
│       └── source.go
├── docs/                      # Auto-generated Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
└── README.md
```

### Making API Requests

Example: Fetch artifacts
```bash
curl "http://localhost:3001/api/artifacts/fetch?platform=windows&limit=10&offset=0"
```

Example: Search natives
```bash
curl "http://localhost:3001/api/natives/search?q=player&limit=20"
```

Example: Get health status
```bash
curl http://localhost:3001/health
```

### Adding New Routes

1. Create a route file in `internal/routes/`
2. Define Swagger comments for documentation
3. Implement handlers in `internal/handlers/`
4. Create service logic in `internal/services/` if needed
5. Register routes in `main.go`
6. Run `make swagger` to regenerate Swagger docs

Example handler with Swagger docs:
```go
// @Summary Get resource
// @Description Get a specific resource by ID
// @Tags Resources
// @Param id path string true "Resource ID"
// @Produce json
// @Success 200 {object} Resource
// @Failure 404 {object} ErrorResponse
// @Router /api/resource/{id} [get]
func GetResource(c *fiber.Ctx) error {
    id := c.Params("id")
    return c.JSON(fiber.Map{"id": id})
}
```

## Building for Production

### Build binary
```bash
make build
# or manually: go build -o fixfx-backend main.go
```

### Docker
```bash
docker build -t fixfx-backend .
docker run -p 3001:3001 -e GITHUB_TOKEN=your_token fixfx-backend
```

See Dockerfile for multi-stage build configuration.

## Environment Variables

```bash
PORT=3001                     # Server port (default: 3001)
GITHUB_TOKEN=                 # GitHub API token for artifacts (optional)
LOG_LEVEL=info               # Logging level
ENVIRONMENT=development      # Environment: development or production
```

## Performance Characteristics

- **Artifacts API**: ~50-100ms (first call with API fetch, cached afterwards)
- **Natives API**: ~30-80ms for queries
- **Source API**: ~10-20ms for file reads
- **Memory usage**: ~20-50MB
- **Caching**: 1-hour TTL for all cached data

## Testing

```bash
make test          # Run test suite
make build         # Verify code compiles
make swagger       # Verify Swagger generation
```

## Makefile Commands

```bash
make install       # Install dependencies and tools
make dev          # Run with hot-reload (air)
make build        # Build production binary
make run          # Run compiled binary
make test         # Run tests
make swagger      # Generate Swagger docs
make clean        # Remove binary and generated files
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Future Roadmap

See [CHANGELOG.md](./CHANGELOG.md) for planned features including:
- Chat API endpoint
- Contributors API with GitHub integration
- Global search service (backend)
- Redis caching layer
- Rate limiting
- Authentication/Authorization
- Comprehensive test suite

## License

This project is licensed under the AGPL 3.0 License - see the LICENSE file for details.