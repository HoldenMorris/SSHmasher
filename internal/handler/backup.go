package handler

import (
	"net/http"

	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/internal/view"
)

// Backup holds dependencies for backup/restore handlers.
type Backup struct {
	Dir *ssh.SSHDir
}

func (b *Backup) List(w http.ResponseWriter, r *http.Request) {
	backups, err := ssh.ListBackups(b.Dir)
	if err != nil {
		backups = nil
	}
	if isHTMX(r) {
		view.BackupList(backups).Render(r.Context(), w)
		return
	}
	writeJSON(w, backups)
}

func (b *Backup) Create(w http.ResponseWriter, r *http.Request) {
	if err := ssh.CreateBackup(b.Dir); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b.List(w, r)
}

func (b *Backup) Restore(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")

	// Auto-backup before restore
	if err := ssh.CreateBackup(b.Dir); err != nil {
		http.Error(w, "failed to create safety backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := ssh.RestoreBackup(b.Dir, filename); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b.List(w, r)
}

func (b *Backup) Delete(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if err := ssh.DeleteBackup(b.Dir, filename); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b.List(w, r)
}
