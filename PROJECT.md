Project Plan for Building the App
Assuming we proceed with a custom build (since no perfect match exists), here's an initial project plan for a Golang-based app called "SSHManager" (placeholder name). It will use HTML/CSS/JS for the GUI, with options for self-hosted (web server mode) or embedded (desktop app via webview). The app will focus on managing the ~/.ssh folder: keys, config, known_hosts, etc. We'll aim for a minimal viable product (MVP) first, then iterate.

## Project Tracking

**Always keep these files up to date**

`docs/TODO.md` tracks bugs, open items, and planned work.
When completing a task, fixing a bug, or discovering new work, update `docs/TODO.md` **Before AND After doing work** to reflect the current state.

`docs/plans/{FEATURE NAME}.md` Once a PLAN is devised added to the TODO fil Features section and save the plan in the folder.


1. Project Overview

Goal: Create a user-friendly tool to view, edit, generate, and manage SSH-related files in ~/.ssh without needing CLI commands. It should run locally, support cross-platform (Linux/Mac/Windows), and provide a web-based GUI.
Scope: MVP includes basic CRUD (create/read/update/delete) for keys and config; advanced features like connection testing can come later.
Assumptions: User has basic Golang knowledge; app handles sensitive data securely (e.g., no remote storage); focus on single-user, local use.
Success Criteria: App launches, interacts with ~/.ssh safely, and provides an intuitive interface.

2. Requirements and Features
Break down into must-haves (MVP) and nice-to-haves.
MVP Features:

Key Management: List all keys (public/private), generate new ones (using ssh-keygen via Go exec), edit/delete, view details (e.g., fingerprint).
Config Management: Parse/edit ~/.ssh/config (add/edit hosts, options like HostName, User, IdentityFile).
Known Hosts Management: View/edit/remove entries in known_hosts.
Backup/Restore: Snapshot and restore the entire ~/.ssh folder.
Security: Password-protect the app (optional), confirm destructive actions, handle file permissions.
GUI: Responsive web interface with tabs for sections (Keys, Config, Known Hosts); forms for editing; search/filter.

Nice-to-Haves (Post-MVP):

Test SSH connections from the app.
Import/export keys from other tools.
Multi-folder support (e.g., custom SSH dirs).
Themes or customizable UI.
Logging/audit of changes.

Non-Functional Requirements:

Cross-platform: Works on major OSes.
Performance: Quick file ops, no lag in GUI.
Security: Avoid storing passphrases; use OS keychain if needed.
Dependencies: Minimal external libs.

3. Architecture

Backend: Golang for core logic.
File handling: Use os, io, exec (for ssh-keygen), and libs like golang.org/x/crypto/ssh for parsing/validation.
API: RESTful endpoints (e.g., /api/keys, /api/config) using Gorilla Mux or Gin.

Frontend: HTML/CSS/JS (vanilla or with frameworks like Bootstrap for CSS, htmx/Alpine.js for lightweight interactivity to avoid heavy JS frameworks like React).
Deployment Modes:
Self-Hosted: Run as a local HTTP server (e.g., on localhost:8080), access via browser.
Embedded: Use Wails (Go + WebView) to package as a native desktop app with embedded browser for a seamless GUI experience. This embeds the web frontend into a desktop window.

Data Flow: Frontend sends requests to backend API; backend interacts with filesystem (~/.ssh), returns JSON; frontend renders updates.
Error Handling: Graceful failures (e.g., permission denied), user-friendly messages.

4. Tech Stack

Language: Golang (latest stable, e.g., 1.21+).
Backend Libs:
Web server: net/http or Gin.
SSH parsing: golang.org/x/crypto/ssh.
Config parser: Custom or lib like github.com/kevinburke/ssh_config.

Frontend: HTML5, CSS (Bootstrap/Tailwind), JS (vanilla or lightweight like Vue.js if needed for reactivity).
Build Tools: Go modules; Wails for embedded mode.
Testing: Go's built-in testing; perhaps Cypress for GUI.
Version Control: Git.

5. Development Phases
Estimate for a solo developer with part-time effort (e.g., 10-20 hours/week).

Phase 1: Planning & Setup (1-2 weeks)
Refine requirements (we can iterate here).
Set up repo, install Go/Wails.
Prototype basic backend (file read/write).

Phase 2: Backend Core (2-3 weeks)
Implement key generation/listing/editing.
Parse and edit config/known_hosts.
Add backup/restore.

Phase 3: Frontend GUI (2-3 weeks)
Build HTML/JS pages for each section.
Integrate with backend API.
Add responsiveness and basic styling.

Phase 4: Integration & Modes (1-2 weeks)
Test self-hosted mode.
Package embedded version with Wails.
Add security features (e.g., auth).

Phase 5: Testing & Polish (1-2 weeks)
Unit/integration tests.
Cross-OS testing.
Bug fixes, UI tweaks.

Phase 6: Deployment & Docs (1 week)
Release binary/installer.
Write README, usage guide.


Total Estimated Timeline: 8-13 weeks, depending on scope changes.
6. Risks & Mitigations

Security Risks: Mishandling private keys – Mitigate with read-only views, confirm dialogs.
OS Differences: File paths/permissions vary – Use os.ExpandEnv for ~/.
Dependencies: Wails might have platform quirks – Fallback to pure web mode.
Scope Creep: Stick to MVP; we can add features later.
Learning Curve: If unfamiliar with Wails or SSH libs, allocate extra time for docs.

This is a high-level plan – we can refine it based on your preferences (e.g., prioritize embedded vs. self-hosted, add specific features). What details do you want to flesh out first, like features, tech choices, or a prototype sketch?