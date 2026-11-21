package pathlock

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPerPathLockSerialisesWriters(t *testing.T) {
	key := filepath.Join(t.TempDir(), "secret")

	Lock(key)
	acquired := make(chan struct{})
	go func() {
		Lock(key)
		close(acquired)
		Unlock(key)
	}()

	select {
	case <-acquired:
		t.Fatalf("second writer acquired the lock while the first was held")
	case <-time.After(20 * time.Millisecond):
	}

	Unlock(key)
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatalf("second writer did not acquire the lock after release")
	}
}
