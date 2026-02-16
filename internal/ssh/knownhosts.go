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

// LookupKnownHost searches for a hostname in known_hosts using ssh-keygen -F.
// Returns the matching entries or nil if not found.
func LookupKnownHost(dir *SSHDir, hostname string, port string) ([]model.KnownHostEntry, error) {
	// Build the host:port string if port is provided
	target := hostname
	if port != "" && port != "22" {
		target = fmt.Sprintf("[%s]:%s", hostname, port)
	}

	// Use ssh-keygen -F to lookup the host
	cmd := exec.Command("ssh-keygen", "-F", target, "-f", dir.KnownHostsPath())
	output, err := cmd.CombinedOutput()
	if err != nil {
		// ssh-keygen returns exit code 1 if host not found, which is not an error for us
		if strings.Contains(string(output), "not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("ssh-keygen -F failed: %s: %w", string(output), err)
	}

	// Parse the output
	var entries []model.KnownHostEntry
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Parse the entry
		entry := parseKnownHostLine(0, line)
		if entry.Hosts != "" && entry.KeyType != "" {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// AddKnownHost adds a hostname to known_hosts by scanning it with ssh-keyscan.
// If the host already exists, it will be updated.
func AddKnownHost(dir *SSHDir, hostname string, port string) error {
	if err := dir.EnsureDir(); err != nil {
		return err
	}

	// Use ssh-keyscan to get the host key
	cmd := exec.Command("ssh-keyscan", "-H", "-p", "22", hostname)
	if port != "" && port != "22" {
		cmd = exec.Command("ssh-keyscan", "-H", "-p", port, hostname)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keyscan failed: %s: %w", string(output), err)
	}

	// Remove existing entries for this host first
	existing, _ := ListKnownHosts(dir)
	var newContent strings.Builder
	for _, entry := range existing {
		// Skip entries matching this host
		if !strings.Contains(entry.Hosts, hostname) {
			newContent.WriteString(fmt.Sprintf("%s %s %s\n", entry.Hosts, entry.KeyType, entry.Key))
		}
	}

	// Append new entries
	newContent.WriteString(string(output))

	return os.WriteFile(dir.KnownHostsPath(), []byte(newContent.String()), 0644)
}

// MatchConfigHostsToKnownHosts uses ssh-keygen -F to lookup each config host
// and returns a map of line numbers to config host aliases
func MatchConfigHostsToKnownHosts(dir *SSHDir, configHosts []model.HostEntry) map[int][]string {
	lineToHosts := make(map[int][]string)

	// First check config hosts
	for _, host := range configHosts {
		target := host.HostName
		if host.Port != "" && host.Port != "22" {
			target = fmt.Sprintf("[%s]:%s", host.HostName, host.Port)
		}

		cmd := exec.Command("ssh-keygen", "-F", target, "-f", dir.KnownHostsPath())
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue // Host not found or error
		}

		// Parse output to find line number
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# Host") && strings.Contains(line, "found: line") {
				// Extract line number: "# Host [hostname]:port found: line X"
				parts := strings.Split(line, "line ")
				if len(parts) == 2 {
					lineNumStr := strings.TrimSpace(parts[1])
					lineNum := 0
					fmt.Sscanf(lineNumStr, "%d", &lineNum)
					if lineNum > 0 {
						lineToHosts[lineNum] = append(lineToHosts[lineNum], host.Alias)
					}
				}
			}
		}
	}

	// Also check common hosts
	commonHosts := []string{"github.com", "bitbucket.org", "gitlab.com"}
	for _, hostname := range commonHosts {
		cmd := exec.Command("ssh-keygen", "-F", hostname, "-f", dir.KnownHostsPath())
		output, err := cmd.CombinedOutput()
		if err != nil {
			continue // Host not found or error
		}

		// Parse output to find line number
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "# Host") && strings.Contains(line, "found: line") {
				// Extract line number: "# Host hostname found: line X"
				parts := strings.Split(line, "line ")
				if len(parts) == 2 {
					lineNumStr := strings.TrimSpace(parts[1])
					lineNum := 0
					fmt.Sscanf(lineNumStr, "%d", &lineNum)
					if lineNum > 0 {
						// Check if this hostname is already added for this line
						alreadyAdded := false
						for _, existing := range lineToHosts[lineNum] {
							if existing == hostname {
								alreadyAdded = true
								break
							}
						}
						if !alreadyAdded {
							lineToHosts[lineNum] = append(lineToHosts[lineNum], hostname)
						}
					}
				}
			}
		}
	}

	return lineToHosts
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
