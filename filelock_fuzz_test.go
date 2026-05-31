package filelock_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/brandonkramer/filelock"
)

func FuzzAcquireRelease(f *testing.F) {
	f.Add("daemon")
	ctx := context.Background()
	f.Fuzz(func(t *testing.T, name string) {
		lockPath := filepath.Join(t.TempDir(), name+".lock")
		release, err := filelock.Acquire(ctx, lockPath)
		if err != nil {
			return
		}
		release()
	})
}
