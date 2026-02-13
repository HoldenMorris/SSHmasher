package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/holden/sshmasher/internal/model"
)

func TestListKeysEmpty(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	keys, err := ListKeys(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys, got %d", len(keys))
	}
}

func TestListKeysNonexistentDir(t *testing.T) {
	dir := NewSSHDir(filepath.Join(t.TempDir(), "nonexistent"))
	keys, err := ListKeys(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if keys != nil {
		t.Fatalf("expected nil, got %v", keys)
	}
}

func TestGenerateAndListKeys(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	req := model.KeyGenRequest{
		Name:       "test_key",
		Type:       "ed25519",
		Comment:    "test@test",
		Passphrase: "",
	}

	if err := GenerateKey(dir, req); err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	// Verify files exist
	if !fileExists(dir.Path("test_key")) {
		t.Fatal("private key not created")
	}
	if !fileExists(dir.Path("test_key.pub")) {
		t.Fatal("public key not created")
	}

	// Verify permissions
	info, _ := os.Stat(dir.Path("test_key"))
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected private key perm 0600, got %o", info.Mode().Perm())
	}

	// List keys
	keys, err := ListKeys(dir)
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0].Name != "test_key" {
		t.Fatalf("expected name 'test_key', got '%s'", keys[0].Name)
	}
	if keys[0].Type != "ssh-ed25519" {
		t.Fatalf("expected type 'ssh-ed25519', got '%s'", keys[0].Type)
	}
	if keys[0].Fingerprint == "" {
		t.Fatal("expected non-empty fingerprint")
	}
	if !keys[0].HasPrivate {
		t.Fatal("expected HasPrivate=true")
	}
}

func TestGenerateKeyDuplicate(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	req := model.KeyGenRequest{Name: "dup_key", Type: "ed25519", Passphrase: ""}
	if err := GenerateKey(dir, req); err != nil {
		t.Fatalf("first GenerateKey failed: %v", err)
	}

	err := GenerateKey(dir, req)
	if err == nil {
		t.Fatal("expected error on duplicate key generation")
	}
}

func TestGetKey(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	req := model.KeyGenRequest{Name: "get_test", Type: "ed25519", Passphrase: "", Comment: "hello"}
	if err := GenerateKey(dir, req); err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	key, err := GetKey(dir, "get_test")
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}
	if key.Name != "get_test" {
		t.Fatalf("expected name 'get_test', got '%s'", key.Name)
	}
	if key.Comment != "hello" {
		t.Fatalf("expected comment 'hello', got '%s'", key.Comment)
	}
}

func TestGetKeyNotFound(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	_, err := GetKey(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
}

func TestDeleteKey(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	req := model.KeyGenRequest{Name: "del_test", Type: "ed25519", Passphrase: ""}
	if err := GenerateKey(dir, req); err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if err := DeleteKey(dir, "del_test"); err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	if fileExists(dir.Path("del_test")) {
		t.Fatal("private key still exists after delete")
	}
	if fileExists(dir.Path("del_test.pub")) {
		t.Fatal("public key still exists after delete")
	}
}

func TestDeleteKeyNotFound(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	err := DeleteKey(dir, "nonexistent")
	if err == nil {
		t.Fatal("expected error deleting nonexistent key")
	}
}

func TestGenerateRSAKey(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	req := model.KeyGenRequest{
		Name:       "rsa_test",
		Type:       "rsa",
		Bits:       2048,
		Comment:    "rsa@test",
		Passphrase: "",
	}

	if err := GenerateKey(dir, req); err != nil {
		t.Fatalf("GenerateKey RSA failed: %v", err)
	}

	key, err := GetKey(dir, "rsa_test")
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}
	if key.Type != "ssh-rsa" {
		t.Fatalf("expected type 'ssh-rsa', got '%s'", key.Type)
	}
}
