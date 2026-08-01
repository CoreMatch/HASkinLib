package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/database/migrations"
)

// ServiceName 本项目在共享 schema_migrations 表中的服务标识。
// schema_migrations 表由 HA 管理，本项目不创建/修改表结构，
// 只认 service='HASL' 这一行：不存在则创建，迁移后更新其 version。
const ServiceName = "HASL"

type StartupController struct{}

func NewStartupController() *StartupController {
	return &StartupController{}
}

// EnsureMigrations 执行自有数据库迁移（轻量执行器，替代 golang-migrate）。
// 步骤：
//  1. 检查 schema_migrations 中 service='HASL' 的行是否存在，不存在则创建（version=0）；
//  2. 按文件名顺序执行 embed 的自有迁移（*.up.sql，均为幂等 DDL，不动 users 表）；
//  3. 将 HASL 行的 version 更新为最新迁移版本。
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

	if err := sc.ensureServiceRow(db); err != nil {
		return err
	}

	if err := sc.runEmbeddedMigrations(db); err != nil {
		return err
	}

	return sc.updateServiceVersion(db)
}

// ensureServiceRow 检查 service='HASL' 的行是否存在；不存在则创建（version=0）。
// 只读写本服务自己的行，不触碰 schema_migrations 表结构及其它服务行。
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
