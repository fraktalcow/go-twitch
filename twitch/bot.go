package twitch

import (
	"log"
	"os"
	"strings"

	irc "github.com/gempir/go-twitch-irc/v4"
)

var forChannel = "fraktalcow"

func BotCommands() {
	log.Println("[BOT] Starting BotCommands...")
	botUsername := os.Getenv("TWITCH_BOT_USERNAME")
	botToken := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
	log.Printf("[BOT] Username: %s, Token length: %d, Channel: %s", botUsername, len(botToken), forChannel)
	if botUsername == "" || botToken == "" {
		log.Println("[BOT] Bot username or token not set in environment variables.")
		return
	}

	client := irc.NewClient(botUsername, "oauth:"+botToken)
	client.OnPrivateMessage(func(m irc.PrivateMessage) {
		log.Printf("[BOT] Received message in %s from %s: %s", m.Channel, m.User.Name, m.Message)
		if m.Channel == forChannel {
			log.Printf("[BOT] Handling command in %s: %s", m.Channel, m.Message)
			handleCommandsIRC(client, m)
		}
	})
	client.OnConnect(func() {
		log.Printf("[BOT] Bot connected to Twitch IRC, joining #%s", forChannel)
	})
	client.OnNoticeMessage(func(m irc.NoticeMessage) {
		log.Printf("[BOT][NOTICE][%s] %s", m.Channel, m.Message)
	})
	client.Join(forChannel)
	if err := client.Connect(); err != nil {
		log.Printf("[BOT] Bot IRC connection error: %v", err)
	}
}

func handleCommandsIRC(client *irc.Client, msg irc.PrivateMessage) {
	switch {
	case strings.HasPrefix(msg.Message, "!ping"):
		log.Printf("[BOT] Responding to !ping in %s", msg.Channel)
		client.Say(msg.Channel, "pong")
	case strings.HasPrefix(msg.Message, "foo"):
		log.Printf("[BOT] Responding to foo in %s", msg.Channel)
		client.Say(msg.Channel, "bar")
	}
}
