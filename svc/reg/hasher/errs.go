package hasher

import "errors"

var (
	ErrInitFuncParams        = errors.New("not enough parameters were provided for init function")
	ErrFailedToGenerateBytes = errors.New("faild to generate bytes")
	ErrInvalidHash           = errors.New("invalid hash")
)
