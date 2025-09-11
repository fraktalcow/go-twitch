// auth.go
package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"bufio"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var (
	RedirectURL = os.Getenv("TWITCH_REDIRECT_URI")
)

// In-memory storage for user access token and expiry
type OAuthToken struct {
	AccessToken string
	ExpiresAt   time.Time
}

var oauthToken OAuthToken

// UpdateEnvFile updates or adds a key-value pair in the .env file.
func UpdateEnvFile(key, value string) error {
	file, err := os.OpenFile(".env", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key+"=") {
			lines = append(lines, key+"="+value)
			found = true
		} else {
			lines = append(lines, line)
		}
	}
	if !found {
		lines = append(lines, key+"="+value)
	}
	file.Truncate(0)
	file.Seek(0, 0)
	for _, l := range lines {
		file.WriteString(l + "\n")
	}
	return nil
}

// HandleTwitchAuth redirects the user to Twitch to authorize the application.
func HandleTwitchAuth(w http.ResponseWriter, r *http.Request) {
	scopes := []string{
		"chat:read",
		"chat:edit",
		"user:read:email",
	}
	twitchClientID := os.Getenv("TWITCH_CLIENT_ID")
	authURL := fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		twitchClientID, url.QueryEscape(RedirectURL), url.QueryEscape(strings.Join(scopes, " ")))
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleTwitchCallback handles the callback from Twitch after authorization.
func HandleTwitchCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Removed token and EventSub logic

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// InitAppAccessToken fetches a Twitch app access token and stores it in the .env file
func InitAppAccessToken() error {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("TWITCH_CLIENT_ID or TWITCH_CLIENT_SECRET not set in environment")
	}

	// Prepare request for client credentials flow
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")

	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		return fmt.Errorf("failed to request app access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read app access token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get app access token, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to unmarshal app access token: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return fmt.Errorf("received empty app access token from Twitch")
	}

	// Compute expiry timestamp
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Read existing env, tolerate missing .env by starting fresh
	envMap, err := godotenv.Read(".env")
	if err != nil {
		envMap = map[string]string{}
	}
	envMap["TWITCH_APP_ACCESS_TOKEN"] = tokenResp.AccessToken
	envMap["TWITCH_APP_ACCESS_TOKEN_EXPIRES_AT"] = expiresAt.Format(time.RFC3339)
	if err := godotenv.Write(envMap, ".env"); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	// Export to current process environment as well
	os.Setenv("TWITCH_APP_ACCESS_TOKEN", tokenResp.AccessToken)
	os.Setenv("TWITCH_APP_ACCESS_TOKEN_EXPIRES_AT", expiresAt.Format(time.RFC3339))

	return nil
}

func AuthorizeHandler(c *fiber.Ctx) error {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	redirectURI := os.Getenv("TWITCH_REDIRECT_URI")
	scopes := "chat:read chat:edit user:read:email"
	authURL := "https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=" + url.QueryEscape(clientID) + "&redirect_uri=" + url.QueryEscape(redirectURI) + "&scope=" + url.QueryEscape(scopes)
	return c.Redirect(authURL)
}

func AuthCallbackHandler(c *fiber.Ctx) error {
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
	oauthToken.AccessToken = tokenResp.AccessToken
	oauthToken.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	_ = UpdateEnvFile("TWITCH_USER_ACCESS_TOKEN", oauthToken.AccessToken)
	_ = UpdateEnvFile("TWITCH_USER_ACCESS_TOKEN_EXPIRES_AT", oauthToken.ExpiresAt.Format(time.RFC3339))
	os.Setenv("TWITCH_USER_ACCESS_TOKEN", oauthToken.AccessToken)
	os.Setenv("TWITCH_USER_ACCESS_TOKEN_EXPIRES_AT", oauthToken.ExpiresAt.Format(time.RFC3339))
	return c.SendString("<h2>Authorized!</h2><p>Access token set. Expires at: " + oauthToken.ExpiresAt.Format(time.RFC1123) + "</p>")
}

func AuthStatusHandler(c *fiber.Ctx) error {
	if oauthToken.AccessToken == "" {
		return c.JSON(fiber.Map{"authorized": false})
	}
	return c.JSON(fiber.Map{
		"authorized": true,
		"expires_at": oauthToken.ExpiresAt,
	})
}

func CallbackAliasHandler(c *fiber.Ctx) error {
	return c.Redirect("/auth/callback?" + c.Context().QueryArgs().String())
}
