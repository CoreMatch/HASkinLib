package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goccy/go-yaml"
	"github.com/golang-migrate/migrate/v4"
	mysqldriver "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/database/migrations"
)

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
			"cors": "*",
		},
		"database": map[string]any{
			"host":         "127.0.0.1",
			"db_name":      "hrpa",
			"user":         "hrpa",
			"password":     "hrpa",
			"charset":      "utf8mb4",
			"table_prefix": "hasl_",
		},
		"core_api": map[string]any{
			"base_url":      "http://127.0.0.1:2778",
			"client_id":     "",
			"client_secret": "",
		},
		"textures": map[string]any{
			"storage_dir":               "./data/textures",
			"preview_storage_dir":       "./data/previews",
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

	log.Printf("Config file version mismatch: current=%q, expected=%q, migrating...", currentVersion, config.ConfigVersion)

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
	for _, key := range []string{"storage_dir", "preview_storage_dir"} {
		storageDir, _ := textures[key].(string)
		if storageDir == "" {
			continue
		}

		if err := os.MkdirAll(storageDir, 0755); err != nil {
			return fmt.Errorf("failed to create texture storage directory %s: %v", storageDir, err)
		}
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

// EnsureMigrations 使用 golang-migrate 执行所有待处理迁移。
// 方案 1 明确要求独立数据库，因此这里不再兼容共享时期带 service 列的
// schema_migrations 表；一旦检测到旧结构，会直接报错并提示切换到独立库。
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

	if err := sc.ensureStandardMigrationsTable(db, cfg.DBName); err != nil {
		return err
	}

	driver, err := mysqldriver.WithInstance(db, &mysqldriver.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %v", err)
	}

	migrationFS, err := buildMigrationFS(config.AppConfig.Database.TablePrefix)
	if err != nil {
		return fmt.Errorf("failed to prepare prefixed migration source: %v", err)
	}

	src, err := iofs.New(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source from embedded files: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, cfg.DBName, driver)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %v", err)
	}
	defer func() {
		_, closeErr := m.Close()
		if closeErr != nil {
			log.Printf("warning: migration close error: %v", closeErr)
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		var dirtyErr migrate.ErrDirty
		if errors.As(err, &dirtyErr) {
			return fmt.Errorf("failed to run migrations: database is dirty at version %d; fix the migration state and force version before restarting", dirtyErr.Version)
		}
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Printf("Database migration completed with no applied versions")
			return nil
		}
		return fmt.Errorf("failed to read database migration version: %v", err)
	}

	log.Printf("Database migration completed at version %d (dirty: %t)", version, dirty)
	return nil
}

func buildMigrationFS(tablePrefix string) (fs.FS, error) {
	entries, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return nil, err
	}

	migrationFS := fstest.MapFS{}
	replacer := strings.NewReplacer(
		"`texture_list_skin`", "`"+tablePrefix+"texture_list_skin`",
		"`texture_list_cape`", "`"+tablePrefix+"texture_list_cape`",
	)

	for _, name := range entries {
		data, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return nil, err
		}

		migrationFS[name] = &fstest.MapFile{
			Data: []byte(replacer.Replace(string(data))),
			Mode: 0644,
		}
	}

	return migrationFS, nil
}

// ensureStandardMigrationsTable 确保独立数据库使用标准 schema_migrations 表。
// 如果检测到旧共享库遗留的 service 列，直接报错，避免在错误的数据库上继续运行。
func (sc *StartupController) ensureStandardMigrationsTable(db *sql.DB, dbName string) error {
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

	if tableCount == 0 {
		if _, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version BIGINT NOT NULL,
				dirty BOOLEAN NOT NULL,
				PRIMARY KEY (version)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
		`); err != nil {
			return fmt.Errorf("failed to create schema_migrations table: %v", err)
		}
		log.Printf("Created standard schema_migrations table")
		return nil
	}

	var serviceColumnCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = 'schema_migrations'
		  AND COLUMN_NAME = 'service'
	`, dbName).Scan(&serviceColumnCount)
	if err != nil {
		return fmt.Errorf("failed to inspect schema_migrations columns: %v", err)
	}

	if serviceColumnCount > 0 {
		return fmt.Errorf("legacy shared schema_migrations table detected in database %q; option 1 requires a dedicated HASkinLib database or a fresh standard schema_migrations table without the service column", dbName)
	}

	return nil
}
