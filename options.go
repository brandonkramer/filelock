package filelock

import "os"

//
// ────────────────────────────────────────
// lock options.
//

// DefaultFileMode is the mode used when creating lock files.
const DefaultFileMode = 0o644

// DefaultSidecar is appended to a base path by [WithSidecar].
const DefaultSidecar = ".lock"

type config struct {
	blocking bool
	writePID bool
	fileMode os.FileMode
}

func defaultConfig() config {
	return config{fileMode: DefaultFileMode}
}

// Option configures [Acquire] and [With].
type Option func(*config)

// NonBlocking makes [Acquire] return [ErrBusy] instead of waiting when the lock is held.
// [With] is blocking by default; pass [NonBlocking] to try once and return [ErrBusy].
func NonBlocking() Option {
	return func(c *config) { c.blocking = false }
}

// Blocking makes [Acquire] wait until the lock is available.
func Blocking() Option {
	return func(c *config) { c.blocking = true }
}

// WritePID records the current process id in the lock file after acquisition.
func WritePID() Option {
	return func(c *config) { c.writePID = true }
}

// FileMode sets the permission bits used when creating the lock file.
func FileMode(mode os.FileMode) Option {
	return func(c *config) { c.fileMode = mode }
}
