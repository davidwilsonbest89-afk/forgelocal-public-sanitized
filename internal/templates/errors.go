package templates

import "errors"

var (
	ErrInvalidTemplate     = errors.New("template is invalid")
	ErrNotFound            = errors.New("template not found")
	ErrNameActive          = errors.New("an active template already uses this name")
	ErrVersionNotActive    = errors.New("template version is not active")
	ErrStaleVersion        = errors.New("template active version is stale")
	ErrConflict            = errors.New("template draft conflicts with base draft")
	ErrInvalidTemplateID   = errors.New("template id is invalid")
	ErrInvalidVersion      = errors.New("template version is invalid")
	ErrInvalidTemplateName = errors.New("template name is invalid")
)
