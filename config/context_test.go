package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestNewAppContext_OwnsIndependentRuntime(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := filepath.Join(dir1, "app2")
	os.MkdirAll(dir2, 0755)

	app1 := NewAppContext(nil, Config{DataDir: dir1}, viper.New())
	app2 := NewAppContext(nil, Config{DataDir: dir2}, viper.New())
	if app1 == nil || app2 == nil {
		t.Fatal("expected non-nil app contexts")
	}
	if app1.Pool() == app2.Pool() {
		t.Fatal("expected independent pool references, got shared")
	}
}

func TestAppContextClose_DoesNotPanic(t *testing.T) {
	t.Parallel()
	app := NewAppContext(nil, Config{DataDir: t.TempDir()}, viper.New())
	if err := app.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}