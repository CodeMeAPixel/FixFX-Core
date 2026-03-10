package routes

import (
	"github.com/CodeMeAPixel/FixFX-Core/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

// RegisterGameReferencesRoutes registers all game reference routes under /api/game-references
func RegisterGameReferencesRoutes(api fiber.Router) {
	gr := api.Group("/game-references")

	gr.Get("/blips", handlers.GetBlips)
	gr.Get("/checkpoints", handlers.GetCheckpoints)
	gr.Get("/data-files", handlers.GetDataFiles)
	gr.Get("/game-events", handlers.GetGameEvents)
	gr.Get("/gamer-tags", handlers.GetGamerTags)
	gr.Get("/hud-colors", handlers.GetHUDColors)
	gr.Get("/markers", handlers.GetMarkers)
	gr.Get("/net-game-events", handlers.GetNetGameEvents)
	gr.Get("/ped-models", handlers.GetPedModels)
	gr.Get("/pickup-hashes", handlers.GetPickupHashes)
	gr.Get("/weapon-models", handlers.GetWeaponModels)
	gr.Get("/zones", handlers.GetZones)
	gr.Get("/vehicle-models", handlers.GetVehicleModels)
	gr.Get("/vehicle-colours", handlers.GetVehicleColours)
	gr.Get("/vehicle-flags", handlers.GetVehicleFlags)
}
