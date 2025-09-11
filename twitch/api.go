package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	irc "github.com/gempir/go-twitch-irc/v4"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

const baseURL = "https://api.twitch.tv/helix"

// UserResponse represents the structure of the Twitch API /users response
type UserResponse struct {
	Data []struct {
		ID              string `json:"id"`
		Login           string `json:"login"`
		DisplayName     string `json:"display_name"`
		Type            string `json:"type"`
		BroadcasterType string `json:"broadcaster_type"`
		Description     string `json:"description"`
		ProfileImageURL string `json:"profile_image_url"`
		OfflineImageURL string `json:"offline_image_url"`
		ViewCount       int    `json:"view_count"`
		Email           string `json:"email"`
		CreatedAt       string `json:"created_at"`
	} `json:"data"`
}

// StreamResponse represents the structure of the Twitch API /streams response
type StreamResponse struct {
	Data []struct {
		ID           string   `json:"id"`
		UserID       string   `json:"user_id"`
		UserLogin    string   `json:"user_login"`
		UserName     string   `json:"user_name"`
		GameID       string   `json:"game_id"`
		GameName     string   `json:"game_name"`
		Type         string   `json:"type"`
		Title        string   `json:"title"`
		ViewerCount  int      `json:"viewer_count"`
		StartedAt    string   `json:"started_at"`
		Language     string   `json:"language"`
		ThumbnailURL string   `json:"thumbnail_url"`
		TagIDs       []string `json:"tag_ids"`
	} `json:"data"`
}

// TopGamesResponse represents the structure of the Twitch API /games/top response
type TopGamesResponse struct {
	Data []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		BoxArtURL string `json:"box_art_url"`
	} `json:"data"`
}

// GetUserAccessToken returns the Twitch user access token from environment variables
func GetUserAccessToken() (string, error) {
	token := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
	if token == "" {
		return "", fmt.Errorf("TWITCH_USER_ACCESS_TOKEN environment variable not set")
	}
	return token, nil
}

func GetUserInfo(username string) (*UserResponse, error) {
	token, err := GetUserAccessToken()
	if err != nil {
		return nil, err
	}

	clientID := GetClientID()
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID environment variable not set")
	}

	req, err := http.NewRequest("GET", baseURL+"/users?login="+username, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Client-Id", clientID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userRes UserResponse
	err = json.Unmarshal(bodyBytes, &userRes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &userRes, nil
}

func GetStreamInfo(username string) (*StreamResponse, error) {
	token, err := GetUserAccessToken()
	if err != nil {
		token, err = GetAccessToken()
		if err != nil {
			return nil, err
		}
	}

	clientID := GetClientID()
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID environment variable not set")
	}

	req, err := http.NewRequest("GET", baseURL+"/streams?user_login="+username, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Client-Id", clientID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get stream info: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var streamRes StreamResponse
	err = json.Unmarshal(bodyBytes, &streamRes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal stream info: %w", err)
	}

	return &streamRes, nil
}

func GetTopGames() (*TopGamesResponse, error) {
	token, err := GetAccessToken()
	if err != nil {
		return nil, err
	}

	clientID := GetClientID()
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID environment variable not set")
	}

	req, err := http.NewRequest("GET", baseURL+"/games/top", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Client-Id", clientID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get top games: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get top games: status %d, body %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var topGamesRes TopGamesResponse
	err = json.Unmarshal(bodyBytes, &topGamesRes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal top games: %w", err)
	}

	return &topGamesRes, nil
}

// GetAccessToken returns the Twitch app access token from environment variables
func GetAccessToken() (string, error) {
	token := os.Getenv("TWITCH_APP_ACCESS_TOKEN")
	if token == "" {
		return "", fmt.Errorf("TWITCH_APP_ACCESS_TOKEN environment variable not set")
	}
	return token, nil
}

// GetClientID returns the Twitch client ID from environment variables
func GetClientID() string {
	return os.Getenv("TWITCH_CLIENT_ID")
}

var IRCClients = make(map[string]*irc.Client)
var IRCClientsMu sync.Mutex

var SSESubscribers = make(map[string]map[chan []byte]struct{})
var SSESubscribersMu sync.Mutex

type IRCClient struct {
	*websocket.Conn
}

// StartIRCRelay starts relaying IRC messages for a channel to WebSocket clients only.
func StartIRCRelay(channel string) {
	IRCClientsMu.Lock()
	if _, exists := IRCClients[channel]; exists {
		IRCClientsMu.Unlock()
		return // Already joined
	}
	botUsername := os.Getenv("TWITCH_BOT_USERNAME")
	botToken := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
	client := irc.NewClient(botUsername, "oauth:"+botToken)
	IRCClients[channel] = client
	IRCClientsMu.Unlock()

	client.OnPrivateMessage(func(msg irc.PrivateMessage) {
		//  broadcast to WebSocket
		// (No-op or add your own relay logic here if needed)
	})
	client.OnNoticeMessage(func(m irc.NoticeMessage) {
		log.Printf("[IRC][NOTICE][%s] %s", m.Channel, m.Message)
	})
	go func() {
		client.Join(channel)
		client.Connect()
	}()
}

// StopIRCRelay disconnects and removes the IRC client for a channel.
func StopIRCRelay(channel string) {
	IRCClientsMu.Lock()
	client, exists := IRCClients[channel]
	if exists {
		client.Disconnect()
		delete(IRCClients, channel)
	}
	IRCClientsMu.Unlock()
}

// Handler functions for API endpoints
func UserInfoHandler(c *fiber.Ctx) error {
	username := c.Params("name")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username parameter is missing"})
	}
	data, err := GetUserInfo(username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

func StreamInfoHandler(c *fiber.Ctx) error {
	username := c.Params("name")
	if username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Streamer name parameter is missing"})
	}
	data, err := GetStreamInfo(username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

func TopGamesHandler(c *fiber.Ctx) error {
	data, err := GetTopGames()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

// SendChatMessage sends a message to the specified channel using the bot IRC client.
func SendChatMessage(channel, message string) error {
	IRCClientsMu.Lock()
	client, exists := IRCClients[channel]
	if !exists {
		// If not joined, join first
		botUsername := os.Getenv("TWITCH_BOT_USERNAME")
		botToken := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
		client = irc.NewClient(botUsername, "oauth:"+botToken)
		IRCClients[channel] = client
		IRCClientsMu.Unlock()
		go func() {
			client.Join(channel)
			client.Connect()
		}()

		IRCClientsMu.Lock()
		client = IRCClients[channel]
		IRCClientsMu.Unlock()
	} else {
		IRCClientsMu.Unlock()
	}
	if client == nil {
		return fmt.Errorf("IRC client not available for channel %s", channel)
	}
	client.Say(channel, message)
	return nil
}
