package server

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"go-twitch/twitch"

	irc "github.com/gempir/go-twitch-irc/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// SSEChannelStream streams IRC messages via Server-Sent Events
func SSEChannelStream(c *fiber.Ctx) error {
	channel := c.Params("channel")
	if channel == "" {
		return c.Status(400).SendString("Missing channel name")
	}
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		msgChan := make(chan map[string]interface{}, 100)
		quit := make(chan struct{})

		go twitch.StartIRCRelay(channel)

		botUsername := os.Getenv("TWITCH_BOT_USERNAME")
		botToken := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
		if botUsername == "" || botToken == "" {
			// Notify client once and stop stream
			errMsg := map[string]string{"error": "Bot credentials missing. Set TWITCH_BOT_USERNAME and TWITCH_USER_ACCESS_TOKEN."}
			jsonMsg, _ := json.Marshal(errMsg)
			w.WriteString("data: ")
			w.Write(jsonMsg)
			w.WriteString("\n\n")
			w.Flush()
			return
		}
		ircClient := irc.NewClient(botUsername, "oauth:"+botToken)
		ircClient.Capabilities = []string{irc.TagsCapability, irc.CommandsCapability, irc.MembershipCapability}
		ircClient.OnPrivateMessage(func(m irc.PrivateMessage) {
			if m.Channel == channel {
				msg := map[string]interface{}{
					"user":      m.User.Name,
					"message":   m.Message,
					"channel":   m.Channel,
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				}
				msgChan <- msg
			}
		})
		go func() {
			ircClient.Join(channel)
			ircClient.Connect()
		}()

		for {
			select {
			case msg := <-msgChan:
				jsonMsg, _ := json.Marshal(msg)
				w.WriteString("data: ")
				w.Write(jsonMsg)
				w.WriteString("\n\n")
				w.Flush()
			case <-quit:
				return
			}
		}
	}))
	return nil
}
