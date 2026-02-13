package model

import "time"

// SSHKey represents an SSH key pair found in the SSH directory.
type SSHKey struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // rsa, ed25519, ecdsa, dsa
	Bits        int       `json:"bits"`        // key size
	Fingerprint string    `json:"fingerprint"` // SHA256 fingerprint
	PublicKey   string    `json:"publicKey"`
	Comment     string    `json:"comment"`
	HasPrivate  bool      `json:"hasPrivate"`
	ModTime     time.Time `json:"modTime"`
}

// KeyGenRequest holds parameters for generating a new SSH key.
type KeyGenRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`       // rsa, ed25519, ecdsa
	Bits       int    `json:"bits"`       // key size (for rsa/ecdsa)
	Comment    string `json:"comment"`
	Passphrase string `json:"passphrase"`
}

// HostEntry represents a host block in ~/.ssh/config.
type HostEntry struct {
	Alias        string            `json:"alias"`
	HostName     string            `json:"hostName"`
	User         string            `json:"user"`
	Port         string            `json:"port"`
	IdentityFile string            `json:"identityFile"`
	Options      map[string]string `json:"options"` // all other key=value pairs
}

// KnownHostEntry represents a single line in known_hosts.
type KnownHostEntry struct {
	Line        int    `json:"line"`        // 1-based line number
	Hosts       string `json:"hosts"`       // hostname(s) or IP(s)
	KeyType     string `json:"keyType"`     // ssh-rsa, ssh-ed25519, etc.
	Key         string `json:"key"`         // base64-encoded public key
	Fingerprint string `json:"fingerprint"` // SHA256 fingerprint
	IsHashed    bool   `json:"isHashed"`    // whether hostnames are hashed
}

// Backup represents a tar.gz snapshot of ~/.ssh.
type Backup struct {
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}
