package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Store struct{ root string }

func New(root string) Store { return Store{root: root} }
func (s Store) Root(tenantID string) (string, error) {
	id, err := uuid.Parse(tenantID)
	if err != nil {
		return "", fmt.Errorf("invalid tenant id")
	}
	root := filepath.Join(s.root, "DataPonpes", id.String())
	if err := os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	return root, nil
}
func safeName(name string) error {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || filepath.IsAbs(name) || strings.ContainsAny(name, `/\\`) {
		return fmt.Errorf("invalid filename")
	}
	return nil
}
func (s Store) Put(tenantID, filename string, r io.Reader, max int64) (string, int64, string, error) {
	if err := safeName(filename); err != nil {
		return "", 0, "", err
	}
	root, err := s.Root(tenantID)
	if err != nil {
		return "", 0, "", err
	}
	tmp, err := os.CreateTemp(root, ".upload-*")
	if err != nil {
		return "", 0, "", err
	}
	defer os.Remove(tmp.Name())
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(r, max+1))
	if err != nil {
		_ = tmp.Close()
		return "", 0, "", err
	}
	if n > max {
		_ = tmp.Close()
		return "", 0, "", fmt.Errorf("file too large")
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return "", 0, "", err
	}
	if err := tmp.Close(); err != nil {
		return "", 0, "", err
	}
	key := uuid.NewString() + filepath.Ext(filename)
	if err := os.Rename(tmp.Name(), filepath.Join(root, key)); err != nil {
		return "", 0, "", err
	}
	return key, n, hex.EncodeToString(h.Sum(nil)), nil
}
func (s Store) Path(tenantID, key string) (string, error) {
	if err := safeName(key); err != nil {
		return "", err
	}
	root, err := s.Root(tenantID)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, key), nil
}

func (s Store) Open(tenantID, key string) (*os.File, error) {
	path, err := s.Path(tenantID, key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}
func (s Store) Delete(tenantID, key string) error {
	path, err := s.Path(tenantID, key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
