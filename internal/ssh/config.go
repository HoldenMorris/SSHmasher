package ssh

import (
	"fmt"
	"log"
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
	log.Printf("[DEBUG] OpenTerminal called with alias: %s", alias)
	
	// Validate the alias to prevent command injection
	// Only allow alphanumeric, hyphen, underscore, and dot
	for _, char := range alias {
		if !isValidAliasChar(char) {
			log.Printf("[DEBUG] Invalid alias character: %c", char)
			return fmt.Errorf("invalid alias: contains forbidden characters")
		}
	}

	switch runtime.GOOS {
	case "darwin":
		log.Printf("[DEBUG] macOS detected, using 'open -a Terminal'")
		// macOS: Use 'open -a Terminal' to open the default Terminal.app
		cmd := exec.Command("open", "-a", "Terminal", "--args", "ssh", alias)
		// Disown process so it doesn't die when the request ends
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		log.Printf("[DEBUG] Executing: %s", cmd.String())
		return cmd.Start()

	case "linux":
		log.Printf("[DEBUG] Linux detected, searching for terminal")
		// Try to find the default terminal
		term := getLinuxTerminal()
		if term == "" {
			return fmt.Errorf("no supported terminal emulator found")
		}
		
		log.Printf("[DEBUG] Selected terminal: %s", term)
		
		// Different terminals use different flags
		var cmd *exec.Cmd
		switch term {
		case "gnome-terminal", "xfce4-terminal", "x-terminal-emulator":
			cmd = exec.Command(term, "--", "ssh", alias)
		case "konsole":
			cmd = exec.Command(term, "-e", "ssh", alias)
		case "terminator":
			// Terminator uses -x flag with sh -c to execute shell commands
			// Pass sh, -c, and the command string as separate arguments
			cmd = exec.Command(term, "-x", "sh", "-c", "ssh "+alias+"; exec bash")
		case "xterm", "alacritty", "kitty":
			cmd = exec.Command(term, "-e", "ssh", alias)
		default:
			// Fallback: try with -e flag and keep window open on error
			cmd = exec.Command(term, "-e", "bash", "-c", "ssh "+alias+" || exec bash")
		}
		
		// Disown process so it doesn't die when the request ends
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		
		// Ensure DISPLAY is set for GUI apps
		cmd.Env = append(os.Environ(), "DISPLAY=:0")
		
		log.Printf("[DEBUG] Executing: %s", cmd.String())
		log.Printf("[DEBUG] With DISPLAY=:0")
		return cmd.Start()

	case "windows":
		log.Printf("[DEBUG] Windows detected")
		// Windows: Try Windows Terminal first, then use 'start' command
		if _, err := exec.LookPath("wt"); err == nil {
			log.Printf("[DEBUG] Using Windows Terminal (wt)")
			cmd := exec.Command("wt", "ssh", alias)
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Stdin = nil
			return cmd.Start()
		}
		// Use 'start' to open the default terminal handler
		log.Printf("[DEBUG] Using cmd /c start")
		cmd := exec.Command("cmd", "/c", "start", "ssh", alias)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		return cmd.Start()

	default:
		log.Printf("[DEBUG] Unsupported OS: %s", runtime.GOOS)
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// getLinuxTerminal finds the best available terminal emulator on Linux
func getLinuxTerminal() string {
	// Check for user config first
	if term := getTerminalFromConfig(); term != "" {
		log.Printf("[DEBUG] Using terminal from config: %s", term)
		return term
	}
	
	// 1. Try update-alternatives to get the x-terminal-emulator target
	if term := getTerminalFromAlternatives(); term != "" {
		log.Printf("[DEBUG] Found terminal from update-alternatives: %s", term)
		return term
	}
	
	// 2. Try the x-terminal-emulator symlink directly
	if _, err := exec.LookPath("x-terminal-emulator"); err == nil {
		log.Printf("[DEBUG] Using x-terminal-emulator symlink")
		return "x-terminal-emulator"
	}
	
	// 3. Try exo-open (Xfce, often available on other DEs)
	if _, err := exec.LookPath("exo-open"); err == nil {
		log.Printf("[DEBUG] Using exo-open")
		return "exo-open"
	}
	
	// 4. Fallback list of common terminals in order of preference
	terminals := []string{
		"gnome-terminal",
		"konsole", 
		"xfce4-terminal",
		"terminator",
		"alacritty",
		"kitty",
		"xterm",
	}
	
	for _, term := range terminals {
		if _, err := exec.LookPath(term); err == nil {
			log.Printf("[DEBUG] Found terminal from fallback list: %s", term)
			return term
		}
	}
	
	log.Printf("[DEBUG] No terminal emulator found")
	return ""
}

// getTerminalFromConfig reads terminal preference from ~/.config/sshmasher/config
func getTerminalFromConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	configPath := filepath.Join(home, ".config", "sshmasher", "config")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[DEBUG] Error reading config file: %v", err)
		}
		return ""
	}
	
	// Parse config file looking for terminal= setting
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "terminal=") {
			term := strings.TrimPrefix(line, "terminal=")
			term = strings.TrimSpace(term)
			// Verify the terminal exists
			if _, err := exec.LookPath(term); err == nil {
				return term
			}
			log.Printf("[DEBUG] Config specifies terminal '%s' but it's not in PATH", term)
		}
	}
	
	return ""
}

// getTerminalFromAlternatives uses update-alternatives to find the x-terminal-emulator target
func getTerminalFromAlternatives() string {
	// Run update-alternatives to get the current x-terminal-emulator
	cmd := exec.Command("update-alternatives", "--quiet", "--display", "x-terminal-emulator")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[DEBUG] update-alternatives failed: %v", err)
		return ""
	}
	
	// Parse output to find "link currently points to"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "link currently points to") {
			parts := strings.Split(line, "link currently points to")
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				// Extract the terminal name from the path
				name := filepath.Base(path)
				// Remove .wrapper suffix if present (e.g., gnome-terminal.wrapper)
				name = strings.TrimSuffix(name, ".wrapper")
				log.Printf("[DEBUG] update-alternatives points to: %s (name: %s)", path, name)
				return name
			}
		}
	}
	
	log.Printf("[DEBUG] Could not parse update-alternatives output")
	return ""
}

func isValidAliasChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_' || char == '.'
}
