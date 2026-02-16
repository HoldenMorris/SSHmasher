package ssh

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/holden/sshmasher/internal/model"
	gossh "golang.org/x/crypto/ssh"
)

// ListKeys scans the SSH directory for key pairs and returns metadata about each.
func ListKeys(dir *SSHDir) ([]model.SSHKey, error) {
	entries, err := os.ReadDir(dir.Base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read ssh dir: %w", err)
	}

	// Collect .pub files, then check for corresponding private keys
	var keys []model.SSHKey
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".pub")
		key, err := parseKeyPair(dir, name)
		if err != nil {
			continue // skip unparseable keys
		}
		keys = append(keys, *key)
	}
	return keys, nil
}

// GetKey returns detailed info for a specific key by name.
func GetKey(dir *SSHDir, name string) (*model.SSHKey, error) {
	pubPath := dir.Path(name + ".pub")
	if _, err := os.Stat(pubPath); err != nil {
		return nil, fmt.Errorf("key not found: %s", name)
	}
	return parseKeyPair(dir, name)
}

// GenerateKey runs ssh-keygen to generate a new key pair.
func GenerateKey(dir *SSHDir, req model.KeyGenRequest) error {
	if err := dir.EnsureDir(); err != nil {
		return err
	}

	keyPath := dir.Path(req.Name)

	// Don't overwrite existing keys
	if _, err := os.Stat(keyPath); err == nil {
		return fmt.Errorf("key already exists: %s", req.Name)
	}

	args := []string{
		"-t", req.Type,
		"-f", keyPath,
		"-N", req.Passphrase,
	}

	if req.Comment != "" {
		args = append(args, "-C", req.Comment)
	}

	if req.Bits > 0 && (req.Type == "rsa" || req.Type == "ecdsa") {
		args = append(args, "-b", fmt.Sprintf("%d", req.Bits))
	}

	cmd := exec.Command("ssh-keygen", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keygen failed: %s: %w", string(output), err)
	}

	return dir.SetKeyPermissions(req.Name)
}

// DeleteKey removes both private and public key files.
func DeleteKey(dir *SSHDir, name string) error {
	privPath := dir.Path(name)
	pubPath := dir.Path(name + ".pub")

	// At least one must exist
	privExists := fileExists(privPath)
	pubExists := fileExists(pubPath)

	if !privExists && !pubExists {
		return fmt.Errorf("key not found: %s", name)
	}

	if privExists {
		if err := os.Remove(privPath); err != nil {
			return fmt.Errorf("remove private key: %w", err)
		}
	}
	if pubExists {
		if err := os.Remove(pubPath); err != nil {
			return fmt.Errorf("remove public key: %w", err)
		}
	}
	return nil
}

// UpdateKeyComment changes the comment on a key using ssh-keygen -c.
func UpdateKeyComment(dir *SSHDir, name, newComment string) error {
	keyPath := dir.Path(name)

	if !fileExists(keyPath) {
		return fmt.Errorf("key not found: %s", name)
	}

	cmd := exec.Command("ssh-keygen", "-c", "-f", keyPath, "-C", newComment)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keygen -c failed: %s: %w", string(output), err)
	}

	return nil
}

func parseKeyPair(dir *SSHDir, name string) (*model.SSHKey, error) {
	pubPath := dir.Path(name + ".pub")
	pubData, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}

	pubKey, comment, _, _, err := gossh.ParseAuthorizedKey(pubData)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	info, _ := os.Stat(pubPath)
	hash := sha256.Sum256(pubKey.Marshal())
	fingerprint := "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])

	privPath := dir.Path(name)
	hasPrivate := fileExists(privPath)

	var totalSize int64
	if pubInfo, err := os.Stat(pubPath); err == nil {
		totalSize += pubInfo.Size()
	}
	if privInfo, err := os.Stat(privPath); err == nil {
		totalSize += privInfo.Size()
	}

	return &model.SSHKey{
		Name:        name,
		Type:        pubKey.Type(),
		Fingerprint: fingerprint,
		PublicKey:   strings.TrimSpace(string(pubData)),
		Comment:     comment,
		HasPrivate:  hasPrivate,
		ModTime:     info.ModTime(),
		Size:        totalSize,
	}, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

