package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// ConfigMigration 描述一条配置版本迁移步骤。
type ConfigMigration struct {
	FromVersion string
	ToVersion   string
	Migrate     func(cfg map[string]any) error
}

func configMigrations() []ConfigMigration {
	return []ConfigMigration{
		{FromVersion: "1", ToVersion: "2", Migrate: migrateV1ToV2},
		{FromVersion: "2", ToVersion: "3", Migrate: migrateV2ToV3},
		{FromVersion: "3", ToVersion: "4", Migrate: migrateV3ToV4},
		{FromVersion: "4", ToVersion: "5", Migrate: migrateV4ToV5},
		{FromVersion: "5", ToVersion: "6", Migrate: migrateV5ToV6},
		{FromVersion: "6", ToVersion: "7", Migrate: migrateV6ToV7},
	}
}

// MigrateConfig 将配置逐步升级到当前 ConfigVersion。
func MigrateConfig(cfg map[string]any) (map[string]any, bool, error) {
	current, _ := cfg["version"].(string)
	if current == "" {
		return nil, false, fmt.Errorf("config file is missing the version field; add version: %q or restore a backup", ConfigVersion)
	}

	if current == ConfigVersion {
		return cfg, false, nil
	}

	if VersionMajor(current) > VersionMajor(ConfigVersion) {
		log.Printf("Warning: config file version %q is newer than supported version %q, continuing without migration", current, ConfigVersion)
		return cfg, false, nil
	}

	changed := false
	steps := 0
	for current != ConfigVersion {
		step, found := findMigration(current)
		if !found {
			return nil, changed, fmt.Errorf("no migration path from config version %q to %q; restore a backup or upgrade the program", current, ConfigVersion)
		}
		if err := step.Migrate(cfg); err != nil {
			return nil, changed, fmt.Errorf("migration %s->%s failed: %w", step.FromVersion, step.ToVersion, err)
		}

		next, _ := cfg["version"].(string)
		if next == "" || next == current {
			return nil, changed, fmt.Errorf("migration %s->%s did not advance the version field (got %q)", step.FromVersion, step.ToVersion, next)
		}

		current = next
		changed = true
		steps++
		if steps > len(configMigrations())+1 {
			return nil, changed, fmt.Errorf("config migration did not converge")
		}
	}

	return cfg, changed, nil
}

func findMigration(from string) (*ConfigMigration, bool) {
	for i := range configMigrations() {
		if configMigrations()[i].FromVersion == from {
			m := configMigrations()[i]
			return &m, true
		}
	}
	return nil, false
}

// VersionMajor 取版本号前导整数，例如 "6.1" -> 6。
func VersionMajor(v string) int {
	v = strings.TrimSpace(v)
	i := 0
	for i < len(v) && v[i] >= '0' && v[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0
	}

	n, err := strconv.Atoi(v[:i])
	if err != nil {
		return 0
	}
	return n
}

// BackupConfigFile 在迁移前备份原始配置。
func BackupConfigFile(path, version string) error {
	backupPath := fmt.Sprintf("%s.bak.%s", path, version)
	if _, err := os.Stat(backupPath); err == nil {
		log.Printf("Config backup %s already exists, keeping it", backupPath)
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file for backup: %w", err)
	}
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config backup %s: %w", backupPath, err)
	}

	log.Printf("Config backed up to %s", backupPath)
	return nil
}

func migrateV1ToV2(cfg map[string]any) error {
	server := ensureSection(cfg, "server")
	if _, exists := server["cors"]; !exists {
		server["cors"] = "*"
	}
	cfg["version"] = "2"
	return nil
}

func migrateV2ToV3(cfg map[string]any) error {
	textures := ensureSection(cfg, "textures")
	if _, exists := textures["preview_storage_dir"]; !exists {
		textures["preview_storage_dir"] = "./data/previews"
	}
	cfg["version"] = "3"
	return nil
}

func migrateV3ToV4(cfg map[string]any) error {
	textures := ensureSection(cfg, "textures")
	if _, exists := textures["max_upload_bytes"]; !exists {
		textures["max_upload_bytes"] = 2 << 20
	}
	if _, exists := textures["max_request_bytes"]; !exists {
		textures["max_request_bytes"] = (2 << 20) + (256 << 10)
	}
	if _, exists := textures["rate_limit_per_minute"]; !exists {
		textures["rate_limit_per_minute"] = 5
	}
	if _, exists := textures["rate_limit_window_seconds"]; !exists {
		textures["rate_limit_window_seconds"] = 60
	}
	cfg["version"] = "4"
	return nil
}

func migrateV4ToV5(cfg map[string]any) error {
	coreAPI := ensureSection(cfg, "core_api")
	if _, exists := coreAPI["base_url"]; !exists {
		coreAPI["base_url"] = "http://127.0.0.1:2778"
	}
	if _, exists := coreAPI["client_id"]; !exists {
		coreAPI["client_id"] = ""
	}
	if _, exists := coreAPI["client_secret"]; !exists {
		coreAPI["client_secret"] = ""
	}
	cfg["version"] = "5"
	return nil
}

func migrateV5ToV6(cfg map[string]any) error {
	database := ensureSection(cfg, "database")
	if _, exists := database["table_prefix"]; !exists {
		database["table_prefix"] = "hasl_"
	}
	cfg["version"] = "6"
	return nil
}

// migrateV6ToV7 adds the optional database port field. The default 0 keeps
// the legacy behavior where any port is inlined into the host value
// ("ip:port"); DatabaseConfig.Addr prefers the standalone field when set.
func migrateV6ToV7(cfg map[string]any) error {
	database := ensureSection(cfg, "database")
	if _, exists := database["port"]; !exists {
		database["port"] = 0
	}
	cfg["version"] = "7"
	return nil
}

func ensureSection(cfg map[string]any, key string) map[string]any {
	section, _ := cfg[key].(map[string]any)
	if section == nil {
		section = map[string]any{}
		cfg[key] = section
	}
	return section
}
