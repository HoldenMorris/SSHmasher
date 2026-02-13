package ssh

import (
	"os"
	"testing"

	"github.com/holden/sshmasher/internal/model"
)

func TestListHostsEmpty(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	hosts, err := ListHosts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hosts != nil {
		t.Fatalf("expected nil, got %v", hosts)
	}
}

func TestListHostsParses(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	configContent := `Host myserver
    HostName 192.168.1.100
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa

Host dev
    HostName dev.example.com
    User deploy
`
	if err := os.WriteFile(dir.ConfigPath(), []byte(configContent), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	hosts, err := ListHosts(dir)
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	if hosts[0].Alias != "myserver" {
		t.Fatalf("expected alias 'myserver', got '%s'", hosts[0].Alias)
	}
	if hosts[0].HostName != "192.168.1.100" {
		t.Fatalf("expected hostname '192.168.1.100', got '%s'", hosts[0].HostName)
	}
	if hosts[0].User != "admin" {
		t.Fatalf("expected user 'admin', got '%s'", hosts[0].User)
	}
	if hosts[0].Port != "2222" {
		t.Fatalf("expected port '2222', got '%s'", hosts[0].Port)
	}
	if hosts[0].IdentityFile != "~/.ssh/id_rsa" {
		t.Fatalf("expected identity file '~/.ssh/id_rsa', got '%s'", hosts[0].IdentityFile)
	}
}

func TestAddHost(t *testing.T) {
	dir := NewSSHDir(t.TempDir())

	host := model.HostEntry{
		Alias:    "newhost",
		HostName: "10.0.0.1",
		User:     "root",
		Port:     "22",
	}

	if err := AddHost(dir, host); err != nil {
		t.Fatalf("AddHost failed: %v", err)
	}

	hosts, err := ListHosts(dir)
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Alias != "newhost" {
		t.Fatalf("expected alias 'newhost', got '%s'", hosts[0].Alias)
	}
}

func TestGetHost(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	configContent := `Host target
    HostName 10.0.0.5
    User ops
`
	os.WriteFile(dir.ConfigPath(), []byte(configContent), 0600)

	host, err := GetHost(dir, "target")
	if err != nil {
		t.Fatalf("GetHost failed: %v", err)
	}
	if host.HostName != "10.0.0.5" {
		t.Fatalf("expected hostname '10.0.0.5', got '%s'", host.HostName)
	}
}

func TestGetHostNotFound(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	os.WriteFile(dir.ConfigPath(), []byte("Host other\n    HostName x\n"), 0600)

	_, err := GetHost(dir, "missing")
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestDeleteHost(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	configContent := `Host keep
    HostName keep.example.com

Host remove
    HostName remove.example.com
    User nobody
`
	os.WriteFile(dir.ConfigPath(), []byte(configContent), 0600)

	if err := DeleteHost(dir, "remove"); err != nil {
		t.Fatalf("DeleteHost failed: %v", err)
	}

	hosts, err := ListHosts(dir)
	if err != nil {
		t.Fatalf("ListHosts failed: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host after delete, got %d", len(hosts))
	}
	if hosts[0].Alias != "keep" {
		t.Fatalf("expected remaining host 'keep', got '%s'", hosts[0].Alias)
	}
}

func TestUpdateHost(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	configContent := `Host edit
    HostName old.example.com
    User olduser
`
	os.WriteFile(dir.ConfigPath(), []byte(configContent), 0600)

	updated := model.HostEntry{
		Alias:    "edit",
		HostName: "new.example.com",
		User:     "newuser",
		Port:     "2222",
	}

	if err := UpdateHost(dir, updated); err != nil {
		t.Fatalf("UpdateHost failed: %v", err)
	}

	host, err := GetHost(dir, "edit")
	if err != nil {
		t.Fatalf("GetHost after update failed: %v", err)
	}
	if host.HostName != "new.example.com" {
		t.Fatalf("expected hostname 'new.example.com', got '%s'", host.HostName)
	}
	if host.User != "newuser" {
		t.Fatalf("expected user 'newuser', got '%s'", host.User)
	}
}

func TestWriteConfig(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	content := "Host test\n    HostName test.com\n"

	if err := WriteConfig(dir, content); err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	data, err := os.ReadFile(dir.ConfigPath())
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(data) != content {
		t.Fatalf("content mismatch: got %q", string(data))
	}
}
