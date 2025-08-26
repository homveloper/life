package account

import (
	"github.com/danghamo/life/internal/domain/shared"
)

// AccountID represents a unique account identifier
type AccountID shared.ID

// NewAccountID creates a new account ID
func NewAccountID() AccountID {
	return AccountID(shared.NewID())
}

// String returns string representation
func (id AccountID) String() string {
	return string(id)
}

// UserID represents a unique user identifier for game domain
type UserID shared.ID

// NewUserID creates a new user ID
func NewUserID() UserID {
	return UserID(shared.NewID())
}

// String returns string representation
func (id UserID) String() string {
	return string(id)
}

// Provider represents OAuth provider
type Provider string

const (
	ProviderGuest   Provider = "guest"   // 게스트 로그인
	ProviderGoogle  Provider = "google"
	ProviderGitHub  Provider = "github"
	ProviderDiscord Provider = "discord"
	ProviderApple   Provider = "apple"   // 애플 로그인 추가
)

// String returns string representation
func (p Provider) String() string {
	return string(p)
}

// IsValid checks if provider is valid
func (p Provider) IsValid() bool {
	return p == ProviderGuest || p == ProviderGoogle || p == ProviderGitHub || 
		   p == ProviderDiscord || p == ProviderApple
}

// OAuthProfile represents OAuth user profile information
type OAuthProfile struct {
	ProviderUserID string `json:"provider_user_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	AvatarURL      string `json:"avatar_url,omitempty"`
}

// NewOAuthProfile creates a new OAuth profile
func NewOAuthProfile(providerUserID, email, name string) OAuthProfile {
	return OAuthProfile{
		ProviderUserID: providerUserID,
		Email:          email,
		Name:           name,
	}
}

// Account represents an account aggregate (OAuth-based authentication)
type Account struct {
	ID          AccountID         `json:"id"`
	UserID      UserID           `json:"user_id"`    // Game domain identifier
	Provider    Provider         `json:"provider"`
	Profile     OAuthProfile     `json:"profile"`
	DeviceID    string           `json:"device_id,omitempty"` // 게스트용 기기 식별자
	LinkedAt    *shared.Timestamp `json:"linked_at,omitempty"` // 소셜 연동 시점
	CreatedAt   shared.Timestamp `json:"created_at"`
	UpdatedAt   shared.Timestamp `json:"updated_at"`
}

// NewAccount creates a new account from OAuth profile with new UserID
func NewAccount(provider Provider, profile OAuthProfile) (*Account, error) {
	return NewAccountWithUserID(provider, profile, NewUserID())
}

// NewAccountWithUserID creates a new account from OAuth profile with existing UserID (for N:1 linking)
func NewAccountWithUserID(provider Provider, profile OAuthProfile, userID UserID) (*Account, error) {
	if !provider.IsValid() {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Invalid OAuth provider")
	}

	if profile.ProviderUserID == "" {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Provider user ID cannot be empty")
	}

	if profile.Email == "" {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Email cannot be empty")
	}

	timestamp := shared.NewTimestamp()

	return &Account{
		ID:        NewAccountID(),
		UserID:    userID, // Use provided UserID (could be new or existing)
		Provider:  provider,
		Profile:   profile,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}, nil
}

// NewGuestAccount creates a new guest account with device identifier
func NewGuestAccount(deviceID string) (*Account, error) {
	if deviceID == "" {
		return nil, shared.NewDomainError(shared.ErrCodeInvalidInput, "Device ID cannot be empty for guest account")
	}

	timestamp := shared.NewTimestamp()
	
	// 게스트용 프로필 생성 (deviceID를 ProviderUserID로 사용)
	profile := OAuthProfile{
		ProviderUserID: deviceID,
		Email:          "", // 게스트는 이메일 없음
		Name:           "Guest User",
		AvatarURL:      "",
	}

	return &Account{
		ID:        NewAccountID(),
		UserID:    NewUserID(),
		Provider:  ProviderGuest,
		Profile:   profile,
		DeviceID:  deviceID,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}, nil
}

// IsGuest checks if the account is a guest account
func (a *Account) IsGuest() bool {
	return a.Provider == ProviderGuest
}

// CanLinkToSocial checks if guest account can be linked to social provider
func (a *Account) CanLinkToSocial() bool {
	return a.IsGuest() && a.LinkedAt == nil
}

// LinkToSocialProvider converts guest account to social account
func (a *Account) LinkToSocialProvider(provider Provider, profile OAuthProfile) error {
	if !a.CanLinkToSocial() {
		return shared.NewDomainError(shared.ErrCodeInvalidOperation, "Account cannot be linked to social provider")
	}

	if !provider.IsValid() || provider == ProviderGuest {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Invalid social provider")
	}

	if profile.ProviderUserID == "" || profile.Email == "" {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Social profile must have provider user ID and email")
	}

	timestamp := shared.NewTimestamp()
	
	a.Provider = provider
	a.Profile = profile
	a.LinkedAt = &timestamp
	a.UpdatedAt = timestamp

	return nil
}

// UpdateProfile updates the OAuth profile information
func (a *Account) UpdateProfile(profile OAuthProfile) error {
	if profile.ProviderUserID == "" {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Provider user ID cannot be empty")
	}

	if profile.Email == "" {
		return shared.NewDomainError(shared.ErrCodeInvalidInput, "Email cannot be empty")
	}

	a.Profile = profile
	a.UpdatedAt = shared.NewTimestamp()

	return nil
}

// GetProviderKey returns the unique key for provider + provider_user_id
func (a *Account) GetProviderKey() string {
	return string(a.Provider) + ":" + a.Profile.ProviderUserID
}