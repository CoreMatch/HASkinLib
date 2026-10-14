package models

import "time"

// AuthenticatedUser 是 HASkinLib 业务层消费的最小用户映射。
type AuthenticatedUser struct {
	UID   uint
	UUID  string
	Email string
}

// CoreUserIdentity 对齐主项目 /internal/verify-token 返回的数据形状。
type CoreUserIdentity struct {
	UID   uint   `json:"uid"`
	UUID  string `json:"uuid"`
	Email string `json:"email"`
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
