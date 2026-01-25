# FixFX Backend Changelog

All notable changes to the FixFX backend will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Release Strategy
- **Patch versions** (0.1.x): Bug fixes, minor improvements
- **Minor versions** (0.x.0): New features, non-breaking changes
- **Major versions** (x.0.0): Breaking API changes

## [Unreleased]

### Planned
- [ ] Chat API endpoint with message persistence
- [ ] Contributors API with GitHub integration
- [ ] Global search service (cross-service search)
- [ ] Redis caching layer for improved performance
- [ ] Rate limiting middleware
- [ ] Authentication/Authorization system
- [ ] Comprehensive test suite
- [ ] Docker containerization
- [ ] Monitoring and metrics collection

---

## [0.1.0] - 2026-01-25

### Added

#### Core Infrastructure
- **Fiber v2 web framework** with middleware support
  - CORS middleware for cross-origin requests
  - Custom logging middleware
  - Global error handler with proper HTTP status codes
  - Health check endpoint (`/health`)

- **Swagger/OpenAPI documentation** with ReDoc UI
  - Auto-generated API specification from code comments
  - Beautiful dark-themed ReDoc interface matching frontend design
  - Interactive API documentation at `/docs`
  - JSON spec available at `/docs/doc.json`

#### Services & Handlers

**Artifacts Service** (3 handlers, 4 endpoints)
- GitHub API integration for FiveM server artifacts
- Artifact fetching with intelligent caching (1-hour TTL)
- Version extraction and support status determination
  - Status types: recommended, latest, active, deprecated, eol
- Platform-specific artifact generation (Windows/Linux)
- Advanced filtering:
  - Filter by platform (windows/linux/all)
  - Filter by version
  - Filter by support status
- Sorting capabilities:
  - Sort by version (semantic versioning aware)
  - Sort by date
  - Ascending/descending order
- Pagination support:
  - Configurable limit (default 50, max 200)
  - Offset-based pagination
- Changelog retrieval between versions
- Fallback data handling for API failures

Endpoints:
- `GET /api/artifacts/fetch` - Fetch artifacts with filtering & pagination
- `GET /api/artifacts/version/:version` - Get specific version details
- `GET /api/artifacts/check` - Check artifact availability
- `GET /api/artifacts/changes` - Get changelog between versions

**Natives Service** (4 handlers, 4 endpoints)
- Multi-source native function fetching:
  - GTA5 natives from https://runtime.fivem.net/doc/natives.json
  - RDR3 natives from runtime.fivem.net
  - CFX natives (custom CitizenFX natives)
- Intelligent caching (1-hour TTL per source)
- Environment detection (client/server/shared)
- Namespace-based organization
- Full-text search with relevance scoring
  - Exact name match: +4 points
  - Name contains: +3 points
  - Description match: +2 points
  - Parameter match: +1 point
- Advanced filtering:
  - Filter by game (gta5/rdr3/cfx)
  - Filter by environment (client/server/shared)
  - Filter by namespace
- Pagination support
- Statistics generation (total natives, by game, by environment)

Endpoints:
- `GET /api/natives` - Get natives with filtering & pagination
- `GET /api/natives/search` - Full-text search across all natives
- `GET /api/natives/:hash` - Get specific native by hash
- `GET /api/natives/stats` - Get natives statistics

**Source Service** (1 handler, 1 endpoint)
- Secure file serving with whitelist validation
- Path traversal prevention
  - Prevents `..` directory traversal
  - Absolute path validation
  - Whitelist-only access
- Syntax highlighting detection for 25+ file types
  - TypeScript, JavaScript, Go, Python
  - JSON, YAML, XML
  - SQL, Shell, Dockerfile
  - And more...
- Proper error handling (400/403/404)
- File metadata (size, language, path)
