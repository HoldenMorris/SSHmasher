package ssh

import (
	"fmt"
	"os"
	"path/filepath"
)

// SSHDir provides operations rooted at a given SSH directory.
// Using a configurable base path makes the code testable with temp directories.
type SSHDir struct {
	Base string
}

// DefaultSSHDir returns an SSHDir pointing at ~/.ssh.
func DefaultSSHDir() (*SSHDir, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}
	base := filepath.Join(home, ".ssh")
	return &SSHDir{Base: base}, nil
}

// NewSSHDir creates an SSHDir rooted at the given path.
func NewSSHDir(base string) *SSHDir {
	return &SSHDir{Base: base}
}

// Path returns the full path to a file within the SSH directory.
func (d *SSHDir) Path(name string) string {
	return filepath.Join(d.Base, name)
}

// ConfigPath returns the path to the SSH config file.
func (d *SSHDir) ConfigPath() string {
	return d.Path("config")
}

// KnownHostsPath returns the path to the known_hosts file.
func (d *SSHDir) KnownHostsPath() string {
	return d.Path("known_hosts")
}

// BackupDir returns the path to the backup directory (outside ~/.ssh).
func (d *SSHDir) BackupDir() string {
	return filepath.Join(filepath.Dir(d.Base), ".ssh_backups")
}

// EnsureDir creates the SSH directory if it doesn't exist with 0700 permissions.
func (d *SSHDir) EnsureDir() error {
	return os.MkdirAll(d.Base, 0700)
}

// EnsureBackupDir creates the backup directory if it doesn't exist.
func (d *SSHDir) EnsureBackupDir() error {
	return os.MkdirAll(d.BackupDir(), 0700)
}

// SetKeyPermissions sets the correct permissions for a private key (0600) and public key (0644).
func (d *SSHDir) SetKeyPermissions(name string) error {
	privPath := d.Path(name)
	pubPath := d.Path(name + ".pub")

	if err := os.Chmod(privPath, 0600); err != nil {
		return fmt.Errorf("chmod private key: %w", err)
	}
	if _, err := os.Stat(pubPath); err == nil {
		if err := os.Chmod(pubPath, 0644); err != nil {
			return fmt.Errorf("chmod public key: %w", err)
		}
	}
	return nil
}
