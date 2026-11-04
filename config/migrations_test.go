package config

import "testing"

func TestVersionMajor(t *testing.T) {
	cases := map[string]int{
		"1":   1,
		"4":   4,
		"6":   6,
		"6.1": 6,
		"":    0,
		"abc": 0,
	}

	for in, want := range cases {
		if got := VersionMajor(in); got != want {
			t.Errorf("VersionMajor(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestMigrateConfigUpToDate(t *testing.T) {
	cfg := map[string]any{"version": ConfigVersion}

	out, changed, err := MigrateConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatal("expected no migration for up-to-date config")
	}
	if out["version"] != ConfigVersion {
		t.Fatalf("expected version %q, got %v", ConfigVersion, out["version"])
	}
}

func TestMigrateConfigFutureVersion(t *testing.T) {
	cfg := map[string]any{"version": "9", "database": map[string]any{"db_name": "future"}}

	out, changed, err := MigrateConfig(cfg)
	if err != nil {
		t.Fatalf("future version must remain untouched, got error: %v", err)
	}
	if changed {
		t.Fatal("expected no migration for newer config")
	}
	if out["version"] != "9" {
		t.Fatalf("expected untouched version, got %v", out["version"])
	}
}

func TestMigrateConfigMissingVersion(t *testing.T) {
	_, _, err := MigrateConfig(map[string]any{"database": map[string]any{}})
	if err == nil {
		t.Fatal("expected error for config missing version field")
	}
}

func TestMigrateConfigUnknownOldVersion(t *testing.T) {
	_, _, err := MigrateConfig(map[string]any{"version": "1.5"})
	if err == nil {
		t.Fatal("expected error when no migration path exists")
	}
}

func TestMigrateConfigV1Chain(t *testing.T) {
	cfg := map[string]any{
		"version": "1",
		"server": map[string]any{
			"port": ":2701",
		},
		"database": map[string]any{
			"host":     "127.0.0.1",
			"db_name":  "hrpa",
			"user":     "hrpa",
			"password": "secret",
			"charset":  "utf8mb4",
		},
		"textures": map[string]any{
			"storage_dir": "./data/textures",
		},
	}

	out, changed, err := MigrateConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected chain migration to run")
	}
	if out["version"] != ConfigVersion {
		t.Fatalf("expected final version %q, got %v", ConfigVersion, out["version"])
	}

	server := out["server"].(map[string]any)
	if server["cors"] != "*" {
		t.Errorf("expected server.cors default, got %v", server["cors"])
	}

	textures := out["textures"].(map[string]any)
	if textures["preview_storage_dir"] != "./data/previews" {
		t.Errorf("expected preview storage dir default, got %v", textures["preview_storage_dir"])
	}
	if textures["max_upload_bytes"] != 2<<20 {
		t.Errorf("expected max_upload_bytes default, got %v", textures["max_upload_bytes"])
	}
	if textures["max_request_bytes"] != (2<<20)+(256<<10) {
		t.Errorf("expected max_request_bytes default, got %v", textures["max_request_bytes"])
	}
	if textures["rate_limit_per_minute"] != 5 {
		t.Errorf("expected rate_limit_per_minute default, got %v", textures["rate_limit_per_minute"])
	}
	if textures["rate_limit_window_seconds"] != 60 {
		t.Errorf("expected rate_limit_window_seconds default, got %v", textures["rate_limit_window_seconds"])
	}

	coreAPI := out["core_api"].(map[string]any)
	if coreAPI["base_url"] != "http://127.0.0.1:2778" {
		t.Errorf("expected core_api base_url default, got %v", coreAPI["base_url"])
	}

	database := out["database"].(map[string]any)
	if database["table_prefix"] != "hasl_" {
		t.Errorf("expected database.table_prefix default, got %v", database["table_prefix"])
	}
}

func TestMigrateConfigV5ToV6AddsTablePrefix(t *testing.T) {
	cfg := map[string]any{
		"version": "5",
		"database": map[string]any{
			"host":     "127.0.0.1",
			"db_name":  "hrpa",
			"user":     "hrpa",
			"password": "hrpa",
			"charset":  "utf8mb4",
		},
	}

	out, changed, err := MigrateConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected v5->v6 migration to run")
	}

	database := out["database"].(map[string]any)
	if database["table_prefix"] != "hasl_" {
		t.Fatalf("expected table_prefix hasl_, got %v", database["table_prefix"])
	}
}

func TestMigrateConfigV6ToV7AddsDatabasePort(t *testing.T) {
	cfg := map[string]any{
		"version":  "6",
		"database": map[string]any{"host": "db.example:3307"},
	}

	out, changed, err := MigrateConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected v6->v7 migration to run")
	}
	if out["version"] != "7" {
		t.Fatalf("expected version 7, got %v", out["version"])
	}

	database := out["database"].(map[string]any)
	if database["port"] != 0 {
		t.Errorf("expected default port 0, got %v", database["port"])
	}
	if database["host"] != "db.example:3307" {
		t.Errorf("legacy inline host must stay untouched, got %v", database["host"])
	}
}

func TestMigrateConfigV6ToV7KeepsExistingPort(t *testing.T) {
	cfg := map[string]any{
		"version":  "6",
		"database": map[string]any{"host": "db.example", "port": 3307},
	}

	out, _, err := MigrateConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	database := out["database"].(map[string]any)
	if database["port"] != 3307 {
		t.Errorf("existing port must be preserved, got %v", database["port"])
	}
}
