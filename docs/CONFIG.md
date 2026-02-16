# SSHmasher Configuration

## Terminal Configuration (Linux only)

You can configure your preferred terminal emulator by creating a config file at:

```
~/.config/sshmasher/config
```

### Example Config

```bash
# Set your preferred terminal
terminal=terminator

# Or use any terminal in your PATH
terminal=alacritty
terminal=kitty
terminal=gnome-terminal
```

### Terminal Detection Priority

If no config is set, SSHmasher will try to find a terminal in this order:

1. **Config file** (`~/.config/sshmasher/config`) - Your manual preference
2. **update-alternatives** - Uses the system default (Debian/Ubuntu)
3. **x-terminal-emulator** - The symlink to the default terminal
4. **exo-open** - Xfce's terminal launcher
5. **Fallback list** - Checks for common terminals:
   - gnome-terminal
   - konsole
   - xfce4-terminal
   - terminator
   - alacritty
   - kitty
   - xterm

### Debug Logging

SSHmasher logs terminal detection steps. To see the logs, run the server directly:

```bash
go run ./cmd/server
```

You'll see output like:
```
[DEBUG] OpenTerminal called with alias: myserver
[DEBUG] Linux detected, searching for terminal
[DEBUG] Using terminal from config: terminator
[DEBUG] Selected terminal: terminator
[DEBUG] Executing: /usr/bin/terminator -e ssh myserver
```

### Creating the Config Directory

```bash
mkdir -p ~/.config/sshmasher
echo "terminal=your-terminal" > ~/.config/sshmasher/config
```

## Security Notes

- Terminal aliases are validated to prevent command injection
- Only alphanumeric characters, hyphens, underscores, and dots are allowed in host aliases
- The terminal command is constructed safely without shell interpolation
