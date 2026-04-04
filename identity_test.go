package hand

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewIdentity_SuppliedName(t *testing.T) {
	id := NewIdentity("inspector", "0.1.0", "keen-walnut")
	if id.Name != "keen-walnut" {
		t.Errorf("Name = %q, want %q", id.Name, "keen-walnut")
	}
	if id.Role != "inspector" {
		t.Errorf("Role = %q, want %q", id.Role, "inspector")
	}
	if id.Version != "0.1.0" {
		t.Errorf("Version = %q, want %q", id.Version, "0.1.0")
	}
}

func TestNewIdentity_GeneratedName(t *testing.T) {
	id := NewIdentity("maestro", "0.2.0", "")
	if id.Name == "" {
		t.Fatal("expected generated name, got empty")
	}
	parts := strings.Split(id.Name, "-")
	if len(parts) != 2 {
		t.Errorf("expected adjective-noun format, got %q", id.Name)
	}
	for _, p := range parts {
		if p != strings.ToLower(p) {
			t.Errorf("expected lowercase, got %q in %q", p, id.Name)
		}
	}
}

func TestNewIdentity_GeneratedNamesVary(t *testing.T) {
	names := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := NewIdentity("test", "0.0.0", "")
		names[id.Name] = true
	}
	if len(names) < 2 {
		t.Errorf("expected varied names, got %d unique out of 10", len(names))
	}
}

func TestBanner(t *testing.T) {
	id := Identity{Name: "keen-walnut", Role: "inspector", Version: "0.1.0"}
	var buf bytes.Buffer
	Banner(&buf, id)
	got := buf.String()
	if !strings.Contains(got, "inspector") {
		t.Errorf("banner missing role: %q", got)
	}
	if !strings.Contains(got, "0.1.0") {
		t.Errorf("banner missing version: %q", got)
	}
	if !strings.Contains(got, "keen-walnut") {
		t.Errorf("banner missing name: %q", got)
	}
}
