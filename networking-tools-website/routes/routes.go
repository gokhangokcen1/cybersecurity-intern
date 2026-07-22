package routes

import (
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gokhangokcen1/subnet-backend/handlers"
	"github.com/valyala/fasthttp"
)

var captureUpgrader = websocket.FastHTTPUpgrader{
	CheckOrigin: func(_ *fasthttp.RequestCtx) bool {
		return true
	},
}

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/subnet", handlers.HesaplaSubnet)

	api.Post("/portcheck", handlers.KontrolEtHandler)
	api.Get("/myip", handlers.MevcutIP)

	api.Post("/scan", handlers.TaraHandler)
	api.Post("/sslcheck", handlers.SslKontrol)

	app.Get("/api/devices", handlers.ListDevicesHandler)
	app.Post("/api/capture/start", handlers.StartCaptureHandler)
	app.Post("/api/capture/stop", handlers.StopCaptureHandler)
	app.Get("/ws/capture", func(c fiber.Ctx) error {
		return captureUpgrader.Upgrade(c.RequestCtx(), handlers.CaptureWebSocketHandler)
	})

	api.Post("/packet-sender", handlers.PortCraftHandler)

	app.Get("/api/devices", handlers.ListDevicesHandler)
	app.Post("/api/capture/start", handlers.StartCaptureHandler)
	app.Post("/api/capture/stop", handlers.StopCaptureHandler)

	api.Get("/dnschecker", handlers.CheckDNSHandler) // düzeltildi: Post → Get, path'ten "/api" kaldırıldı
	api.Get("/dnschecker/full", handlers.GetFullRecordsHandler)
}
