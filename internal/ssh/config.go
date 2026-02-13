package ssh

import (
	"fmt"
	"os"
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
func replaceHostBlock(content, alias, replacement string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Host ") {
			blockAlias := strings.TrimSpace(strings.TrimPrefix(trimmed, "Host "))
			if blockAlias == alias {
				inBlock = true
				continue
			} else if inBlock {
				// We hit a new Host block, so the old one is done
				inBlock = false
				if replacement != "" {
					result = append(result, strings.TrimSpace(replacement))
					result = append(result, "")
					replacement = "" // only add once
				}
			}
		} else if inBlock && strings.HasPrefix(trimmed, "Host ") {
			inBlock = false
		}

		if inBlock {
			// Skip lines belonging to the old block
			continue
		}

		result = append(result, line)
	}

	// If block was at the end of file
	if inBlock && replacement != "" {
		result = append(result, strings.TrimSpace(replacement))
	}

	return strings.Join(result, "\n")
}
