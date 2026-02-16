package handler

import (
	"net/http"

	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/internal/view"
)

// Pages holds dependencies for full-page HTML handlers.
type Pages struct {
	Dir *ssh.SSHDir
}

func (p *Pages) KeysPage(w http.ResponseWriter, r *http.Request) {
	keys, err := ssh.ListKeys(p.Dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refCount, err := ssh.KeyRefCount(p.Dir)
	if err != nil {
		refCount = make(map[string]int)
	}
	view.KeysPage(keys, refCount).Render(r.Context(), w)
}

func (p *Pages) ConfigPage(w http.ResponseWriter, r *http.Request) {
	hosts, err := ssh.ListHosts(p.Dir)
	if err != nil {
		hosts = nil // show empty state if config doesn't exist
	}
	keys, err := ssh.ListKeys(p.Dir)
	if err != nil {
		keys = nil
	}
	view.ConfigPage(hosts, keys, p.Dir).Render(r.Context(), w)
}

func (p *Pages) KnownHostsPage(w http.ResponseWriter, r *http.Request) {
	entries, err := ssh.ListKnownHosts(p.Dir)
	if err != nil {
		entries = nil
	}
	configHosts, err := ssh.ListHosts(p.Dir)
	if err != nil {
		configHosts = nil
	}
	// Build map of line numbers to config host aliases using ssh-keygen -F
	lineToHosts := ssh.MatchConfigHostsToKnownHosts(p.Dir, configHosts)
	view.KnownHostsPage(entries, configHosts, lineToHosts).Render(r.Context(), w)
}

func (p *Pages) BackupPage(w http.ResponseWriter, r *http.Request) {
	backups, err := ssh.ListBackups(p.Dir)
	if err != nil {
		backups = nil
	}
	view.BackupPage(backups).Render(r.Context(), w)
}
