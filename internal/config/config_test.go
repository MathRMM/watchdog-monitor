package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/mathrmm/watchdog-monitor/internal/config"
)

func writeTempTOML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "watchdog-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// dado arquivo watchdog.toml válido com nats_url,
// quando Load() é chamado, então retorna *Config com campo populado sem erro.
func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempTOML(t, `nats_url = "nats://localhost:4222"`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.NatsURL != "nats://localhost:4222" {
		t.Errorf("expected NatsURL = 'nats://localhost:4222', got: %q", cfg.NatsURL)
	}
}

// dado arquivo inexistente, quando Load() é chamado,
// então retorna erro contendo o caminho do arquivo.
func TestLoad_FileNotFound(t *testing.T) {
	path := "/nonexistent/path/watchdog.toml"

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("expected error to contain file path %q, got: %v", path, err)
	}
}

// dado arquivo com TOML malformado, quando Load() é chamado,
// então retorna erro descritivo (não panic).
func TestLoad_MalformedTOML(t *testing.T) {
	path := writeTempTOML(t, `nats_url = [invalid toml`)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

// dado arquivo sem campo nats_url, quando Load() é chamado,
// então retorna erro de campo obrigatório ausente.
func TestLoad_MissingNatsURL(t *testing.T) {
	path := writeTempTOML(t, `log_path = "watchdog.log"`)

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing nats_url, got nil")
	}
	if !strings.Contains(err.Error(), "nats_url") {
		t.Errorf("expected error to mention 'nats_url', got: %v", err)
	}
}

// arquivo watchdog.toml com campos extras desconhecidos deve ser aceito
// sem erro (forward compatibility).
func TestLoad_ExtraFieldsIgnored(t *testing.T) {
	path := writeTempTOML(t, `
nats_url       = "nats://localhost:4222"
unknown_field  = "should be ignored"
another_future = 42
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("expected no error for extra fields, got: %v", err)
	}
	if cfg.NatsURL != "nats://localhost:4222" {
		t.Errorf("expected NatsURL populated, got: %q", cfg.NatsURL)
	}
}
