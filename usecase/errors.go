package usecase

import "errors"

// ErrImport is returned when an import file has an unrecoverable format error.
var ErrImport = errors.New("import error")
