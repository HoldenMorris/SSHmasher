package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListBackupsEmpty(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	backups, err := ListBackups(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backups != nil {
		t.Fatalf("expected nil, got %v", backups)
	}
}

func TestCreateAndListBackup(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)

	dir := NewSSHDir(sshDir)

	// Create some test files in the SSH dir
	os.WriteFile(dir.Path("config"), []byte("Host test\n    HostName test.com\n"), 0600)
	os.WriteFile(dir.Path("known_hosts"), []byte("test ssh-rsa key\n"), 0644)

	if err := CreateBackup(dir); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	backups, err := ListBackups(dir)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(backups))
	}
	if backups[0].Size == 0 {
		t.Fatal("expected non-zero backup size")
	}
}

func TestBackupAndRestore(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)

	dir := NewSSHDir(sshDir)

	// Create original files
	originalConfig := "Host original\n    HostName original.com\n"
	os.WriteFile(dir.Path("config"), []byte(originalConfig), 0600)

	// Create backup
	if err := CreateBackup(dir); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Modify the SSH dir
	os.WriteFile(dir.Path("config"), []byte("Host modified\n    HostName modified.com\n"), 0600)

	// Restore
	backups, _ := ListBackups(dir)
	if len(backups) == 0 {
		t.Fatal("no backups found")
	}

	if err := RestoreBackup(dir, backups[0].Filename); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	// Verify original content restored
	data, err := os.ReadFile(dir.Path("config"))
	if err != nil {
		t.Fatalf("read restored config: %v", err)
	}
	if string(data) != originalConfig {
		t.Fatalf("expected original content, got %q", string(data))
	}
}

func TestDeleteBackup(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)

	dir := NewSSHDir(sshDir)
	os.WriteFile(dir.Path("config"), []byte("test"), 0600)

	if err := CreateBackup(dir); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	backups, _ := ListBackups(dir)
	if len(backups) != 1 {
		t.Fatal("expected 1 backup")
	}

	if err := DeleteBackup(dir, backups[0].Filename); err != nil {
		t.Fatalf("DeleteBackup failed: %v", err)
	}

	backups, _ = ListBackups(dir)
	if len(backups) != 0 {
		t.Fatal("expected 0 backups after delete")
	}
}

func TestDeleteBackupPathTraversal(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	err := DeleteBackup(dir, "../../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestRestoreBackupPathTraversal(t *testing.T) {
	dir := NewSSHDir(t.TempDir())
	err := RestoreBackup(dir, "../../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}
