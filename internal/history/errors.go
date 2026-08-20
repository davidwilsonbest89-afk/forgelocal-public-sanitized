package history

import "errors"

var (
	ErrNotFound       = errors.New("profile history not found")
	ErrInvalidVersion = errors.New("invalid history version")
	ErrVersionConflict = errors.New("history version conflict")
	ErrDifferentProfile = errors.New("history versions belong to different profiles")
	ErrInvalidSnapshot = errors.New("invalid history snapshot")
)
