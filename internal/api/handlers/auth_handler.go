package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/api/middleware"
	"github.com/danghamo/life/internal/domain/account"
	"github.com/danghamo/life/pkg/logger"
)

// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
	Google   ProviderConfig `json:"google"`
	GitHub   ProviderConfig `json:"github"`
	Discord  ProviderConfig `json:"discord"`
}

// ProviderConfig represents OAuth provider configuration
type ProviderConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
	Scopes       string `json:"scopes"`
}

// AuthHandler handles OAuth authentication
type AuthHandler struct {
	logger      *logger.Logger
	accountRepo account.Repository
	jwtService  *account.JWTService
	oauthConfig OAuthConfig
	httpClient  *http.Client
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	logger *logger.Logger,
	accountRepo account.Repository,
	jwtService *account.JWTService,
	oauthConfig OAuthConfig,
) *AuthHandler {
	return &AuthHandler{
		logger:      logger.WithComponent("auth-handler"),
		accountRepo: accountRepo,
		jwtService:  jwtService,
		oauthConfig: oauthConfig,
		httpClient:  &http.Client{},
	}
}

// OAuthStartRequest represents OAuth start request
type OAuthStartRequest struct {
	Provider string `json:"provider"`
	State    string `json:"state,omitempty"`
}

// OAuthStartResponse represents OAuth start response
type OAuthStartResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// OAuthCallbackRequest represents OAuth callback request
type OAuthCallbackRequest struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

// OAuthCallbackResponse represents OAuth callback response
type OAuthCallbackResponse struct {
	JWTToken  string `json:"jwt_token"`
	UserID    string `json:"user_id"`
	ExpiresIn int64  `json:"expires_in"`
}

// GuestLoginRequest represents guest login request
type GuestLoginRequest struct {
	DeviceID string `json:"device_id"`
}

// GuestLoginResponse represents guest login response
type GuestLoginResponse struct {
	JWTToken  string `json:"jwt_token"`
	UserID    string `json:"user_id"`
	IsGuest   bool   `json:"is_guest"`
	ExpiresIn int64  `json:"expires_in"`
}

// LinkSocialRequest represents social account linking request
type LinkSocialRequest struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	State    string `json:"state"`
}

// OAuthTokenResponse represents OAuth token response from provider
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// UserProfile represents user profile from OAuth provider
type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// HandleOAuthStart handles POST /api/v1/auth.OAuthStart
func (h *AuthHandler) HandleOAuthStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params OAuthStartRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	provider := account.Provider(params.Provider)
	if !provider.IsValid() {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid provider")
		return
	}

	config, err := h.getProviderConfig(provider)
	if err != nil {
		h.logger.Error("Provider configuration error", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Provider configuration error")
		return
	}

	state := params.State
	if state == "" {
		state = generateRandomState()
	}

	authURL := h.buildAuthURL(config, state)

	response := OAuthStartResponse{
		AuthURL: authURL,
		State:   state,
	}

	h.logger.Info("OAuth flow started", zap.String("provider", params.Provider))
	jsonrpcx.Success(w, req.ID, response)
}

// HandleOAuthCallback handles POST /api/v1/auth.OAuthCallback
func (h *AuthHandler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params OAuthCallbackRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	provider := account.Provider(params.Provider)
	if !provider.IsValid() {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid provider")
		return
	}

	config, err := h.getProviderConfig(provider)
	if err != nil {
		h.logger.Error("Provider configuration error", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Provider configuration error")
		return
	}

	// Exchange code for access token
	token, err := h.exchangeCodeForToken(config, params.Code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Failed to exchange code for token")
		return
	}

	// Get user profile
	profile, err := h.getUserProfile(config, token.AccessToken)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Failed to get user profile")
		return
	}

	// Create or get account
	acc, err := h.getOrCreateAccount(r.Context(), provider, profile)
	if err != nil {
		h.logger.Error("Failed to create account", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to create account")
		return
	}

	// Generate JWT token
	jwtToken, err := h.jwtService.GenerateToken(acc)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to generate JWT token")
		return
	}

	response := OAuthCallbackResponse{
		JWTToken:  jwtToken,
		UserID:    acc.UserID.String(),
		ExpiresIn: 86400, // 24 hours
	}

	h.logger.Info("User authenticated successfully", 
		zap.String("userId", acc.UserID.String()),
		zap.String("provider", string(provider)))
	jsonrpcx.Success(w, req.ID, response)
}

