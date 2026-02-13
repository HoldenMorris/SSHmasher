package ssh

import (
	"os"
	"testing"
)

func TestListKnownHostsEmpty(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	entries, err := ListKnownHosts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Fatalf("expected nil, got %v", entries)
	}
}

func TestListKnownHostsParses(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	// Use a real-ish known_hosts format
	content := `github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7 comment
`
	os.WriteFile(dir.KnownHostsPath(), []byte(content), 0644)

	entries, err := ListKnownHosts(dir)
	if err != nil {
		t.Fatalf("ListKnownHosts failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Hosts != "github.com" {
		t.Fatalf("expected host 'github.com', got '%s'", entries[0].Hosts)
	}
	if entries[0].KeyType != "ssh-ed25519" {
		t.Fatalf("expected key type 'ssh-ed25519', got '%s'", entries[0].KeyType)
	}
	if entries[0].Line != 1 {
		t.Fatalf("expected line 1, got %d", entries[0].Line)
	}
}

func TestListKnownHostsSkipsComments(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	content := `# This is a comment
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl

# Another comment
`
	os.WriteFile(dir.KnownHostsPath(), []byte(content), 0644)

	entries, err := ListKnownHosts(dir)
	if err != nil {
		t.Fatalf("ListKnownHosts failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestFilterKnownHosts(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	content := `github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
`
	os.WriteFile(dir.KnownHostsPath(), []byte(content), 0644)

	entries, _ := ListKnownHosts(dir)
	filtered := FilterKnownHosts(entries, "git")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered entries, got %d", len(filtered))
	}

	filtered = FilterKnownHosts(entries, "192.168")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered entry, got %d", len(filtered))
	}
}

func TestRemoveKnownHost(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	content := `github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
192.168.1.1 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
`
	os.WriteFile(dir.KnownHostsPath(), []byte(content), 0644)

	if err := RemoveKnownHost(dir, 2); err != nil {
		t.Fatalf("RemoveKnownHost failed: %v", err)
	}

	entries, _ := ListKnownHosts(dir)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries after removal, got %d", len(entries))
	}
	// First entry should still be github.com
	if entries[0].Hosts != "github.com" {
		t.Fatalf("expected first entry 'github.com', got '%s'", entries[0].Hosts)
	}
}

func TestRemoveKnownHostOutOfRange(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	os.WriteFile(dir.KnownHostsPath(), []byte("host ssh-rsa key\n"), 0644)

	if err := RemoveKnownHost(dir, 5); err == nil {
		t.Fatal("expected error for out of range line")
	}
}

func TestWriteKnownHosts(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	content := "host1 ssh-rsa key1\nhost2 ssh-ed25519 key2\n"

	if err := WriteKnownHosts(dir, content); err != nil {
		t.Fatalf("WriteKnownHosts failed: %v", err)
	}

	data, err := os.ReadFile(dir.KnownHostsPath())
	if err != nil {
		t.Fatalf("read known_hosts: %v", err)
	}
	if string(data) != content {
		t.Fatalf("content mismatch")
	}
}

func TestHashedKnownHost(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	content := `|1|abc123|def456 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
`
	os.WriteFile(dir.KnownHostsPath(), []byte(content), 0644)

	entries, err := ListKnownHosts(dir)
	if err != nil {
		t.Fatalf("ListKnownHosts failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].IsHashed {
		t.Fatal("expected IsHashed=true")
	}
}
