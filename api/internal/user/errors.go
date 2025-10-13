package user

import "errors"

// ErrNotFound indicates the requested user does not exist.
var ErrNotFound = errors.New("user not found")
