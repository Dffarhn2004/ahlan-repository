package jobs

import (
	"testing"
	"time"
)

func TestDefaultLease(t *testing.T) {
	if (Repository{}).Lease != 0 {
		t.Fatal("zero repository should use runtime default")
	}
	if time.Minute <= 0 {
		t.Fatal("lease must be positive")
	}
}
