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

const ConfigFileName = "config.yaml"
const ConfigFileDir = "./"

var AppConfig *Config

func Load() {
	configPath := filepath.Join(ConfigFileDir, ConfigFileName)

	// 默认值，config 文件缺失或未配置时兜底。
	// 数据库默认值沿用 HRPAuth，便于共用同一数据库。
	port := ":8080"
	database := DatabaseConfig{
		Host:     "127.0.0.1",
		DBName:   "hrpa",
		User:     "hrpa",
		Password: "hrpa",
		Charset:  "utf8mb4",
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Warning: failed to read config file %s: %v, using defaults", configPath, err)
	} else {
		var yamlConfig map[string]interface{}
		if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
			log.Fatalf("Failed to parse config file: %v", err)
		}
		if server, ok := yamlConfig["server"].(map[string]interface{}); ok {
			if p := getString(server, "port"); p != "" {
				port = p
			}
		}
		if db, ok := yamlConfig["database"].(map[string]interface{}); ok {
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
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: port,
		},
		Database: database,
	}

	log.Println("Configuration loaded successfully")
}

func getString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}
