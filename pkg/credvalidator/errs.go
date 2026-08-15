package credvalidator

import "errors"

var (
	ErrPasswordIsTooShort      = errors.New("Password is too short")
	ErrPasswordNoSpecial       = errors.New("Password must contain at least 1 special character.")
	ErrPasswordNoUpper         = errors.New("Password must contain at least 1 upper character.")
	ErrPasswordNoNumber        = errors.New("Password must contain at least 1 number.")
	ErrPasswordIsEmpty         = errors.New("Password can not be empty.")
	ErrPasswordIsTooLong       = errors.New("Password is too long.")
	ErrMinLenForFuncIsTooShort = errors.New("Min len for password for a given func must be at least 8 characters")
	ErrMaxLenForFuncIsTooLong  = errors.New("Max len for password for a given func is 128 characters")
	ErrEmailNotValid           = errors.New("Email is not valid.")
)
