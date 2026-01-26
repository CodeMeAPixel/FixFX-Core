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

## [0.2.0] - 2026-01-25

### Added
- **Full version string for hosting panels** - Added `FullVersion` field to artifact entries
  - Format: `{version}-{hash}` (e.g., `24769-315823736cfbc085104ca0d32779311cd2f1a5a8`)
  - Compatible with Pterodactyl, Pelican, and similar hosting panel egg configurations

- **Artifact statistics in API response** - Added `stats` object to metadata
  - Includes counts for: `total`, `recommended`, `latest`, `active`, `deprecated`, `eol`
  - Calculated from filtered results before pagination
  - Enables frontend to show accurate totals regardless of current page

### Fixed
- **Pagination total count** - Fixed incorrect total count in pagination metadata
  - Previously returned count of paginated results instead of total filtered results
  - Created `ArtifactsResult` struct to properly track total count after filtering but before pagination
  - `hasMore` now correctly indicates if more pages are available

- **Latest vs Recommended logic** - Fixed support status assignment per CFX EOL policy
  - **Latest** = Single newest version (for testing/bleeding edge)
  - **Recommended** = Next 3 versions after Latest (stable for production)
  - Support status now dynamically assigned based on version position, not hardcoded thresholds
  - See https://aka.cfx.re/eol for CFX official policy

- **EOL filter default** - Changed `includeEol` default from `true` to `false`
  - EOL artifacts are now excluded by default for safety
  - Users must explicitly opt-in to see end-of-life versions

### Changed
- Updated all artifact handlers to use new `ArtifactsResult` return type
- Refactored `generateFullVersion()` helper to accept hash parameter
- Added `ArtifactStats` struct and `calculateStats()` helper function
- Refactored `ProcessGitHubTags` to dynamically assign Latest/Recommended based on sorted position
- Simplified `determineSupportStatus()` to only handle Active/Deprecated/EOL thresholds

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
