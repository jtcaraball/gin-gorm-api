package api

import "errors"

// ErrMissingForm is used to signal that a handler could not find an expected
// form in a request's context.
var ErrMissingForm = errors.New("form missing from context")
