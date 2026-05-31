// Package filelock provides cross-platform advisory file locks for single-instance
// services, registries, and other on-disk critical sections.
//
// Use [Acquire] for a held lock with an explicit release callback. By default
// acquisition is non-blocking; combine [Blocking] and [WritePID] for daemon locks.
// Pass a [context.Context] to cancel blocking waits.
//
// Use [With] or [WithSidecar] for scoped, callback-style exclusive access.
//
// # Examples
//
//	release, err := filelock.Acquire(context.Background(), "/var/run/myapp/daemon.lock", filelock.WritePID())
//	if err != nil {
//	    return err
//	}
//	defer release()
//
//	err = filelock.WithSidecar(context.Background(), "/var/lib/myapp/state.json", filelock.DefaultSidecar, func() error {
//	    return updateState()
//	})
package filelock
