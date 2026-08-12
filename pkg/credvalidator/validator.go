package credvalidator

import (
	"errors"
	"net/mail"
	"unicode"
)

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

func ValidatePassword(pwd string) []error {
	errs := make([]error, 0, 4)
	var isUpper, isSpecial, isLong, isNumber bool
	if len(pwd) >= 8 {
		isLong = true
	}
	for _, ch := range pwd {
		switch true {
		case unicode.IsUpper(ch):
			isUpper = true
			continue
		case spec(ch):
			isSpecial = true
			continue
		case unicode.IsNumber(ch):
			isNumber = true
			continue
		}
	}
	if !isUpper {
		errs = append(errs, errors.New("Password must contain at least 1 upper character."))
	}
	if !isSpecial {
		errs = append(errs, errors.New("Password must contain at least 1 speacial character."))
	}
	if !isLong {
		errs = append(errs, errors.New("Password must be at least 8 characters long."))
	}
	if !isNumber {
		errs = append(errs, errors.New("Password must contain at least 1 number."))
	}
	return errs
}

func spec(ch rune) bool {
	switch ch {
	case '`', '@', '!', '#', '$', '%', '^', '&', '*', '(', ')', '_', '-', '=', '+', ';', ':', '.', ',', '"', '>', '<', '/':
		return true
	default:
		return false
	}
}
