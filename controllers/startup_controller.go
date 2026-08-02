package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goccy/go-yaml"
	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/database/migrations"
)

// ServiceName 本项目在共享 schema_migrations 表中的服务标识。
// 若 schema_migrations 不存在，本项目会按当前所需最小结构创建它；
// 之后只认 service='HASL' 这一行：不存在则创建，迁移后更新其 version。
const ServiceName = "HASL"

type StartupController struct{}

func NewStartupController() *StartupController {
	return &StartupController{}
}

func (sc *StartupController) InitializeConfig() error {
	configPath := filepath.Join(config.ConfigFileDir, config.ConfigFileName)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found at %s, creating default config...", configPath)
		return sc.createDefaultConfig(configPath)
	} else if err != nil {
		return fmt.Errorf("failed to inspect config file: %v", err)
	}

	log.Printf("Config file found at %s", configPath)
	return sc.checkAndMigrateConfig(configPath)
}

func (sc *StartupController) buildDefaultConfig() map[string]any {
	return map[string]any{
		"version": config.ConfigVersion,
		"server": map[string]any{
			"port": ":2701",
		},
		"database": map[string]any{
			"host":     "127.0.0.1",
			"db_name":  "hrpa",
			"user":     "hrpa",
			"password": "hrpa",
			"charset":  "utf8mb4",
		},
		"textures": map[string]any{
			"storage_dir":               "./data/textures",
			"max_upload_bytes":          2 << 20,
			"max_request_bytes":         (2 << 20) + (256 << 10),
			"rate_limit_per_minute":     5,
			"rate_limit_window_seconds": 60,
		},
	}
}

func (sc *StartupController) createDefaultConfig(path string) error {
	cfgDir := filepath.Dir(path)
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	defaultConfig := sc.buildDefaultConfig()
	if err := sc.ensureTextureStorageDir(defaultConfig); err != nil {
		return err
	}

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	log.Printf("Default config file created at %s", path)
	return nil
}

func (sc *StartupController) checkAndMigrateConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var currentConfig map[string]any
	if unmarshalErr := yaml.Unmarshal(data, &currentConfig); unmarshalErr != nil {
		return fmt.Errorf("failed to parse config file: %v", unmarshalErr)
	}

	currentVersion, _ := currentConfig["version"].(string)
	if currentVersion == config.ConfigVersion {
		log.Printf("Config file version %s is up-to-date", currentVersion)
		return sc.ensureTextureStorageDir(currentConfig)
	}

	if currentVersion == "" {
		log.Printf("Config file is missing version field, migrating to version %s", config.ConfigVersion)
	} else {
		log.Printf("Config file version mismatch: current=%q, expected=%q, migrating...", currentVersion, config.ConfigVersion)
	}

	defaultConfig := sc.buildDefaultConfig()
	mergedConfig := mergeConfigMaps(defaultConfig, currentConfig)
	if ensureDirErr := sc.ensureTextureStorageDir(mergedConfig); ensureDirErr != nil {
		return ensureDirErr
	}

	data, err = yaml.Marshal(mergedConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal migrated config: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write migrated config file: %v", err)
	}

	log.Printf("Config file migrated to version %s at %s", config.ConfigVersion, path)
	return nil
}

func (sc *StartupController) ensureTextureStorageDir(cfg map[string]any) error {
	textures, _ := cfg["textures"].(map[string]any)
	storageDir, _ := textures["storage_dir"].(string)
	if storageDir == "" {
		return nil
	}

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create texture storage directory %s: %v", storageDir, err)
	}
	return nil
}

func mergeConfigMaps(defaultCfg, currentCfg map[string]any) map[string]any {
	result := make(map[string]any, len(defaultCfg))
	for key, defaultVal := range defaultCfg {
		currentVal, exists := currentCfg[key]
		if !exists {
			result[key] = deepCopyValue(defaultVal)
			continue
		}

		defaultMap, defaultIsMap := defaultVal.(map[string]any)
		currentMap, currentIsMap := currentVal.(map[string]any)
		if defaultIsMap && currentIsMap {
			result[key] = mergeConfigMaps(defaultMap, currentMap)
			continue
		}

		result[key] = currentVal
	}
	return result
}

func deepCopyValue(v any) any {
	data, err := yaml.Marshal(v)
	if err != nil {
		return v
	}

	var copy any
	if err := yaml.Unmarshal(data, &copy); err != nil {
		return v
	}
	return copy
}

// EnsureMigrations 执行自有数据库迁移（轻量执行器，替代 golang-migrate）。
// 步骤：
//  1. 确保 schema_migrations 表存在；
//  2. 确保 schema_migrations.dirty 列存在且默认值为 0（仅兼容，不参与逻辑）；
//  3. 检查 schema_migrations 中 service='HASL' 的行是否存在，不存在则创建（version=0）；
//  4. 按文件名顺序执行 embed 的自有迁移（*.up.sql，均为幂等 DDL，不动 users 表）；
//  5. 将 HASL 行的 version 更新为最新迁移版本。
func (sc *StartupController) EnsureMigrations() error {
	cfg := config.AppConfig.Database
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.DBName, cfg.Charset,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %v", err)
	}
	defer db.Close()

	if err := sc.ensureMigrationsTable(db, cfg.DBName); err != nil {
		return err
	}

	if err := sc.ensureDirtyColumnDefault(db, cfg.DBName); err != nil {
		return err
	}

	if err := sc.ensureServiceRow(db); err != nil {
		return err
	}

	if err := sc.runEmbeddedMigrations(db); err != nil {
		return err
	}

	return sc.updateServiceVersion(db)
}

