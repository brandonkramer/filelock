package filelock_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/brandonkramer/filelock"
)

func BenchmarkAcquireRelease(b *testing.B) {
	ctx := context.Background()
	lockPath := filepath.Join(b.TempDir(), "daemon.lock")
	for b.Loop() {
		release, err := filelock.Acquire(ctx, lockPath)
		if err != nil {
			b.Fatal(err)
		}
		release()
	}
}
