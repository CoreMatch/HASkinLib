package models

import (
	"time"
)

// User 映射 HRPAuth 的 users 表（只读复用，表由 HRPAuth 迁移创建）。
type User struct {
	UID           uint       `gorm:"primaryKey;column:uid"`
	UUID          string     `gorm:"type:varchar(32);column:uuid;index:idx_uuid"`
	Email         string     `gorm:"type:varchar(255);column:email"`
	Avatar        string     `gorm:"type:varchar(255);column:avatar"`
	Password      string     `gorm:"type:varchar(255);not null;column:password"`
	IP            string     `gorm:"type:varchar(255);column:ip"`
	Permission    int        `gorm:"default:0;column:permission"`
	LastSignAt    *time.Time `gorm:"column:last_sign_at"`
	RegisterAt    *time.Time `gorm:"column:register_at"`
	Verified      bool       `gorm:"type:tinyint(1);default:0;column:verified"`
	RememberToken string     `gorm:"type:varchar(100);column:remember_token"`
	Username      string     `gorm:"type:varchar(255);column:username"`
	RegIP         string     `gorm:"type:varchar(40);column:regip"`
	TOTP          string     `gorm:"type:varchar(32);column:totp"`
	CBH           bool       `gorm:"type:tinyint(1);not null;default:1;column:cbh"`
	MBE           bool       `gorm:"type:tinyint(1);not null;default:0;column:mbe"`
	MojangUUID    *string    `gorm:"type:varchar(32);column:mojang_uuid;uniqueIndex:uk_users_mojang_uuid"`
}

func (User) TableName() string {
	return "users"
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
