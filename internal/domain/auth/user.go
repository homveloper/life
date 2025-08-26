package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/danghamo/life/internal/domain/shared"
)

// UserID represents a unique user identifier
type UserID shared.ID

// NewUserID creates a new user ID
func NewUserID() UserID {
	return UserID(shared.NewID())
}

// String returns string representation
func (id UserID) String() string {
	return string(id)
}

// Username represents a username
type Username struct {
	value string
}

// NewUsername creates a new username
func NewUsername(value string) (Username, error) {
	if len(value) < 3 || len(value) > 20 {
		return Username{}, shared.NewDomainError(shared.ErrCodeInvalidInput, "Username must be between 3 and 20 characters")
	}
	return Username{value: value}, nil
}

// Value returns the username value
func (u Username) Value() string {
	return u.value
}

// HashedPassword represents a bcrypt hashed password
type HashedPassword struct {
	hash string
}

// NewHashedPassword creates a hashed password from plain text
func NewHashedPassword(plainPassword string) (HashedPassword, error) {
	if len(plainPassword) < 6 {
		return HashedPassword{}, shared.NewDomainError(shared.ErrCodeInvalidInput, "Password must be at least 6 characters")
	}
	
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return HashedPassword{}, err
	}
	
	return HashedPassword{hash: string(hash)}, nil
}

// NewHashedPasswordFromHash creates from existing hash
func NewHashedPasswordFromHash(hash string) HashedPassword {
	return HashedPassword{hash: hash}
}

// Hash returns the password hash
func (p HashedPassword) Hash() string {
	return p.hash
}

// Verify checks if the plain password matches the hash
func (p HashedPassword) Verify(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainPassword))
	return err == nil
}

// User represents a user aggregate
type User struct {
	ID        UserID           `json:"id"`
	Username  Username         `json:"username"`
	Password  HashedPassword   `json:"password"`
	CreatedAt shared.Timestamp `json:"created_at"`
	UpdatedAt shared.Timestamp `json:"updated_at"`
}

// NewUser creates a new user
func NewUser(username Username, plainPassword string) (*User, error) {
	hashedPassword, err := NewHashedPassword(plainPassword)
	if err != nil {
		return nil, err
	}
	
	timestamp := shared.NewTimestamp()
	
	return &User{
		ID:        NewUserID(),
		Username:  username,
		Password:  hashedPassword,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}, nil
}

// Authenticate verifies the password
func (u *User) Authenticate(plainPassword string) bool {
	return u.Password.Verify(plainPassword)
}

// ChangePassword changes the user's password
func (u *User) ChangePassword(newPlainPassword string) error {
	newHashedPassword, err := NewHashedPassword(newPlainPassword)
	if err != nil {
		return err
	}
	
	u.Password = newHashedPassword
	u.UpdatedAt = shared.NewTimestamp()
	
	return nil
}