package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	// Microsoft Graph PowerShell public client ID - widely used and authorized
	PublicClientID = "14d82eec-204b-4c2f-b7e8-296a70dab67e"
	// Common tenant allows personal and work accounts
	CommonTenant = "common"
	// Local redirect URI for browser authentication - using port 12345 to avoid conflicts with dev servers
	RedirectURI = "http://localhost:12345/auth/callback"
)

type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	TenantID     string `json:"tenant_id"`
	RedirectURI  string `json:"redirect_uri"`
	UsePublic    bool   `json:"use_public_client"`
}

type TokenStore struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "calendar-widget", "config.json")
}

func GetTokenPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "calendar-widget", "token.json")
}

func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return default public client config if no config exists
		if os.IsNotExist(err) {
			return &Config{
				ClientID:    PublicClientID,
				TenantID:    CommonTenant,
				RedirectURI: RedirectURI,
				UsePublic:   true,
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Migrate old configs to use public client if not explicitly set
	if config.ClientID == "" {
		config.ClientID = PublicClientID
		config.TenantID = CommonTenant
		config.RedirectURI = RedirectURI
		config.UsePublic = true
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

func LoadTokenStore() (*TokenStore, error) {
	tokenPath := GetTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token file exists
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token TokenStore
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return &token, nil
}

func SaveTokenStore(token *TokenStore) error {
	tokenPath := GetTokenPath()
	tokenDir := filepath.Dir(tokenPath)

	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	return os.WriteFile(tokenPath, data, 0600)
}

func IsTokenValid(token *TokenStore) bool {
	if token == nil {
		return false
	}
	// Check if token expires within next 5 minutes
	return time.Now().Add(5 * time.Minute).Before(token.ExpiresAt)
}

func GetCredential() (azcore.TokenCredential, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if config.UsePublic {
		// Use interactive browser authentication for better user experience
		credential, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
			ClientID:    config.ClientID,
			TenantID:    config.TenantID,
			RedirectURL: config.RedirectURI,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create interactive browser credential: %w", err)
		}
		return credential, nil
	}

	// Legacy support for custom app registrations - fallback to device code
	credential, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		ClientID: config.ClientID,
		TenantID: config.TenantID,
		UserPrompt: func(ctx context.Context, message azidentity.DeviceCodeMessage) error {
			fmt.Println(message.Message)
			return nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return credential, nil
}

func GetAccessToken(ctx context.Context) (azcore.AccessToken, error) {
	return GetAccessTokenWithOptions(ctx, false)
}

func GetAccessTokenWithOptions(ctx context.Context, allowInteractive bool) (azcore.AccessToken, error) {
	return GetAccessTokenWithOptionsAndForceRefresh(ctx, allowInteractive, false)
}

func GetAccessTokenWithOptionsAndForceRefresh(ctx context.Context, allowInteractive bool, forceRefresh bool) (azcore.AccessToken, error) {
	// Check for cached token first (unless force refresh is requested)
	if !forceRefresh {
		tokenStore, err := LoadTokenStore()
		if err == nil && IsTokenValid(tokenStore) {
			return azcore.AccessToken{
				Token:     tokenStore.AccessToken,
				ExpiresOn: tokenStore.ExpiresAt,
			}, nil
		}
	}

	// If not interactive and no valid cached token, return error
	if !allowInteractive {
		return azcore.AccessToken{}, fmt.Errorf("authentication required: no valid cached token and interactive login disabled")
	}

	// Get new token
	credential, err := GetCredential()
	if err != nil {
		return azcore.AccessToken{}, err
	}

	token, err := credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/Calendars.Read", "https://graph.microsoft.com/User.Read"},
	})
	if err != nil {
		return azcore.AccessToken{}, fmt.Errorf("failed to get access token: %w", err)
	}

	// Cache the token
	tokenStore := &TokenStore{
		AccessToken: token.Token,
		ExpiresAt:   token.ExpiresOn,
		TokenType:   "Bearer",
	}

	if saveErr := SaveTokenStore(tokenStore); saveErr != nil {
		fmt.Printf("Warning: failed to cache token: %v\n", saveErr)
	}

	return token, nil
}

// ClearTokens removes stored tokens, forcing re-authentication on next use
func ClearTokens() error {
	tokenPath := GetTokenPath()
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}
	return nil
}

// GetGraphServiceClientWithAuth returns a credential for backwards compatibility
func GetGraphServiceClientWithAuth() (azcore.TokenCredential, error) {
	return GetCredential()
}