// HandleGuestLogin handles guest login
func (h *AuthHandler) HandleGuestLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params GuestLoginRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	if params.DeviceID == "" {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Device ID is required")
		return
	}

	// Try to find existing guest account
	existingAccount, err := h.accountRepo.GetByDeviceID(r.Context(), params.DeviceID)
	if err != nil {
		h.logger.Error("Failed to get account by device ID", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Internal error")
		return
	}

	var guestAccount *account.Account

	if existingAccount != nil {
		guestAccount = existingAccount
		h.logger.Info("Existing guest account found", zap.String("deviceId", params.DeviceID))
	} else {
		// Create new guest account
		newAccount, err := account.NewGuestAccount(params.DeviceID)
		if err != nil {
			h.logger.Error("Failed to create guest account", zap.Error(err))
			jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to create guest account")
			return
		}

		err = h.accountRepo.FindOneAndInsert(r.Context(), newAccount.ID, func() (*account.Account, error) {
			return newAccount, nil
		})
		if err != nil {
			h.logger.Error("Failed to save guest account", zap.Error(err))
			jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to save guest account")
			return
		}

		guestAccount = newAccount
		h.logger.Info("New guest account created", 
			zap.String("deviceId", params.DeviceID),
			zap.String("userId", guestAccount.UserID.String()))
	}

	// Generate JWT token
	jwtToken, err := h.jwtService.GenerateToken(guestAccount)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to generate JWT token")
		return
	}

	response := GuestLoginResponse{
		JWTToken:  jwtToken,
		UserID:    guestAccount.UserID.String(),
		IsGuest:   true,
		ExpiresIn: 86400, // 24 hours
	}

	jsonrpcx.Success(w, req.ID, response)
}

// HandleLinkSocial handles linking guest account to social provider
func (h *AuthHandler) HandleLinkSocial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonrpcx.WithError(r, nil, jsonrpcx.MethodNotFound, "Method not allowed")
		return
	}

	// JWT에서 사용자 정보 가져오기
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		jsonrpcx.WithError(r, nil, jsonrpcx.InvalidRequest, "User not authenticated")
		return
	}

	req, err := jsonrpcx.ParseRequest(r)
	if err != nil {
		jsonrpcx.WithError(r, nil, jsonrpcx.ParseError, "Invalid JSON-RPC request")
		return
	}

	var params LinkSocialRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid params")
		return
	}

	provider := account.Provider(params.Provider)
	if !provider.IsValid() || provider == account.ProviderGuest {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Invalid social provider")
		return
	}

	// Get existing account
	existingAccount, err := h.accountRepo.GetByUserID(r.Context(), account.UserID(userID))
	if err != nil {
		h.logger.Error("Failed to get account", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Internal error")
		return
	}

	if existingAccount == nil || !existingAccount.IsGuest() {
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Only guest accounts can be linked")
		return
	}

	// OAuth 플로우와 동일하게 토큰 교환 및 사용자 정보 가져오기
	config, err := h.getProviderConfig(provider)
	if err != nil {
		h.logger.Error("Provider configuration error", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Provider configuration error")
		return
	}

	token, err := h.exchangeCodeForToken(config, params.Code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Failed to exchange code for token")
		return
	}

	profile, err := h.getUserProfile(config, token.AccessToken)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InvalidParams, "Failed to get user profile")
		return
	}

	// 소셜 계정 연동
	oauthProfile := account.NewOAuthProfile(profile.ID, profile.Email, profile.Name)
	if profile.AvatarURL != "" {
		oauthProfile.AvatarURL = profile.AvatarURL
	}

	err = h.accountRepo.FindOneAndUpdate(r.Context(), existingAccount.ID, func(acc *account.Account) (*account.Account, error) {
		return acc, acc.LinkToSocialProvider(provider, oauthProfile)
	})
	if err != nil {
		h.logger.Error("Failed to link social account", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to link social account")
		return
	}

	// 새로운 JWT 토큰 생성
	updatedAccount, err := h.accountRepo.GetByID(r.Context(), existingAccount.ID)
	if err != nil {
		h.logger.Error("Failed to get updated account", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Internal error")
		return
	}

	jwtToken, err := h.jwtService.GenerateToken(updatedAccount)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		jsonrpcx.WithError(r, req.ID, jsonrpcx.InternalError, "Failed to generate JWT token")
		return
	}

	response := OAuthCallbackResponse{
		JWTToken:  jwtToken,
		UserID:    updatedAccount.UserID.String(),
		ExpiresIn: 86400,
	}

	h.logger.Info("Guest account linked to social provider", 
		zap.String("userId", userID),
		zap.String("provider", string(provider)))

	jsonrpcx.Success(w, req.ID, response)
}

