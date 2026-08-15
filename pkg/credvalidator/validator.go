package credvalidator

import (
	"errors"
	"net/mail"
	"unicode"
)

type PasswordPolicy struct {
	policies []func(string) error
}

var DefaultPasswordPolicy = CreatePasswordPolicies(IsEmpty, IsLong(pwdmaxLength), IsShort(pwdminLength), HasSpec, IsNumber, IsUpper)

const pwdmaxLength = 128
const pwdminLength = 8

func CreatePasswordPolicies(f ...func(string) error) PasswordPolicy {
	funcs := make([]func(string) error, 0, len(f))
	funcs = append(funcs, f...)
	return PasswordPolicy{policies: funcs}
}

func (p PasswordPolicy) GetPolicices() []func(string) error {
	return p.policies
}
func (p PasswordPolicy) ApplyPolicies(pwd string) []error {
	l := len(p.policies)
	if l == 0 {
		return nil
	}
	errs := make([]error, 0, l)
	for i := range l {
		if err := p.policies[i](pwd); err != nil {
			if errors.Is(err, ErrPasswordIsTooLong) {
				errs = append(errs, err)
				break
			}
			errs = append(errs, err)
		}
	}
	return errs
}

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrEmailNotValid
	}
	return nil
}

func HasSpec(pwd string) error {
	var found bool
	for _, ch := range pwd {
		switch ch {
		case '`', '@', '!', '#', '$', '%', '^', '&', '*', '(', ')', '_', '-', '=', '+', ';', ':', '.', ',', '"', '>', '<', '/':
			found = true
		}
		if found {
			break
		}
	}
	if !found {
		return ErrPasswordNoSpecial
	}
	return nil
}

// Policy enforces passwords to be at least 8 characters long.
// If given minLen will be less than 8
// function returns ErrMinLenForFuncsIsTooShort.
// Otherwise ErrPasswordIsTooShort or nil.
func IsShort(minLen int) func(string) error {
	if minLen < pwdminLength {
		return func(string) error { return ErrMinLenForFuncIsTooShort }
	}
	return func(pwd string) error {
		if len(pwd) < minLen {
			return ErrPasswordIsTooShort
		}
		return nil
	}
}

func IsLong(maxLen int) func(string) error {
	if maxLen > pwdmaxLength {
		return func(string) error { return ErrMaxLenForFuncIsTooLong }
	}
	return func(pwd string) error {
		if len(pwd) > maxLen {
			return ErrPasswordIsTooLong
		}
		return nil
	}
}

func IsUpper(pwd string) error {
	var found bool
	for _, ch := range pwd {
		if unicode.IsUpper(ch) {
			found = true
			break
		}
	}
	if !found {
		return ErrPasswordNoUpper
	}
	return nil
}

func IsNumber(pwd string) error {
	var found bool
	for _, ch := range pwd {
		if unicode.IsNumber(ch) {
			found = true
			break
		}
	}
	if !found {
		return ErrPasswordNoNumber
	}
	return nil
}

func IsEmpty(pwd string) error {
	if pwd == "" {
		return ErrPasswordIsEmpty
	}
	return nil
}
