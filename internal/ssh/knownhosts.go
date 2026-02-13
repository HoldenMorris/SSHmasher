package ssh

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/holden/sshmasher/internal/model"
	gossh "golang.org/x/crypto/ssh"
)

// ListKnownHosts parses the known_hosts file and returns all entries.
func ListKnownHosts(dir *SSHDir) ([]model.KnownHostEntry, error) {
	data, err := os.ReadFile(dir.KnownHostsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read known_hosts: %w", err)
	}

	var entries []model.KnownHostEntry
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		entry := parseKnownHostLine(i+1, line)
		entries = append(entries, entry)
	}

	return entries, nil
}

// FilterKnownHosts filters entries by search string (matches hosts or key type).
func FilterKnownHosts(entries []model.KnownHostEntry, search string) []model.KnownHostEntry {
	search = strings.ToLower(search)
	var filtered []model.KnownHostEntry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Hosts), search) ||
			strings.Contains(strings.ToLower(e.KeyType), search) ||
			strings.Contains(strings.ToLower(e.Fingerprint), search) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// RemoveKnownHost removes the entry at the given 1-based line number.
func RemoveKnownHost(dir *SSHDir, line int) error {
	data, err := os.ReadFile(dir.KnownHostsPath())
	if err != nil {
		return fmt.Errorf("read known_hosts: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if line < 1 || line > len(lines) {
		return fmt.Errorf("line %d out of range", line)
	}

	// Remove the line (1-based index)
	lines = append(lines[:line-1], lines[line:]...)
	return os.WriteFile(dir.KnownHostsPath(), []byte(strings.Join(lines, "\n")), 0644)
}

// WriteKnownHosts overwrites the known_hosts file.
func WriteKnownHosts(dir *SSHDir, content string) error {
	if err := dir.EnsureDir(); err != nil {
		return err
	}
	return os.WriteFile(dir.KnownHostsPath(), []byte(content), 0644)
}

func parseKnownHostLine(lineNum int, line string) model.KnownHostEntry {
	entry := model.KnownHostEntry{Line: lineNum}

	parts := strings.Fields(line)
	if len(parts) < 3 {
		entry.Hosts = line
		return entry
	}

	entry.Hosts = parts[0]
	entry.KeyType = parts[1]
	entry.Key = parts[2]
	entry.IsHashed = strings.HasPrefix(entry.Hosts, "|1|")

	// Try to compute fingerprint
	keyBytes, err := base64.StdEncoding.DecodeString(entry.Key)
	if err == nil {
		pubKey, err := gossh.ParsePublicKey(keyBytes)
		if err == nil {
			hash := sha256.Sum256(pubKey.Marshal())
			entry.Fingerprint = "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])
		}
	}

	return entry
}
