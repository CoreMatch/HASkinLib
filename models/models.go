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

// TextureList 映射 HASkinLib 自有表 texture_list。
type TextureList struct {
	ID          uint      `gorm:"primaryKey;autoIncrement;column:id"`
	Hash        string    `gorm:"type:varchar(64);column:hash;uniqueIndex:uk_texture_list_uid_hash_type,priority:2"`
	Type        string    `gorm:"type:enum('skin','cape');not null;column:type;uniqueIndex:uk_texture_list_uid_hash_type,priority:3"`
	UID         uint      `gorm:"column:uid;index:idx_texture_list_uid;uniqueIndex:uk_texture_list_uid_hash_type,priority:1"`
	Model       string    `gorm:"type:enum('default','slim');default:'default';column:model"`
	Width       int       `gorm:"not null;default:0;column:width"`
	Height      int       `gorm:"not null;default:0;column:height"`
	FileName    string    `gorm:"type:varchar(255);column:file_name"`
	Name        string    `gorm:"type:varchar(255);column:name"`
	Description string    `gorm:"type:text;column:description"`
	Tags        string    `gorm:"type:varchar(255);column:tags"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (TextureList) TableName() string {
	return "texture_list"
}
