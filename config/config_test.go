package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseConfigAddr(t *testing.T) {
	cases := []struct {
		name string
		cfg  DatabaseConfig
		want string
	}{
		{"standalone port", DatabaseConfig{Host: "db.example", Port: 3307}, "db.example:3307"},
		{"legacy inline port without field", DatabaseConfig{Host: "db.example:3306"}, "db.example:3306"},
		{"bare host without port", DatabaseConfig{Host: "db.example"}, "db.example"},
		{"standalone port wins over inline", DatabaseConfig{Host: "db.example:3306", Port: 3307}, "db.example:3307"},
	}
	for _, tc := range cases {
		if got := tc.cfg.Addr(); got != tc.want {
			t.Errorf("%s: Addr() = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// loadFromYAML writes a minimal config file into a temp dir and runs Load
// against it, returning the parsed database section.
func loadFromYAML(t *testing.T, body string) DatabaseConfig {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), []byte(body), 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	Load()
	if AppConfig == nil {
		t.Fatal("Load did not populate AppConfig")
	}
	return AppConfig.Database
}

func TestLoadDatabasePort(t *testing.T) {
	withPort := loadFromYAML(t, `
version: "7"
database:
    host: db.example
    port: 3307
`)
	if withPort.Port != 3307 {
		t.Errorf("expected port 3307, got %d", withPort.Port)
	}
	if withPort.Addr() != "db.example:3307" {
		t.Errorf("Addr() = %q, want db.example:3307", withPort.Addr())
	}

	withoutPort := loadFromYAML(t, `
version: "7"
database:
    host: db.example:3306
`)
	if withoutPort.Port != 0 {
		t.Errorf("expected port 0 when absent, got %d", withoutPort.Port)
	}
	if withoutPort.Addr() != "db.example:3306" {
		t.Errorf("legacy inline host must keep working, got %q", withoutPort.Addr())
	}

	inlineAndField := loadFromYAML(t, `
version: "7"
database:
    host: db.example:3306
    port: 3307
`)
	if inlineAndField.Addr() != "db.example:3307" {
		t.Errorf("standalone port must win over inline, got %q", inlineAndField.Addr())
	}
}
