package server

import (
	"net/http"

	"go-twitch/config"
	"go-twitch/handlers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// New creates and configures the Fiber app with routes and middleware.
func New(cfg config.Config) *fiber.App {
	app := fiber.New()

	// Static files
	app.Use("/", filesystem.New(filesystem.Config{
		Root:  http.Dir("./static"),
		Index: "mainpage.html",
	}))

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.SendFile("./static/dashboard.html")
	})

	// REST endpoints
	app.Get("/user/:name", handlers.GetUser)
	app.Get("/stream/:name", handlers.GetStream)
	app.Get("/games/top", handlers.GetTopGames)

	// OAuth endpoints
	app.Get("/authorize", handlers.AuthorizePage)
	app.Get("/auth/start", handlers.AuthStart)
	app.Get("/auth/callback", handlers.AuthCallback)
	app.Get("/auth/status", handlers.AuthStatus)
	app.Get("/callback", handlers.CallbackAlias)

	// IRC endpoints
	app.Post("/irc/subscribe", handlers.IRCSubscribe)
	app.Post("/irc/subscribe/:channel", handlers.IRCSubscribeParam)
	app.Get("/irc/subscribe/:channel", handlers.IRCSubscribeParamHTML)
	app.Post("/irc/unsubscribe", handlers.IRCUnsubscribe)
	app.Get("/irc/unsubscribe/:channel", handlers.IRCUnsubscribeParamHTML)
	app.Post("/irc/send", handlers.IRCSend)

	// WebSocket and SSE
	app.Get("/ws", websocket.New(WebsocketHandler))
	app.Get("/irc/:channel/stream", SSEChannelStream)

	return app
}
