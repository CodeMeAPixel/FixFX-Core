# FixFX Backend

A high-performance Go backend API for the FixFX documentation platform.

## Quick Start

```bash
# Install dependencies
go mod download

# Generate Swagger documentation
swag init

# Run the server
go run main.go
```

The API will be available at `http://localhost:3001`

## Documentation

- Interactive API Docs: http://localhost:3001/docs
- Swagger JSON: http://localhost:3001/docs/json

## Architecture

The backend is organized into:

- **Routes** (`internal/routes/`) - API endpoint definitions with Swagger documentation
- **Handlers** (`internal/handlers/`) - Request handlers
- **Services** (`internal/services/`) - Business logic (fetching, processing, filtering)
- **Models** (`internal/models/`) - Data structures

## APIs

### Artifacts API
Fetch FiveM/RedM server artifacts with filtering and pagination.

```
GET /api/artifacts/fetch
  - platform (windows/linux/all)
  - status (recommended/latest/active/deprecated/eol)
  - sortBy (version/date)
  - limit, offset
```

### Natives API
Query game native functions database.

```
GET /api/natives
  - game (gta5/rdr3/all)
  - environment (server/client/all)
  - search, namespace
  - limit, offset
```

### Source API
Securely serve source code files.

```
GET /api/source?path=path/to/file
```

### Search API
Full-text search across documentation.

```
GET /api/search?q=query&type=all
```

## Migration Status

This backend is being progressively populated with functionality from the Next.js frontend API routes.

### Completed
- ✅ Project structure
- ✅ Swagger/OpenAPI documentation setup
- ✅ Route definitions
- ✅ CORS and logging middleware

### In Progress
- 🔄 Artifacts service implementation
- 🔄 Natives service implementation
- 🔄 Source file serving
- 🔄 Search functionality

### Planned
- ⏳ Caching (Redis)
- ⏳ Rate limiting
- ⏳ Authentication
- ⏳ Database integration
- ⏳ Performance optimization

## Development

For detailed development guide, see [README.md](README.md)

## License

Apache License 2.0
