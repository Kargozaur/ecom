package main

import "pkg/credvalidator"

func CreatePolicies(policies ...func(string) error) credvalidator.PasswordPolicy {
	if len(policies) == 0 {
		return credvalidator.DefaultPasswordPolicy
	} else {
		return credvalidator.CreatePasswordPolicies(policies...)
	}
}