// getOrCreateAccount gets existing account or creates new one with N:1 UserID linking
func (h *AuthHandler) getOrCreateAccount(ctx context.Context, provider account.Provider, profile *UserProfile) (*account.Account, error) {
	// Try to find existing account for this specific provider
	existing, err := h.accountRepo.GetByProvider(ctx, provider, profile.ID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update profile if needed
		oauthProfile := account.NewOAuthProfile(profile.ID, profile.Email, profile.Name)
		if profile.AvatarURL != "" {
			oauthProfile.AvatarURL = profile.AvatarURL
		}

		err = h.accountRepo.FindOneAndUpdate(ctx, existing.ID, func(acc *account.Account) (*account.Account, error) {
			return acc, acc.UpdateProfile(oauthProfile)
		})
		if err != nil {
			return nil, err
		}

		return existing, nil
	}

	// Create OAuth profile for new account
	oauthProfile := account.NewOAuthProfile(profile.ID, profile.Email, profile.Name)
	if profile.AvatarURL != "" {
		oauthProfile.AvatarURL = profile.AvatarURL
	}

	// Check if any account with same email exists (for N:1 UserID linking)
	var newAccount *account.Account
	if profile.Email != "" {
		existingByEmail, err := h.accountRepo.GetByEmail(ctx, profile.Email)
		if err != nil {
			return nil, err
		}
		
		if existingByEmail != nil {
			// Link to existing UserID (N:1 relationship)
			h.logger.Info("Linking new provider to existing UserID",
				zap.String("email", profile.Email),
				zap.String("newProvider", string(provider)),
				zap.String("existingUserID", existingByEmail.UserID.String()))
				
			newAccount, err = account.NewAccountWithUserID(provider, oauthProfile, existingByEmail.UserID)
		} else {
			// Create completely new account with new UserID
			newAccount, err = account.NewAccount(provider, oauthProfile)
		}
	} else {
		// No email provided, create new account
		newAccount, err = account.NewAccount(provider, oauthProfile)
	}
	
	if err != nil {
		return nil, err
	}

	err = h.accountRepo.FindOneAndInsert(ctx, newAccount.ID, func() (*account.Account, error) {
		return newAccount, nil
	})
	if err != nil {
		return nil, err
	}

	return newAccount, nil
}

// getProviderConfig gets OAuth configuration for provider
func (h *AuthHandler) getProviderConfig(provider account.Provider) (*ProviderConfig, error) {
	switch provider {
	case account.ProviderGoogle:
		return &h.oauthConfig.Google, nil
	case account.ProviderGitHub:
		return &h.oauthConfig.GitHub, nil
	case account.ProviderDiscord:
		return &h.oauthConfig.Discord, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// buildAuthURL builds OAuth authorization URL
func (h *AuthHandler) buildAuthURL(config *ProviderConfig, state string) string {
	params := url.Values{}
	params.Set("client_id", config.ClientID)
	params.Set("redirect_uri", config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", config.Scopes)
	params.Set("state", state)

	return config.AuthURL + "?" + params.Encode()
}

// exchangeCodeForToken exchanges authorization code for access token
func (h *AuthHandler) exchangeCodeForToken(config *ProviderConfig, code string) (*OAuthTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.RedirectURI)

	req, err := http.NewRequest("POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// getUserProfile gets user profile from OAuth provider
func (h *AuthHandler) getUserProfile(config *ProviderConfig, accessToken string) (*UserProfile, error) {
	req, err := http.NewRequest("GET", config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// generateRandomState generates a random state for OAuth flow
func generateRandomState() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback if crypto/rand fails
		return fmt.Sprintf("state_%d", mathrand.Int63())
	}
	return hex.EncodeToString(bytes)
}