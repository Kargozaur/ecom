package userstructs

import "pkg/credvalidator"

type Register struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (r *Register) ValidateData() []error {
	errs := make([]error, 0, 5)
	emailErr := credvalidator.ValidateEmail(r.Email)
	if emailErr != nil {
		errs = append(errs, emailErr)
	}
	pwdErr := credvalidator.ValidatePassword(r.Password)
	if len(pwdErr) != 0 {
		errs = append(errs, pwdErr...)
	}
	return errs
}
