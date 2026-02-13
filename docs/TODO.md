# SSHmasher TODO

## Current Sprint

- [ ] Add handler tests for API endpoints
- [ ] Add loading indicators (`hx-indicator`) on all HTMX actions
- [ ] Empty state improvements (icons, call-to-action buttons)
- [ ] Error toast messages instead of plain text errors
- [ ] Search/filter on keys table

## Bugs

- [ ] Wails webview doesn't follow HTTP redirects — root `/` renders keys page directly as workaround
- [ ] Config update (`replaceHostBlock`) may misbehave with multi-pattern `Host` lines (e.g. `Host foo bar`)
- [ ] Known hosts line numbers shift after deletion — UI should refresh full table (currently does)

## Short Term

- [ ] Dark mode toggle (Pico CSS supports `data-theme="dark"`)
- [ ] Copy public key to clipboard button
- [ ] Key passphrase change (ssh-keygen -p)
- [ ] Show key file sizes and modification dates in table
- [ ] Confirmation dialog component (replace browser `confirm()` with styled modal)
- [ ] Config host duplicate detection before add
- [ ] Backup auto-cleanup (keep last N backups option)
- [ ] Backup download via browser (Content-Disposition header)

## Medium Term

- [ ] SSH connection testing (dial TCP + SSH handshake)
- [ ] Import keys from file upload
- [ ] Export key pairs as zip
- [ ] Multi-folder support (custom SSH dirs beyond ~/.ssh)
- [ ] Undo/redo for config and known_hosts edits
- [ ] Audit log of all changes made through the app
- [ ] Config syntax validation before save
- [ ] Known hosts: resolve hashed entries where possible

## Long Term

- [ ] Agent forwarding management (ssh-add integration)
- [ ] SSH certificate support (CA signing)
- [ ] FIDO2/security key support for key generation
- [ ] Multi-user mode with authentication
- [ ] Remote SSH dir management (manage keys on remote hosts)
- [ ] Theme customization
- [ ] Keyboard shortcuts
- [ ] Wails v3 migration (when stable)
- [ ] CI/CD pipeline with cross-platform builds
- [ ] Homebrew / AUR / Snap packaging
