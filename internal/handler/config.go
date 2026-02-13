package handler

import (
	"io"
	"net/http"
	"os"

	"github.com/holden/sshmasher/internal/model"
	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/internal/view"
)

// Config holds dependencies for SSH config handlers.
type Config struct {
	Dir *ssh.SSHDir
}

func (c *Config) ListHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := ssh.ListHosts(c.Dir)
	if err != nil {
		hosts = nil
	}
	if isHTMX(r) {
		view.ConfigHostsTable(hosts).Render(r.Context(), w)
		return
	}
	writeJSON(w, hosts)
}

func (c *Config) GetHost(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	host, err := ssh.GetHost(c.Dir, alias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if isHTMX(r) {
		view.ConfigEditForm(*host).Render(r.Context(), w)
		return
	}
	writeJSON(w, host)
}

func (c *Config) AddHost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	host := model.HostEntry{
		Alias:        r.FormValue("alias"),
		HostName:     r.FormValue("hostname"),
		User:         r.FormValue("user"),
		Port:         r.FormValue("port"),
		IdentityFile: r.FormValue("identityfile"),
	}

	if host.Alias == "" || host.HostName == "" {
		http.Error(w, "alias and hostname are required", http.StatusBadRequest)
		return
	}

	if err := ssh.AddHost(c.Dir, host); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.ListHosts(w, r)
}

func (c *Config) UpdateHost(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	host := model.HostEntry{
		Alias:        alias,
		HostName:     r.FormValue("hostname"),
		User:         r.FormValue("user"),
		Port:         r.FormValue("port"),
		IdentityFile: r.FormValue("identityfile"),
	}

	if err := ssh.UpdateHost(c.Dir, host); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.ListHosts(w, r)
}

func (c *Config) DeleteHost(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	if err := ssh.DeleteHost(c.Dir, alias); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.ListHosts(w, r)
}

func (c *Config) GetRaw(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(c.Dir.ConfigPath())
	if err != nil {
		content = []byte{}
	}
	if isHTMX(r) {
		view.ConfigRawEditor(string(content)).Render(r.Context(), w)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

func (c *Config) PutRaw(w http.ResponseWriter, r *http.Request) {
	var content string
	if r.Header.Get("Content-Type") == "application/json" {
		body, _ := io.ReadAll(r.Body)
		content = string(body)
	} else {
		r.ParseForm()
		content = r.FormValue("content")
	}

	if err := ssh.WriteConfig(c.Dir, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		hosts, _ := ssh.ListHosts(c.Dir)
		view.ConfigHostsTable(hosts).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
