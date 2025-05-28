package models

import (
	"time"
)

// User represents the main user entity
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Username    string    `json:"username"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
	Status      string    `json:"status" gorm:"default:'active'"` // active/banned/deleted
}

// OAuthAccount represents a third-party OAuth account linked to a user
type OAuthAccount struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id"`
	User             User      `json:"-" gorm:"foreignKey:UserID"`
	Provider         string    `json:"provider" gorm:"index:idx_provider_user_id,unique"` // linuxdo/github/google/wechat
	ProviderUserID   string    `json:"provider_user_id" gorm:"index:idx_provider_user_id,unique"`
	ProviderUsername string    `json:"provider_username"`
	ProviderEmail    string    `json:"provider_email"`
	ProviderAvatar   string    `json:"provider_avatar"`
	AccessToken      string    `json:"-"` // Stored encrypted
	RefreshToken     *string   `json:"-"` // Stored encrypted, 可能为空
	TokenExpiresAt   time.Time `json:"token_expires_at"`
	CreatedAt        time.Time `json:"created_at"`
	LastUsedAt       time.Time `json:"last_used_at"`
	Status           string    `json:"status" gorm:"default:'active'"` // active/revoked
}

// OAuthProvider represents a configured OAuth provider in the system
type OAuthProvider struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"unique"` // linuxdo/github/google
	DisplayName  string    `json:"display_name"`
	ClientID     string    `json:"client_id"`
	ClientSecret *string   `json:"-"` // Stored encrypted, 可能为空
	AuthURL      string    `json:"auth_url"`
	TokenURL     string    `json:"token_url"`
	UserInfoURL  string    `json:"user_info_url"`
	Scope        string    `json:"scope"`
	Enabled      bool      `json:"enabled" gorm:"default:true"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}
