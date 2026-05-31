package filelock

import (
	"context"
	"errors"
	"fmt"
	"time"
)

//
// ────────────────────────────────────────
// acquire orchestration.
//

const defaultAcquireRetry = 10 * time.Millisecond

var lockNewTicker = time.NewTicker

func acquireLock(ctx context.Context, lockPath string, cfg config) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("filelock: acquire %s: %w", lockPath, err)
	}
	if !cfg.blocking {
		return acquireOnce(lockPath, cfg)
	}
	try := cfg
	try.blocking = false
	ticker := lockNewTicker(defaultAcquireRetry)
	defer ticker.Stop()
	for {
		release, err := acquireOnce(lockPath, try)
		if err == nil {
			return release, nil
		}
		if !errors.Is(err, ErrBusy) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("filelock: acquire %s: %w", lockPath, ctx.Err())
		case <-ticker.C:
		}
	}
}
