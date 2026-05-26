package cmd

import (
	"testing"
)

func TestInitApp_CreatesUsableAppContextWithoutGlobalPool(t *testing.T) {
	app = nil
	initApp()
	if app == nil {
		t.Fatal("expected app context")
	}
	if app.System() == nil {
		t.Fatal("expected runtime system on app context")
	}
}
