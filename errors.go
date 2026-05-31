package filelock

import "errors"

//
// ────────────────────────────────────────
// sentinel errors.
//

// ErrBusy is returned when a non-blocking lock attempt finds the file already locked.
var ErrBusy = errors.New("filelock: busy")
