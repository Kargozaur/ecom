package main

import "pkg/credvalidator"

func CreatePasswordPolicies(policies ...func(string) error) credvalidator.PasswordPolicy {
	if len(policies) == 0 {
		return credvalidator.DefaultPasswordPolicy
	} else {
		return credvalidator.CreatePasswordPolicies(policies...)
	}
}

func CreateGRPCRetryPolicy() string {
	return `{
		"methodConfig": [{
			"name": [{"service": ""}],
			"retryPolicy": {
				"maxAttempts": 4,
				"initialBackoff": "1s",
				"maxBackoff": "10s",
				"backoffMultiplier": 2.0,
				"retryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
			}
		}]
	}`
}
