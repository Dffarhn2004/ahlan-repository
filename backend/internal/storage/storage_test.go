package storage

import (
	"strings"
	"testing"
)

func TestPutRejectsTraversal(t *testing.T) {
	s := New(t.TempDir())
	if _, _, _, err := s.Put("00000000-0000-0000-0000-000000000001", "../secret.txt", strings.NewReader("x"), 20); err == nil {
		t.Fatal("expected traversal rejection")
	}
}
func TestPutStoresChecksum(t *testing.T) {
	s := New(t.TempDir())
	key, size, sum, err := s.Put("00000000-0000-0000-0000-000000000001", "a.txt", strings.NewReader("hello"), 20)
	if err != nil || key == "" || size != 5 || sum != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("unexpected result: %s %d %s %v", key, size, sum, err)
	}
}
