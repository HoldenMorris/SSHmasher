package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
	keys, err := ssh.ListKeys(c.Dir)
	if err != nil {
		keys = nil
	}

	search := r.URL.Query().Get("search")
	if search != "" {
		var filtered []model.HostEntry
		searchLower := strings.ToLower(search)
		for _, host := range hosts {
			if strings.Contains(strings.ToLower(host.Alias), searchLower) ||
				strings.Contains(strings.ToLower(host.HostName), searchLower) ||
				strings.Contains(strings.ToLower(host.User), searchLower) ||
				strings.Contains(strings.ToLower(host.IdentityFile), searchLower) {
				filtered = append(filtered, host)
			}
		}
		hosts = filtered
	}

	if isHTMX(r) {
		view.ConfigHostsTable(hosts, keys, c.Dir).Render(r.Context(), w)
		return
	}
	writeJSON(w, hosts)
}

func (c *Config) NewHost(w http.ResponseWriter, r *http.Request) {
	keys, err := ssh.ListKeys(c.Dir)
	if err != nil {
		keys = nil
	}
	if isHTMX(r) {
		view.ConfigAddForm(keys).Render(r.Context(), w)
		return
	}
	writeJSON(w, nil)
}

func (c *Config) GetHost(w http.ResponseWriter, r *http.Request) {
	alias := r.PathValue("alias")
	host, err := ssh.GetHost(c.Dir, alias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	keys, err := ssh.ListKeys(c.Dir)
	if err != nil {
		keys = nil
	}
	if isHTMX(r) {
		view.ConfigEditForm(*host, keys).Render(r.Context(), w)
		return
	}
	writeJSON(w, host)
}

func (c *Config) AddHost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	alias := r.FormValue("alias")
	hostName := r.FormValue("hostname")

	if alias == "" || hostName == "" {
		http.Error(w, "alias and hostname are required", http.StatusBadRequest)
		return
	}

	// Check for duplicate
	existing, err := ssh.GetHost(c.Dir, alias)
	if err == nil && existing != nil {
		http.Error(w, fmt.Sprintf("host '%s' already exists", alias), http.StatusConflict)
		return
	}

	host := model.HostEntry{
		Alias:        alias,
		HostName:     hostName,
		User:         r.FormValue("user"),
		Port:         r.FormValue("port"),
		IdentityFile: r.FormValue("identityfile"),
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
		keys, _ := ssh.ListKeys(c.Dir)
		view.ConfigHostsTable(hosts, keys, c.Dir).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
