package domain

import "errors"

var (
	ErrRequired         = errors.New("required value")
	ErrNotFound         = errors.New("not found")
	ErrNil              = errors.New("nil data")
	ErrNegative         = errors.New("negative value")
	ErrInvalidUserID    = errors.New("invalid user ID")
	ErrInvalidBookIDs   = errors.New("invalid book IDs")
	ErrNoUserInContext  = errors.New("no user in context")
	ErrMissingMetadata  = errors.New("missing grpc metadata")
	ErrMissingUserID    = errors.New("missing user-id in metadata")
	ErrInvalidUserEmail = errors.New("invalid user-email in metadata")
)
