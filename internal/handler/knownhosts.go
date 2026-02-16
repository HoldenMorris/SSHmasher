package handler

import (
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/internal/view"
)

// KnownHosts holds dependencies for known hosts handlers.
type KnownHosts struct {
	Dir *ssh.SSHDir
}

func (kh *KnownHosts) List(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	entries, err := ssh.ListKnownHosts(kh.Dir)
	if err != nil {
		entries = nil
	}
	configHosts, _ := ssh.ListHosts(kh.Dir)
	if search != "" {
		entries = ssh.FilterKnownHosts(entries, search)
	}
	if isHTMX(r) {
		view.KnownHostsTable(entries, configHosts).Render(r.Context(), w)
		return
	}
	writeJSON(w, entries)
}

func (kh *KnownHosts) Delete(w http.ResponseWriter, r *http.Request) {
	lineStr := r.PathValue("line")
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		http.Error(w, "invalid line number", http.StatusBadRequest)
		return
	}

	if err := ssh.RemoveKnownHost(kh.Dir, line); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entries, _ := ssh.ListKnownHosts(kh.Dir)
	configHosts, _ := ssh.ListHosts(kh.Dir)
	if isHTMX(r) {
		view.KnownHostsTable(entries, configHosts).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (kh *KnownHosts) GetRaw(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(kh.Dir.KnownHostsPath())
	if err != nil {
		content = []byte{}
	}
	if isHTMX(r) {
		view.KnownHostsRawEditor(string(content)).Render(r.Context(), w)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

func (kh *KnownHosts) PutRaw(w http.ResponseWriter, r *http.Request) {
	var content string
	if r.Header.Get("Content-Type") == "application/json" {
		body, _ := io.ReadAll(r.Body)
		content = string(body)
	} else {
		r.ParseForm()
		content = r.FormValue("content")
	}

	if err := ssh.WriteKnownHosts(kh.Dir, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		entries, _ := ssh.ListKnownHosts(kh.Dir)
		configHosts, _ := ssh.ListHosts(kh.Dir)
		view.KnownHostsTable(entries, configHosts).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (kh *KnownHosts) Lookup(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")
	port := r.URL.Query().Get("port")
	
	if hostname == "" {
		http.Error(w, "hostname required", http.StatusBadRequest)
		return
	}
	
	entries, err := ssh.LookupKnownHost(kh.Dir, hostname, port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if isHTMX(r) {
		view.KnownHostLookupResults(hostname, port, entries).Render(r.Context(), w)
		return
	}
	writeJSON(w, entries)
}

func (kh *KnownHosts) Add(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	hostname := r.FormValue("hostname")
	port := r.FormValue("port")

	if hostname == "" {
		http.Error(w, "hostname required", http.StatusBadRequest)
		return
	}

	if err := ssh.AddKnownHost(kh.Dir, hostname, port); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated list
	entries, _ := ssh.ListKnownHosts(kh.Dir)
	configHosts, _ := ssh.ListHosts(kh.Dir)
	if isHTMX(r) {
		view.KnownHostsTable(entries, configHosts).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, entries)
}
