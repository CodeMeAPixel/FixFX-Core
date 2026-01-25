# FixFX Backend API

A high-performance Go backend for the FixFX documentation and tools platform, built with Fiber and featuring comprehensive Swagger documentation.

## Features

- 🚀 **Fast & Lightweight** - Built with Fiber web framework for high performance
- 📚 **Artifacts API** - Fetch and filter FiveM/RedM server artifacts
- 🔎 **Natives API** - Query game natives database
- 📖 **Source API** - Securely serve source code files
- 🔍 **Search API** - Full-text search across documentation
- 📖 **Swagger Docs** - Auto-generated API documentation with ReDoc
- 🛡️ **Security** - Path whitelisting, input validation, CORS support

## Prerequisites

- Go 1.22 or higher
- A recent version of the Fiber framework
- Swaggo for Swagger documentation

## Installation

1. **Clone the repository**
```bash
git clone https://github.com/CodeMeAPixel/FixFX.git
cd FixFX/backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Create environment file**
```bash
cp .env.example .env
```

4. **Generate Swagger docs**
```bash
swag init
```

5. **Run the server**
```bash
go run main.go
```

The API will be available at `http://localhost:3001`

## API Endpoints

### Health Check
- `GET /health` - Check API status

### Artifacts
- `GET /api/artifacts/fetch` - Get FiveM/RedM artifacts
- `GET /api/artifacts/version/{version}` - Get specific artifact version
- `GET /api/artifacts/check` - Check artifact availability
- `GET /api/artifacts/changes` - Get artifact changelog

### Natives
- `GET /api/natives` - Get game natives
- `GET /api/natives/search` - Search natives
- `GET /api/natives/{hash}` - Get native by hash
- `GET /api/natives/stats` - Get natives statistics

### Source
- `GET /api/source` - Read source file (requires `path` parameter)

### Search
- `GET /api/search` - Global search across documentation

## Documentation

### Swagger/OpenAPI
- Access interactive API docs at `http://localhost:3001/docs`
- View JSON spec at `http://localhost:3001/docs/json`
- ReDoc documentation available via the Swagger UI

## Development

### Project Structure
```
backend/
├── main.go              # Application entry point
├── go.mod             # Go modules file
├── internal/
│   ├── routes/        # API route definitions
│   │   ├── artifacts.go
│   │   ├── natives.go
│   │   ├── source.go
│   │   └── types.go   # Type definitions for Swagger
│   ├── models/        # Data models
│   ├── services/      # Business logic
│   └── handlers/      # Request handlers
├── docs/              # Generated Swagger documentation
└── README.md
```

### Adding New Routes

1. Create a route file in `internal/routes/`
2. Define Swagger comments for documentation
3. Implement handlers
4. Register routes in `main.go`
5. Run `swag init` to regenerate Swagger docs

Example:
```go
// @Summary Get resource
// @Description Get a specific resource
// @Tags Resources
// @Produce json
// @Success 200 {object} Resource
// @Router /api/resource/{id} [get]
func GetResource(c fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "success"})
}
```

## Building for Production

### Build binary
```bash
go build -o fixfx-backend main.go
```

### Docker build
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o fixfx-backend main.go

FROM alpine:latest
WORKDIR /root
COPY --from=builder /app/fixfx-backend .
EXPOSE 3001
CMD ["./fixfx-backend"]
```

## Environment Variables

```bash
PORT=3001                  # Server port
GITHUB_TOKEN=              # GitHub API token (optional, for higher rate limits)
LOG_LEVEL=info            # Logging level
ENVIRONMENT=development   # Environment (development/production)
```

## Performance

- **Artifacts API**: ~50-100ms for full dataset fetch
- **Natives API**: ~30-80ms for queries
- **Source API**: ~10-20ms for file reads
- Memory footprint: ~20-50MB

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## Support

For issues and questions:
- 📧 Email: support@fixfx.dev
- 🐛 Bug Reports: [GitHub Issues](https://github.com/CodeMeAPixel/FixFX/issues)
- 💬 Discord: [Join our community](https://discord.gg/fixfx)

## Roadmap

- [ ] Implement Artifacts service
- [ ] Implement Natives service
- [ ] Implement Source file serving
- [ ] Add caching layer (Redis)
- [ ] Add rate limiting
- [ ] Add authentication/authorization
- [ ] Performance optimization
- [ ] Database integration
- [ ] WebSocket support for real-time updates

## Acknowledgments

- [Fiber](https://gofiber.io) - Express-inspired web framework
- [Swaggo](https://github.com/swaggo/swag) - Swagger documentation generator
- [FiveM/RedM](https://fivem.net) - Game modding frameworks
