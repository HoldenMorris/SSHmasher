package handler

import (
	"io/fs"
	"net/http"

	"github.com/holden/sshmasher/internal/ssh"
)

// NewRouter creates the central HTTP router with all routes wired up.
// staticFS should contain the static/ directory contents.
func NewRouter(dir *ssh.SSHDir, staticFS fs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	pages := &Pages{Dir: dir}
	keys := &Keys{Dir: dir}
	config := &Config{Dir: dir}
	knownhosts := &KnownHosts{Dir: dir}
	backup := &Backup{Dir: dir}

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	// Page routes — render keys page at root (no redirect, for Wails compat)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			pages.KeysPage(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /keys", pages.KeysPage)
	mux.HandleFunc("GET /config", pages.ConfigPage)
	mux.HandleFunc("GET /knownhosts", pages.KnownHostsPage)
	mux.HandleFunc("GET /backup", pages.BackupPage)

	// API: Keys
	mux.HandleFunc("GET /api/keys", keys.List)
	mux.HandleFunc("POST /api/keys", keys.Generate)
	mux.HandleFunc("GET /api/keys/{name}", keys.Get)
	mux.HandleFunc("DELETE /api/keys/{name}", keys.Delete)

	// API: Config
	mux.HandleFunc("GET /api/config/hosts", config.ListHosts)
	mux.HandleFunc("POST /api/config/hosts", config.AddHost)
	mux.HandleFunc("GET /api/config/hosts/{alias}", config.GetHost)
	mux.HandleFunc("PUT /api/config/hosts/{alias}", config.UpdateHost)
	mux.HandleFunc("DELETE /api/config/hosts/{alias}", config.DeleteHost)
	mux.HandleFunc("GET /api/config/raw", config.GetRaw)
	mux.HandleFunc("PUT /api/config/raw", config.PutRaw)

	// API: Known Hosts
	mux.HandleFunc("GET /api/knownhosts", knownhosts.List)
	mux.HandleFunc("DELETE /api/knownhosts/{line}", knownhosts.Delete)
	mux.HandleFunc("GET /api/knownhosts/raw", knownhosts.GetRaw)
	mux.HandleFunc("PUT /api/knownhosts/raw", knownhosts.PutRaw)

	// API: Backup
	mux.HandleFunc("GET /api/backup", backup.List)
	mux.HandleFunc("POST /api/backup", backup.Create)
	mux.HandleFunc("POST /api/backup/{filename}/restore", backup.Restore)
	mux.HandleFunc("DELETE /api/backup/{filename}", backup.Delete)

	return mux
}

// WithMiddleware wraps a handler with all middleware.
func WithMiddleware(h http.Handler) http.Handler {
	return Logging(SecurityHeaders(h))
}
