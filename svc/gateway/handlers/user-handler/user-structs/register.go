package userstructs

import "pkg/credvalidator"

type Register struct {
	Email    string `json:"email,case:ignore"`
	Password string `json:"password,case:ignore"`
	Name     string `json:"name,case:ignore"`
}

func (r *Register) ValidateData(policy credvalidator.PasswordPolicy) []error {
	errs := make([]error, 0, len(policy.GetPolicices()))
	emailErr := credvalidator.ValidateEmail(r.Email)
	if emailErr != nil {
		errs = append(errs, emailErr)
	}
	pwdErr := policy.ApplyPolicies(r.Password)
	if len(pwdErr) != 0 {
		errs = append(errs, pwdErr...)
	}
	return errs
}
