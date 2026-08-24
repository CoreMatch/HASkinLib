package models

import (
	"time"
)

// User 映射 HRPAuth 的 users 表（只读复用，表由 HRPAuth 迁移创建）。
type User struct {
	UID           uint    `gorm:"primaryKey;column:uid"`
	UUID          string  `gorm:"type:varchar(32);uniqueIndex:idx_uuid;column:uuid"`
	Email         *string `gorm:"type:varchar(255);column:email"`
	Password      string  `gorm:"type:varchar(255);column:password"`
	RememberToken *string `gorm:"type:varchar(100);column:remember_token"`
	Verified      bool    `gorm:"type:tinyint(1);not null;default:0;column:verified"`
	TOTP          *string `gorm:"type:varchar(32);column:totp"`
	TwoFA         bool    `gorm:"type:tinyint(1);not null;default:0;column:2FA"`
	MBE           bool    `gorm:"type:tinyint(1);not null;default:0;column:mbe"`
	MojangUUID    *string `gorm:"type:varchar(32);column:mojang_uuid;uniqueIndex:uk_users_mojang_uuid"`
}

func (User) TableName() string {
	return "users"
}

// OAuth2AccessToken 映射 HRPAuth 的 oauth2_access_tokens 表。
type OAuth2AccessToken struct {
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	AccessToken string     `gorm:"type:varchar(255);column:access_token;uniqueIndex:uk_oauth2_access_tokens_access_token"`
	ClientID    string     `gorm:"type:varchar(100);column:client_id"`
	UserID      *string    `gorm:"type:varchar(32);column:user_id;index:idx_oauth2_access_tokens_user_id"`
	Scopes      string     `gorm:"type:text;column:scopes"`
	SubjectType string     `gorm:"type:enum('user','service');column:subject_type"`
	TargetUID   *uint      `gorm:"column:target_uid;index:idx_oauth2_access_tokens_target_uid"`
	TargetEmail *string    `gorm:"type:varchar(255);column:target_email;index:idx_oauth2_access_tokens_target_email"`
	ExpiresAt   time.Time  `gorm:"column:expires_at;index:idx_oauth2_access_tokens_expires_at"`
	RevokedAt   *time.Time `gorm:"column:revoked_at"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
}

func (OAuth2AccessToken) TableName() string {
	return "oauth2_access_tokens"
}

// OAuth2RefreshToken 映射 HRPAuth 的 oauth2_refresh_tokens 表。
type OAuth2RefreshToken struct {
	ID            uint       `gorm:"primaryKey;autoIncrement"`
	RefreshToken  string     `gorm:"type:varchar(255);column:refresh_token;uniqueIndex:uk_oauth2_refresh_tokens_refresh_token"`
	AccessTokenID uint       `gorm:"column:access_token_id;index:idx_oauth2_refresh_tokens_access_token_id"`
	ClientID      string     `gorm:"type:varchar(100);column:client_id"`
	UserID        string     `gorm:"type:varchar(32);column:user_id;index:idx_oauth2_refresh_tokens_user_id"`
	Scopes        string     `gorm:"type:text;column:scopes"`
	ExpiresAt     time.Time  `gorm:"column:expires_at;index:idx_oauth2_refresh_tokens_expires_at"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

func (OAuth2RefreshToken) TableName() string {
	return "oauth2_refresh_tokens"
}

// Profile 映射 HRPAuth 的 profiles 表。
type Profile struct {
	ID        string    `gorm:"primaryKey;type:varchar(32);column:id"`
	UserID    uint      `gorm:"column:user_id;index:idx_profile_user_id"`
	Name      string    `gorm:"type:varchar(255);column:name;uniqueIndex:uk_profile_name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Profile) TableName() string {
	return "profiles"
}

// TextureRecord 是接口层使用的统一纹理记录结构。
type TextureRecord struct {
	ID          uint
	Hash        string
	Type        string
	UID         uint
	Model       string
	Width       int
	Height      int
	FileName    string
	PreviewFile string
	Name        string
	Description string
	Tags        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TextureListSkinBase struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Hash        string    `gorm:"type:varchar(64);column:hash;uniqueIndex:uk_texture_list_uid_hash,priority:2"`
	UID         uint      `gorm:"column:uid;index:idx_texture_list_uid;uniqueIndex:uk_texture_list_uid_hash,priority:1"`
	Model       string    `gorm:"type:enum('default','slim');default:'default';column:model"`
	Width       int       `gorm:"not null;default:0;column:width"`
	Height      int       `gorm:"not null;default:0;column:height"`
	FileName    string    `gorm:"type:varchar(255);column:file_name"`
	PreviewFile string    `gorm:"type:varchar(255);column:previewfile"`
	Name        string    `gorm:"type:varchar(20);column:name"`
	Description string    `gorm:"type:text;column:description"`
	Tags        string    `gorm:"type:varchar(255);column:tags"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

type TextureListSkin struct {
	TextureListSkinBase
}

func (TextureListSkin) TableName() string {
	return "texture_list_skin"
}

type TextureListCape struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Hash        string    `gorm:"type:varchar(64);column:hash;uniqueIndex:uk_texture_list_uid_hash,priority:2"`
	UID         uint      `gorm:"column:uid;index:idx_texture_list_uid;uniqueIndex:uk_texture_list_uid_hash,priority:1"`
	Width       int       `gorm:"not null;default:0;column:width"`
	Height      int       `gorm:"not null;default:0;column:height"`
	FileName    string    `gorm:"type:varchar(255);column:file_name"`
	PreviewFile string    `gorm:"type:varchar(255);column:previewfile"`
	Name        string    `gorm:"type:varchar(20);column:name"`
	Description string    `gorm:"type:text;column:description"`
	Tags        string    `gorm:"type:varchar(255);column:tags"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (TextureListCape) TableName() string {
	return "texture_list_cape"
}
