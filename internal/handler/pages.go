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
	view.KeysPage(keys).Render(r.Context(), w)
}

func (p *Pages) ConfigPage(w http.ResponseWriter, r *http.Request) {
	hosts, err := ssh.ListHosts(p.Dir)
	if err != nil {
		hosts = nil // show empty state if config doesn't exist
	}
	view.ConfigPage(hosts).Render(r.Context(), w)
}

func (p *Pages) KnownHostsPage(w http.ResponseWriter, r *http.Request) {
	entries, err := ssh.ListKnownHosts(p.Dir)
	if err != nil {
		entries = nil
	}
	view.KnownHostsPage(entries).Render(r.Context(), w)
}

func (p *Pages) BackupPage(w http.ResponseWriter, r *http.Request) {
	backups, err := ssh.ListBackups(p.Dir)
	if err != nil {
		backups = nil
	}
	view.BackupPage(backups).Render(r.Context(), w)
}
