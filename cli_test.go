package hand

import (
	"testing"
	"time"
)

func TestParseCLI_AllFlags(t *testing.T) {
	var cli struct {
		CLI
	}
	err := ParseCLI("inspector", "test", &cli, []string{
		"--name", "keen-walnut",
		"--verbose",
		"--timeout", "30m",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.Name != "keen-walnut" {
		t.Errorf("Name = %q, want %q", cli.Name, "keen-walnut")
	}
	if !cli.Verbose {
		t.Error("Verbose = false, want true")
	}
	if cli.Timeout != 30*time.Minute {
		t.Errorf("Timeout = %v, want %v", cli.Timeout, 30*time.Minute)
	}
}

func TestParseCLI_Defaults(t *testing.T) {
	var cli struct {
		CLI
	}
	err := ParseCLI("inspector", "test", &cli, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.Name != "" {
		t.Errorf("Name = %q, want empty", cli.Name)
	}
	if cli.Verbose {
		t.Error("Verbose = true, want false")
	}
	if cli.Timeout != 15*time.Minute {
		t.Errorf("Timeout = %v, want %v", cli.Timeout, 15*time.Minute)
	}
}

func TestParseCLI_WithProjectDir(t *testing.T) {
	var cli struct {
		CLI
		ProjectDir string `kong:"arg,optional,help='Project directory'"`
	}
	err := ParseCLI("inspector", "test", &cli, []string{"/tmp/project"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.ProjectDir != "/tmp/project" {
		t.Errorf("ProjectDir = %q, want %q", cli.ProjectDir, "/tmp/project")
	}
}
