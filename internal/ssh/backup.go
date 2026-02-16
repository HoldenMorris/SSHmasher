package ssh

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/holden/sshmasher/internal/model"
)

// ListBackups returns all backup files, sorted newest first.
func ListBackups(dir *SSHDir) ([]model.Backup, error) {
	backupDir := dir.BackupDir()
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}

	var backups []model.Backup
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		backups = append(backups, model.Backup{
			Filename:  entry.Name(),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// CreateBackup creates a tar.gz snapshot of the SSH directory.
func CreateBackup(dir *SSHDir) error {
	if err := dir.EnsureBackupDir(); err != nil {
		return err
	}

	filename := fmt.Sprintf("ssh-backup-%s.tar.gz", time.Now().Format("20060102-150405"))
	outPath := filepath.Join(dir.BackupDir(), filename)

	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create backup file: %w", err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(dir.Base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(dir.Base, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tw, f)
		return err
	})
}

// RestoreBackup extracts a tar.gz backup into the SSH directory.
func RestoreBackup(dir *SSHDir, filename string) error {
	backupPath := filepath.Join(dir.BackupDir(), filename)

	// Validate filename to prevent path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return fmt.Errorf("invalid backup filename")
	}

	f, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Clear existing SSH directory contents (but not the dir itself)
	entries, _ := os.ReadDir(dir.Base)
	for _, entry := range entries {
		os.RemoveAll(filepath.Join(dir.Base, entry.Name()))
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		target := filepath.Join(dir.Base, header.Name)

		// Prevent path traversal
		if !strings.HasPrefix(target, dir.Base) {
			return fmt.Errorf("invalid path in backup: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// DeleteBackup removes a backup file.
func DeleteBackup(dir *SSHDir, filename string) error {
	// Validate filename
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return fmt.Errorf("invalid backup filename")
	}
	return os.Remove(filepath.Join(dir.BackupDir(), filename))
}

// GetBackupPath returns the full path to a backup file.
func GetBackupPath(dir *SSHDir, filename string) (string, error) {
	// Validate filename to prevent path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return "", fmt.Errorf("invalid backup filename")
	}
	backupPath := filepath.Join(dir.BackupDir(), filename)
	// Check if file exists
	if _, err := os.Stat(backupPath); err != nil {
		return "", fmt.Errorf("backup not found: %s", filename)
	}
	return backupPath, nil
}
