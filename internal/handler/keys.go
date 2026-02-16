package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/holden/sshmasher/internal/model"
	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/internal/view"
)

// Keys holds dependencies for key management handlers.
type Keys struct {
	Dir *ssh.SSHDir
}

func (k *Keys) List(w http.ResponseWriter, r *http.Request) {
	keys, err := ssh.ListKeys(k.Dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refCount, err := ssh.KeyRefCount(k.Dir)
	if err != nil {
		refCount = make(map[string]int)
	}

	search := r.URL.Query().Get("search")
	if search != "" {
		var filtered []model.SSHKey
		searchLower := strings.ToLower(search)
		for _, key := range keys {
			if strings.Contains(strings.ToLower(key.Name), searchLower) ||
				strings.Contains(strings.ToLower(key.Type), searchLower) ||
				strings.Contains(strings.ToLower(key.Comment), searchLower) ||
				strings.Contains(strings.ToLower(key.Fingerprint), searchLower) {
				filtered = append(filtered, key)
			}
		}
		keys = filtered
	}

	if isHTMX(r) {
		view.KeysTable(keys, refCount).Render(r.Context(), w)
		return
	}
	writeJSON(w, keys)
}

func (k *Keys) NewKey(w http.ResponseWriter, r *http.Request) {
	if isHTMX(r) {
		view.KeyGenerateForm().Render(r.Context(), w)
		return
	}
	writeJSON(w, nil)
}

func (k *Keys) Get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	key, err := ssh.GetKey(k.Dir, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if isHTMX(r) {
		view.KeyDetail(*key).Render(r.Context(), w)
		return
	}
	writeJSON(w, key)
}

func (k *Keys) Generate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	bits, _ := strconv.Atoi(r.FormValue("bits"))
	req := model.KeyGenRequest{
		Name:       r.FormValue("name"),
		Type:       r.FormValue("type"),
		Bits:       bits,
		Comment:    r.FormValue("comment"),
		Passphrase: r.FormValue("passphrase"),
	}

	if req.Name == "" || req.Type == "" {
		http.Error(w, "name and type are required", http.StatusBadRequest)
		return
	}

	if err := ssh.GenerateKey(k.Dir, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated key list
	keys, err := ssh.ListKeys(k.Dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refCount, err := ssh.KeyRefCount(k.Dir)
	if err != nil {
		refCount = make(map[string]int)
	}
	if isHTMX(r) {
		view.KeysTable(keys, refCount).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, keys)
}

func (k *Keys) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := ssh.DeleteKey(k.Dir, name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keys, err := ssh.ListKeys(k.Dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refCount, err := ssh.KeyRefCount(k.Dir)
	if err != nil {
		refCount = make(map[string]int)
	}
	if isHTMX(r) {
		view.KeysTable(keys, refCount).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (k *Keys) UpdateComment(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	newComment := r.FormValue("comment")

	if err := ssh.UpdateKeyComment(k.Dir, name, newComment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	keys, err := ssh.ListKeys(k.Dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	refCount, err := ssh.KeyRefCount(k.Dir)
	if err != nil {
		refCount = make(map[string]int)
	}
	if isHTMX(r) {
		view.KeysTable(keys, refCount).Render(r.Context(), w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
