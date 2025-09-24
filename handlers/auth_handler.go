package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"go-twitch/twitch"

	"github.com/gofiber/fiber/v2"
)

func AuthorizePage(c *fiber.Ctx) error {
	return c.SendFile("./static/authorize.html")
}

func AuthStart(c *fiber.Ctx) error {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	redirectURI := os.Getenv("TWITCH_REDIRECT_URI")
	scopes := "chat:read chat:edit user:read:email"
	state := "secure_random_state"
	authURL := "https://id.twitch.tv/oauth2/authorize" +
		"?client_id=" + url.QueryEscape(clientID) +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&response_type=code" +
		"&scope=" + url.QueryEscape(scopes) +
		"&state=" + url.QueryEscape(state)
	return c.Redirect(authURL)
}

func AuthCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(400).SendString("Missing code parameter")
	}
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	redirectURI := os.Getenv("TWITCH_REDIRECT_URI")
	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return c.Status(500).SendString("Failed to exchange code: " + err.Error())
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return c.Status(500).SendString("Twitch token exchange failed: " + string(body))
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return c.Status(500).SendString("Failed to parse token response: " + err.Error())
	}
	// preserve existing behavior writing to env through twitch.UpdateEnvFile
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	_ = twitch.UpdateEnvFile("TWITCH_USER_ACCESS_TOKEN", tokenResp.AccessToken)
	_ = twitch.UpdateEnvFile("TWITCH_USER_ACCESS_TOKEN_EXPIRES_AT", expiresAt.Format(time.RFC3339))
	os.Setenv("TWITCH_USER_ACCESS_TOKEN", tokenResp.AccessToken)
	os.Setenv("TWITCH_USER_ACCESS_TOKEN_EXPIRES_AT", expiresAt.Format(time.RFC3339))
	return c.Redirect("/authorize")
}

func AuthStatus(c *fiber.Ctx) error {
	userToken := os.Getenv("TWITCH_USER_ACCESS_TOKEN")
	userExpiresStr := os.Getenv("TWITCH_USER_ACCESS_TOKEN_EXPIRES_AT")
	appToken := os.Getenv("TWITCH_APP_ACCESS_TOKEN")
	appExpiresStr := os.Getenv("TWITCH_APP_ACCESS_TOKEN_EXPIRES_AT")
	
	var userExpires time.Time
	var appExpires time.Time
	var appAuthorized bool
	
	if userToken != "" && userExpiresStr != "" {
		if parsed, err := time.Parse(time.RFC3339, userExpiresStr); err == nil {
			userExpires = parsed
		}
	}
	
	if appToken != "" && appExpiresStr != "" {
		if parsed, err := time.Parse(time.RFC3339, appExpiresStr); err == nil {
			appExpires = parsed
			appAuthorized = true
		}
	}
	
	return c.JSON(fiber.Map{
		"authorized":      userToken != "",
		"user_expires_at": userExpires,
		"app_authorized":  appAuthorized,
		"app_expires_at":  appExpires,
	})
}

func CallbackAlias(c *fiber.Ctx) error {
	return c.Redirect("/auth/callback?" + c.Context().QueryArgs().String())
}
