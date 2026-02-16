# SSHmasher TODO

## Current Sprint

- [ ] Add handler tests for API endpoints
- [x] Add loading indicators (`hx-indicator`) on all HTMX actions
- [x] Empty state improvements (icons, call-to-action buttons)
- [x] Search/filter on keys table
- [x] Search/filter on config table
- [x] Search/filter on known hosts table (already existed)

## Bugs

- [x] Wails webview doesn't follow HTTP redirects — root `/` renders keys page directly as workaround
- [x] Config update (`replaceHostBlock`) may misbehave with multi-pattern `Host` lines (e.g. `Host foo bar`)
- [x] Make action buttons icons because the words "edit,delete etc" are text wrapping.
- [x] Config table does not shoe 'Identity File' in the table
- [x] The div.grid under the hgroup need some bottom margin
- [x] Known hosts line numbers shift after deletion — UI should refresh full table (currently does)
- [x] Add/Edit host forms should be modal boxes instead of inline expand/collapse
- [x] Generate new key should use a modal instead of inline expand/collapse
- [x] Add column to keys table showing config references
- [x] Show red color on config row if it references a broken key


## Short Term

- [x] Add a col in the keys table to indicate how many configs refrence it as an Identitiy file.
- [x] Copy public key to clipboard button
- [x] Key comment edit
- [x] Show key file sizes in table
- [x] Show key modification dates (n months ago) in table
- [x] Backup download via browser (Content-Disposition header)
- [ ] Lookup config hostsname:port in known_hosts to diaplay in the known hosts table e.g.   ssh-keygen -F "[sta-securemail-pl-www1.synaq.com]:50022" shows ```# Host [sta-securemail-pl-www1.synaq.com]:50022 found: line 2
|1|vGWEBG0Tr73LLXn3RdPNwKP6qwM=|F/ckrdoJu5W7Mk8FMqc0Wa2qRkU= ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEApGiVep7IeEzt3TCm5jveTdD6lYI4sAlwk1LDxd6ZKw7cW7Rp88ct8fVaDJIgQ6MiLEdnEff/VgHJHLpcy1lOCY8DsEx8iY69iNtLyrJmrLkhgt9VhZHfdebhEIKY2qehqVnq2sJ/RbovaJU/SR6oPZu9FoJk4Ot/qV96BsdH3EeLH31r/nomKn3EH4ce66G1dujGh9SvLZmcEwmd8LzyObjKw89Hqvw58+DXKBFXuHXFNWL5TLUNBraSThZgNWRRst3ti5HIN62ByZ4eq0vW9XDR7GLGBNggNOmUOP5cFb+yxFVt16B3ReV08Fj5/oIsRJlBrtP1h7XjihLBPQ0/FQ==
```
- [ ] lookup common hostname like github.com and bitbucket.org in the same way
- [ ] Key passphrase change (ssh-keygen -p)
- [ ] Confirmation dialog component (replace browser `confirm()` with styled modal)
- [x] Config host duplicate detection before add
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
