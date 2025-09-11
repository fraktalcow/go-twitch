package server

import (
	"encoding/json"
	"log"
	"os"

	irc "github.com/gempir/go-twitch-irc/v4"
	"github.com/gofiber/contrib/websocket"
)

// WebsocketHandler handles /ws chat relay
func WebsocketHandler(c *websocket.Conn) {
	defer c.Close()
	monitored := make(map[string]*irc.Client)
	msgChan := make(chan []byte, 100)
	quit := make(chan struct{})

	// Per-connection event preferences (defaults enabled)
	prefs := struct {
		Notice     bool
		UserNotice bool
		ClearChat  bool
		RoomState  bool
	}{Notice: true, UserNotice: true, ClearChat: true, RoomState: true}

	go func() {
		for msg := range msgChan {
			c.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	for {
		msgType, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
		if msgType == websocket.TextMessage {
			var cmd struct {
				Action  string          `json:"action"`
				Channel string          `json:"channel"`
				Prefs   map[string]bool `json:"prefs"`
			}
			if err := json.Unmarshal(msg, &cmd); err == nil {
				switch cmd.Action {
				case "subscribe":
					if _, ok := monitored[cmd.Channel]; ok {
						continue
					}
					botUsername := os.Getenv("TWITCH_BOT_USERNAME")
					botToken := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
					ircClient := irc.NewClient(botUsername, "oauth:"+botToken)
					ircClient.Capabilities = []string{irc.TagsCapability, irc.CommandsCapability, irc.MembershipCapability}
					ircClient.OnPrivateMessage(func(m irc.PrivateMessage) {
						if m.Channel == cmd.Channel {
							chatMsg := map[string]interface{}{
								"user":    m.User.Name,
								"message": m.Message,
								"channel": m.Channel,
							}
							jsonMsg, _ := json.Marshal(chatMsg)
							select {
							case msgChan <- jsonMsg:
							default:
							}
						}
					})
					// Relay generic NOTICE messages
					ircClient.OnNoticeMessage(func(m irc.NoticeMessage) {
						if m.Channel == cmd.Channel {
							if prefs.Notice {
								notice := map[string]interface{}{
									"type":    "notice",
									"channel": m.Channel,
									"system":  m.Message,
								}
								jsonMsg, _ := json.Marshal(notice)
								select {
								case msgChan <- jsonMsg:
								default:
								}
							}
						}
					})
					// Relay USERNOTICE events (subs, resubs, gifts, raids, etc.)
					ircClient.OnUserNoticeMessage(func(m irc.UserNoticeMessage) {
						if m.Channel == cmd.Channel {
							if prefs.UserNotice {
								notice := map[string]interface{}{
									"type":    "usernotice",
									"channel": m.Channel,
									"system":  m.SystemMsg,
									"msg_id":  m.MsgID,
								}
								jsonMsg, _ := json.Marshal(notice)
								select {
								case msgChan <- jsonMsg:
								default:
								}
							}
						}
					})
					// Relay CLEARCHAT (timeouts/bans)
					ircClient.OnClearChatMessage(func(m irc.ClearChatMessage) {
						if m.Channel == cmd.Channel {
							if prefs.ClearChat {
								payload := map[string]interface{}{
									"type":    "clearchat",
									"channel": m.Channel,
								}
								jsonMsg, _ := json.Marshal(payload)
								select {
								case msgChan <- jsonMsg:
								default:
								}
							}
						}
					})
					// Relay ROOMSTATE changes (slow mode, emote-only, etc.)
					ircClient.OnRoomStateMessage(func(m irc.RoomStateMessage) {
						if m.Channel == cmd.Channel {
							if prefs.RoomState {
								payload := map[string]interface{}{
									"type":    "roomstate",
									"channel": m.Channel,
								}
								jsonMsg, _ := json.Marshal(payload)
								select {
								case msgChan <- jsonMsg:
								default:
								}
							}
						}
					})
					monitored[cmd.Channel] = ircClient
					go func(channel string) {
						ircClient.Join(channel)
						if err := ircClient.Connect(); err != nil {
							log.Printf("IRC client connection error for channel %s: %v", channel, err)
						}
					}(cmd.Channel)
					// send subscription acknowledgement to the client
					ack := map[string]interface{}{
						"type":    "subscribed",
						"channel": cmd.Channel,
					}
					ackBytes, _ := json.Marshal(ack)
					select {
					case msgChan <- ackBytes:
					default:
					}
				case "unsubscribe":
					if ircClient, ok := monitored[cmd.Channel]; ok {
						ircClient.Disconnect()
						delete(monitored, cmd.Channel)
					}
				case "setPreferences":
					if cmd.Prefs != nil {
						if v, ok := cmd.Prefs["notice"]; ok {
							prefs.Notice = v
						}
						if v, ok := cmd.Prefs["usernotice"]; ok {
							prefs.UserNotice = v
						}
						if v, ok := cmd.Prefs["clearchat"]; ok {
							prefs.ClearChat = v
						}
						if v, ok := cmd.Prefs["roomstate"]; ok {
							prefs.RoomState = v
						}
					}
				}
			}
		}
	}
	for _, ircClient := range monitored {
		ircClient.Disconnect()
	}
	close(msgChan)
	close(quit)
}
