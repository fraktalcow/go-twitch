package handlers

import (
	"encoding/json"
	"os"

	"go-twitch/twitch"

	"github.com/gofiber/fiber/v2"
)

// Subscribe endpoints
func IRCSubscribe(c *fiber.Ctx) error {
	var req struct {
		Channel string `json:"channel"`
	}
	if err := c.BodyParser(&req); err != nil || req.Channel == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Missing channel name"})
	}
	twitch.IRCClientsMu.Lock()
	_, alreadyJoined := twitch.IRCClients[req.Channel]
	twitch.IRCClientsMu.Unlock()
	if alreadyJoined {
		return c.JSON(fiber.Map{"success": false, "message": "Already subscribed to IRC chat for channel " + req.Channel})
	}
	go twitch.StartIRCRelay(req.Channel)
	return c.JSON(fiber.Map{"success": true, "message": "Subscribed to IRC chat for channel " + req.Channel})
}

func IRCSubscribeParam(c *fiber.Ctx) error {
	channel := c.Params("channel")
	if channel == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Missing channel name"})
	}
	twitch.IRCClientsMu.Lock()
	_, alreadyJoined := twitch.IRCClients[channel]
	twitch.IRCClientsMu.Unlock()
	if alreadyJoined {
		return c.JSON(fiber.Map{"success": false, "message": "Already subscribed to IRC chat for channel " + channel})
	}
	go twitch.StartIRCRelay(channel)
	return c.JSON(fiber.Map{"success": true, "message": "Subscribed to IRC chat for channel " + channel})
}

func IRCSubscribeParamHTML(c *fiber.Ctx) error {
	channel := c.Params("channel")
	if channel == "" {
		return c.Status(400).SendString("Missing channel name")
	}
	twitch.IRCClientsMu.Lock()
	_, alreadyJoined := twitch.IRCClients[channel]
	twitch.IRCClientsMu.Unlock()
	var result map[string]interface{}
	if alreadyJoined {
		result = fiber.Map{"success": false, "message": "Already subscribed to IRC chat for channel " + channel}
	} else {
		go twitch.StartIRCRelay(channel)
		result = fiber.Map{"success": true, "message": "Subscribed to IRC chat for channel " + channel}
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return c.SendString("<pre style='background:#161b22;color:#c9d1d9;padding:16px;border-radius:8px;font-size:1.1em;'>" + string(jsonBytes) + "</pre>")
}

// Unsubscribe endpoints
func IRCUnsubscribe(c *fiber.Ctx) error {
	var req struct {
		Channel string `json:"channel"`
	}
	if err := c.BodyParser(&req); err != nil || req.Channel == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Missing channel name"})
	}
	twitch.IRCClientsMu.Lock()
	_, exists := twitch.IRCClients[req.Channel]
	twitch.IRCClientsMu.Unlock()
	if !exists {
		return c.JSON(fiber.Map{"success": false, "message": "Not subscribed to IRC chat for channel " + req.Channel})
	}
	twitch.StopIRCRelay(req.Channel)
	return c.JSON(fiber.Map{"success": true, "message": "Unsubscribed from IRC chat for channel " + req.Channel})
}

func IRCUnsubscribeParamHTML(c *fiber.Ctx) error {
	channel := c.Params("channel")
	if channel == "" {
		return c.Status(400).SendString("Missing channel name")
	}
	twitch.IRCClientsMu.Lock()
	_, exists := twitch.IRCClients[channel]
	twitch.IRCClientsMu.Unlock()
	var result map[string]interface{}
	if !exists {
		result = fiber.Map{"success": false, "message": "Not subscribed to IRC chat for channel " + channel}
	} else {
		twitch.StopIRCRelay(channel)
		result = fiber.Map{"success": true, "message": "Unsubscribed from IRC chat for channel " + channel}
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return c.SendString("<pre style='background:#161b22;color:#c9d1d9;padding:16px;border-radius:8px;font-size:1.1em;'>" + string(jsonBytes) + "</pre>")
}

// Send message endpoint
func IRCSend(c *fiber.Ctx) error {
	var req struct {
		Channel string `json:"channel"`
		Message string `json:"message"`
	}
	if err := c.BodyParser(&req); err != nil || req.Channel == "" || req.Message == "" {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Missing channel or message"})
	}
	if err := twitch.SendChatMessage(req.Channel, req.Message); err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Message sent as " + os.Getenv("TWITCH_BOT_USERNAME")})
}

// WebsocketHandler moved to server/websocket.go

// SSE stream moved to server/sse.go
