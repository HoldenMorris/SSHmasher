package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/holden/sshmasher/internal/model"
	sshconfig "github.com/kevinburke/ssh_config"
)

// ListHosts parses the SSH config and returns all host entries.
func ListHosts(dir *SSHDir) ([]model.HostEntry, error) {
	f, err := os.Open(dir.ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	cfg, err := sshconfig.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	var hosts []model.HostEntry
	for _, host := range cfg.Hosts {
		patterns := host.Patterns
		if len(patterns) == 0 {
			continue
		}

		alias := patterns[0].String()
		if alias == "*" {
			continue // skip global wildcard
		}

		entry := model.HostEntry{
			Alias:   alias,
			Options: make(map[string]string),
		}

		for _, node := range host.Nodes {
			kv, ok := node.(*sshconfig.KV)
			if !ok {
				continue
			}
			switch strings.ToLower(kv.Key) {
			case "hostname":
				entry.HostName = kv.Value
			case "user":
				entry.User = kv.Value
			case "port":
				entry.Port = kv.Value
			case "identityfile":
				entry.IdentityFile = kv.Value
			default:
				entry.Options[kv.Key] = kv.Value
			}
		}

		hosts = append(hosts, entry)
	}

	return hosts, nil
}

// GetHost returns a single host entry by alias.
func GetHost(dir *SSHDir, alias string) (*model.HostEntry, error) {
	hosts, err := ListHosts(dir)
	if err != nil {
		return nil, err
	}
	for _, h := range hosts {
		if h.Alias == alias {
			return &h, nil
		}
	}
	return nil, fmt.Errorf("host not found: %s", alias)
}

// AddHost appends a new host block to the SSH config file.
func AddHost(dir *SSHDir, host model.HostEntry) error {
	if err := dir.EnsureDir(); err != nil {
		return err
	}

	block := formatHostBlock(host)

	f, err := os.OpenFile(dir.ConfigPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(block)
	return err
}

// UpdateHost replaces a host block in the config file.
func UpdateHost(dir *SSHDir, host model.HostEntry) error {
	content, err := os.ReadFile(dir.ConfigPath())
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	updated := replaceHostBlock(string(content), host.Alias, formatHostBlock(host))
	return os.WriteFile(dir.ConfigPath(), []byte(updated), 0600)
}

// DeleteHost removes a host block from the config file.
func DeleteHost(dir *SSHDir, alias string) error {
	content, err := os.ReadFile(dir.ConfigPath())
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	updated := replaceHostBlock(string(content), alias, "")
	return os.WriteFile(dir.ConfigPath(), []byte(updated), 0600)
}

// WriteConfig overwrites the SSH config file with the given content.
func WriteConfig(dir *SSHDir, content string) error {
	if err := dir.EnsureDir(); err != nil {
		return err
	}
	return os.WriteFile(dir.ConfigPath(), []byte(content), 0600)
}

func formatHostBlock(host model.HostEntry) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\nHost %s\n", host.Alias))
	if host.HostName != "" {
		b.WriteString(fmt.Sprintf("    HostName %s\n", host.HostName))
	}
	if host.User != "" {
		b.WriteString(fmt.Sprintf("    User %s\n", host.User))
	}
	if host.Port != "" {
		b.WriteString(fmt.Sprintf("    Port %s\n", host.Port))
	}
	if host.IdentityFile != "" {
		b.WriteString(fmt.Sprintf("    IdentityFile %s\n", host.IdentityFile))
	}
	for k, v := range host.Options {
		b.WriteString(fmt.Sprintf("    %s %s\n", k, v))
	}
	return b.String()
}

// replaceHostBlock finds the "Host <alias>" block in content and replaces it with replacement.
// If replacement is empty, the block is removed.
// Handles multi-pattern Host lines like "Host foo bar" by checking if alias matches any pattern.
func replaceHostBlock(content, alias, replacement string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Host ") {
			blockPatterns := strings.TrimSpace(strings.TrimPrefix(trimmed, "Host "))
			patterns := strings.Fields(blockPatterns)
			aliasMatched := false
			for _, p := range patterns {
				if p == alias {
					aliasMatched = true
					break
				}
			}

			if aliasMatched {
				inBlock = true
				continue
			} else if inBlock {
				inBlock = false
				if replacement != "" {
					result = append(result, strings.TrimSpace(replacement))
					result = append(result, "")
					replacement = ""
				}
			}
		} else if inBlock && strings.HasPrefix(trimmed, "Host ") {
			inBlock = false
		}

		if inBlock {
			continue
		}

		result = append(result, line)
	}

	if inBlock && replacement != "" {
		result = append(result, strings.TrimSpace(replacement))
	}

	return strings.Join(result, "\n")
}

// KeyRefCount returns a map of key names to the number of config entries that reference them.
func KeyRefCount(dir *SSHDir) (map[string]int, error) {
	hosts, err := ListHosts(dir)
	if err != nil {
		return nil, err
	}

	refCount := make(map[string]int)
	for _, host := range hosts {
		if host.IdentityFile != "" {
			refCount[host.IdentityFile]++
		}
	}
	return refCount, nil
}

// IsKeyFile checks if a key file (or its .pub variant) exists in the SSH directory.
// The identityFile parameter can be a full path or just a name like "id_ed25519".
func (d *SSHDir) IsKeyFile(identityFile string) bool {
	// Check as-is (full path or relative path)
	if _, err := os.Stat(d.Path(identityFile)); err == nil {
		return true
	}
	// Check just the filename in ~/.ssh
	name := filepath.Base(identityFile)
	if _, err := os.Stat(d.Path(name)); err == nil {
		return true
	}
	return false
}

// OpenTerminal opens a terminal and runs SSH to connect to the given host.
// It opens the default terminal for the current OS.
func OpenTerminal(alias string) error {
	// Validate the alias to prevent command injection
	// Only allow alphanumeric, hyphen, underscore, and dot
	for _, char := range alias {
		if !isValidAliasChar(char) {
			return fmt.Errorf("invalid alias: contains forbidden characters")
		}
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: Uses AppleScript to tell Terminal to run the command
		script := fmt.Sprintf("tell application \"Terminal\" to do script \"ssh %s\"", alias)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Start()

	case "linux":
		// Try common Linux terminals
		terminals := []string{"gnome-terminal", "konsole", "xfce4-terminal", "xterm", "alacritty", "kitty"}
		for _, term := range terminals {
			if _, err := exec.LookPath(term); err == nil {
				var cmd *exec.Cmd
				switch term {
				case "gnome-terminal", "xfce4-terminal":
					cmd = exec.Command(term, "--", "ssh", alias)
				case "konsole":
					cmd = exec.Command(term, "-e", "ssh", alias)
				case "xterm", "alacritty", "kitty":
					cmd = exec.Command(term, "-e", "ssh", alias)
				default:
					cmd = exec.Command(term, "-e", "ssh", alias)
				}
				return cmd.Start()
			}
		}
		return fmt.Errorf("no supported terminal emulator found")

	case "windows":
		// Windows: Try Windows Terminal first, then cmd
		if _, err := exec.LookPath("wt"); err == nil {
			cmd := exec.Command("wt", "ssh", alias)
			return cmd.Start()
		}
		cmd := exec.Command("cmd", "/c", "start", "ssh", alias)
		return cmd.Start()

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func isValidAliasChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_' || char == '.'
}
