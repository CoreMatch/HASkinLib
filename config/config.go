package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// Config 最小化配置，仅保留服务器运行所需字段。
type Config struct {
	Server ServerConfig
}

type ServerConfig struct {
	Port string
}

const ConfigFileName = "config.yaml"
const ConfigFileDir = "./"

var AppConfig *Config

func Load() {
	configPath := filepath.Join(ConfigFileDir, ConfigFileName)

	// 默认端口，config 文件缺失或未配置时兜底。
	port := ":8080"

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
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: port,
		},
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
