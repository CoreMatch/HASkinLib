package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Config 最小化配置，仅保留服务器运行所需字段。
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Textures TextureConfig
}

type ServerConfig struct {
	Port string
}

// DatabaseConfig 数据库连接配置，字段与 HRPAuth 保持一致。
type DatabaseConfig struct {
	Host     string
	DBName   string
	User     string
	Password string
	Charset  string
}

// TextureConfig 保存材质文件存放配置。
type TextureConfig struct {
	StorageDir             string
	PreviewStorageDir      string
	MaxUploadBytes         int64
	MaxRequestBytes        int64
	RateLimitPerMinute     int
	RateLimitWindowSeconds int
}

const ConfigFileName = "config.yaml"
const ConfigFileDir = "./"
const ConfigVersion = "3"

var AppConfig *Config

func Load() {
	configPath := filepath.Join(ConfigFileDir, ConfigFileName)

	// 默认值，config 文件缺失或未配置时兜底。
	// 数据库默认值沿用 HRPAuth，便于共用同一数据库。
	port := ":2701"
	database := DatabaseConfig{
		Host:     "127.0.0.1",
		DBName:   "hrpa",
		User:     "hrpa",
		Password: "hrpa",
		Charset:  "utf8mb4",
	}
	textures := TextureConfig{
		StorageDir:             "./data/textures",
		PreviewStorageDir:      "./data/previews",
		MaxUploadBytes:         2 << 20,
		MaxRequestBytes:        (2 << 20) + (256 << 10),
		RateLimitPerMinute:     5,
		RateLimitWindowSeconds: 60,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: failed to read config file %s: %v, using defaults", configPath, err)
	} else {
		var yamlConfig map[string]any
		if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}
		if server, ok := yamlConfig["server"].(map[string]any); ok {
			if p := getString(server, "port"); p != "" {
				port = p
			}
		}
		if db, ok := yamlConfig["database"].(map[string]any); ok {
			if v := getString(db, "host"); v != "" {
				database.Host = v
			}
			if v := getString(db, "db_name"); v != "" {
				database.DBName = v
			}
			if v := getString(db, "user"); v != "" {
				database.User = v
			}
			if v := getString(db, "password"); v != "" {
				database.Password = v
			}
			if v := getString(db, "charset"); v != "" {
				database.Charset = v
			}
		}
		if textureCfg, ok := yamlConfig["textures"].(map[string]any); ok {
			if v := getString(textureCfg, "storage_dir"); v != "" {
				textures.StorageDir = v
			}
			if v := getString(textureCfg, "preview_storage_dir"); v != "" {
				textures.PreviewStorageDir = v
			}
			if v := getInt64(textureCfg, "max_upload_bytes"); v > 0 {
				textures.MaxUploadBytes = v
			}
			if v := getInt64(textureCfg, "max_request_bytes"); v > 0 {
				textures.MaxRequestBytes = v
			}
			if v := getInt(textureCfg, "rate_limit_per_minute"); v > 0 {
				textures.RateLimitPerMinute = v
			}
			if v := getInt(textureCfg, "rate_limit_window_seconds"); v > 0 {
				textures.RateLimitWindowSeconds = v
			}
		}
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: port,
		},
		Database: database,
		Textures: textures,
	}

	log.Println("Configuration loaded successfully")
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}

	switch value := m[key].(type) {
	case int:
		return value
	case int8:
		return int(value)
	case int16:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case uint:
		return int(value)
	case uint8:
		return int(value)
	case uint16:
		return int(value)
	case uint32:
		return int(value)
	case uint64:
		return int(value)
	case float64:
		return int(value)
	case float32:
		return int(value)
	default:
		return 0
	}
}

func getInt64(m map[string]any, key string) int64 {
	if m == nil {
		return 0
	}

	switch value := m[key].(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)
	case float64:
		return int64(value)
	case float32:
		return int64(value)
	default:
		return 0
	}
}