// ensureMigrationsTable 确保共享 schema_migrations 表存在。
// 表结构只包含当前项目所需的最小字段：service、version、dirty。
func (sc *StartupController) ensureMigrationsTable(db *sql.DB, dbName string) error {
	var tableCount int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = 'schema_migrations'
	`, dbName).Scan(&tableCount)
	if err != nil {
		return fmt.Errorf("failed to inspect schema_migrations table: %v", err)
	}

	if tableCount > 0 {
		return nil
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			service VARCHAR(64) NOT NULL,
			version INT NOT NULL DEFAULT 0,
			dirty TINYINT(1) NOT NULL DEFAULT 0,
			PRIMARY KEY (service)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %v", err)
	}

	log.Printf("Created schema_migrations table")
	return nil
}

// ensureDirtyColumnDefault 确保 schema_migrations 的 dirty 列存在且默认值为 0。
// 该列仅用于兼容外部迁移表结构，本项目不读取也不依赖它。
func (sc *StartupController) ensureDirtyColumnDefault(db *sql.DB, dbName string) error {
	var columnCount int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = 'schema_migrations'
		  AND COLUMN_NAME = 'dirty'
	`, dbName).Scan(&columnCount)
	if err != nil {
		return fmt.Errorf("failed to inspect schema_migrations.dirty column: %v", err)
	}

	if columnCount == 0 {
		if _, execErr := db.Exec("ALTER TABLE schema_migrations ADD COLUMN dirty TINYINT(1) NOT NULL DEFAULT 0"); execErr != nil {
			return fmt.Errorf("failed to add schema_migrations.dirty column: %v", execErr)
		}
		log.Printf("Added schema_migrations.dirty column with default 0")
		return nil
	}

	var columnDefault sql.NullString
	err = db.QueryRow(`
		SELECT COLUMN_DEFAULT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = 'schema_migrations'
		  AND COLUMN_NAME = 'dirty'
	`, dbName).Scan(&columnDefault)
	if err != nil {
		return fmt.Errorf("failed to inspect schema_migrations.dirty default: %v", err)
	}

	if !columnDefault.Valid || columnDefault.String != "0" {
		if _, err := db.Exec("ALTER TABLE schema_migrations MODIFY COLUMN dirty TINYINT(1) NOT NULL DEFAULT 0"); err != nil {
			return fmt.Errorf("failed to set default value for schema_migrations.dirty: %v", err)
		}
		log.Printf("Updated schema_migrations.dirty default to 0")
	}

	return nil
}

// ensureServiceRow 检查 service='HASL' 的行是否存在；不存在则创建（version=0）。
// 只读写本服务自己的行，不触碰其它服务行。
func (sc *StartupController) ensureServiceRow(db *sql.DB) error {
	var version int
	err := db.QueryRow("SELECT version FROM schema_migrations WHERE service = ?", ServiceName).Scan(&version)
	if err == nil {
		log.Printf("schema_migrations row for service %s exists (version %d)", ServiceName, version)
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to query schema_migrations for service %s: %v", ServiceName, err)
	}

	if _, err := db.Exec("INSERT INTO schema_migrations (service, version) VALUES (?, 0)", ServiceName); err != nil {
		return fmt.Errorf("failed to create schema_migrations row for service %s: %v", ServiceName, err)
	}
	log.Printf("Created schema_migrations row for service %s", ServiceName)
	return nil
}

// runEmbeddedMigrations 按文件名顺序执行 embed 的自有迁移（仅 *.up.sql）。
// 迁移 SQL 均为幂等 DDL（CREATE TABLE IF NOT EXISTS），重复执行安全。
func (sc *StartupController) runEmbeddedMigrations(db *sql.DB) error {
	entries, err := fs.Glob(migrations.FS, "*.up.sql")
	if err != nil {
		return fmt.Errorf("failed to list migration files: %v", err)
	}
	sort.Strings(entries)

	for _, name := range entries {
		data, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", name, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("failed to run migration %s: %v", name, err)
		}
		log.Printf("Applied migration %s", name)
	}
	return nil
}

// updateServiceVersion 将 HASL 行的 version 更新为最新迁移版本。
func (sc *StartupController) updateServiceVersion(db *sql.DB) error {
	latest := latestMigrationVersion()
	if _, err := db.Exec("UPDATE schema_migrations SET version = ? WHERE service = ?", latest, ServiceName); err != nil {
		return fmt.Errorf("failed to update schema_migrations version for service %s: %v", ServiceName, err)
	}
	log.Printf("Updated schema_migrations version for service %s to %d", ServiceName, latest)
	return nil
}

// latestMigrationVersion 从 embed 的迁移文件名（如 000002_texture_list.up.sql）解析出最大版本号。
func latestMigrationVersion() int {
	entries, _ := fs.Glob(migrations.FS, "*.up.sql")
	max := 0
	for _, name := range entries {
		v := migrationVersion(name)
		if v > max {
			max = v
		}
	}
	return max
}

func migrationVersion(name string) int {
	base := strings.TrimSuffix(name, ".up.sql")
	idx := strings.Index(base, "_")
	if idx <= 0 {
		return 0
	}
	v, err := strconv.Atoi(base[:idx])
	if err != nil {
		return 0
	}
	return v
}
