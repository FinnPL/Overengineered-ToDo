package todo

import "errors"

// ErrNotFound indicates the requested todo could not be located.
var ErrNotFound = errors.New("todo not found")
