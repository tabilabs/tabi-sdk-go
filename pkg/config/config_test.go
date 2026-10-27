package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigReturnsErrorForMissingPath(t *testing.T) {
	// ReadConfig must return an error instead of panicking on a malformed or
	// missing config path.
	err := ReadConfig(filepath.Join(t.TempDir(), "does-not-exist.toml"))
	if err == nil {
		t.Fatal("expected error for missing config file, got nil")
	}
}

func TestReadConfigParsesTemplate(t *testing.T) {
	// The shipped template must be a valid config with all required fields, so
	// copy-on-edit setups stay working.
	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "config", "template.toml"))
	if err != nil {
		t.Fatalf("cannot read template.toml: %v", err)
	}
	target := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(target, src, 0o600); err != nil {
		t.Fatal(err)
	}
	ClientConfig = Config{}
	if err := ReadConfig(target); err != nil {
		t.Fatalf("template.toml must parse cleanly, got: %v", err)
	}
	if ClientConfig.Chain.GrpcAddr == "" || ClientConfig.Chain.NodeAddr == "" {
		t.Fatal("template.toml must configure both grpc_addr and node_addr")
	}
	if ClientConfig.Chain.ChainId == "" {
		t.Fatal("template.toml must configure chain_id")
	}
}
