package handlers

import (
	"go-twitch/twitch"

	"github.com/gofiber/fiber/v2"
)

func GetUser(c *fiber.Ctx) error {
	username := c.Params("name")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username parameter is missing"})
	}
	data, err := twitch.GetUserInfo(username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}
