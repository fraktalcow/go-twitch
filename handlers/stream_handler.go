package handlers

import (
	"go-twitch/twitch"

	"github.com/gofiber/fiber/v2"
)

func GetStream(c *fiber.Ctx) error {
	username := c.Params("name")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Streamer name parameter is missing"})
	}
	data, err := twitch.GetStreamInfo(username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

func GetTopGames(c *fiber.Ctx) error {
	data, err := twitch.GetTopGames()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}
