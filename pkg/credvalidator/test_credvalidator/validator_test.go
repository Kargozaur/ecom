package credvalidator_test

import (
	"pkg/credvalidator"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "Valid standard email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "Valid email with display name",
			email:   "John Doe <john@example.com>",
			wantErr: false,
		},
		{
			name:    "Invalid: missing @ symbol",
			email:   "userexample.com",
			wantErr: true,
		},
		{
			name:    "Invalid: missing domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "Invalid: empty string",
			email:   "",
			wantErr: true,
		},
		{
			name:    "Invalid: spaces only",
			email:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := credvalidator.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name         string
		password     string
		wantErrCount int
		wantErrMsgs  []string
	}{
		{
			name:         "Valid strong password",
			password:     "P@ssword1",
			wantErrCount: 0,
		},
		{
			name:         "Missing upper character",
			password:     "p@ssword1",
			wantErrCount: 1,
			wantErrMsgs:  []string{"Password must contain at least 1 upper character."},
		},
		{
			name:         "Missing special character",
			password:     "Password123",
			wantErrCount: 1,
			wantErrMsgs:  []string{"Password must contain at least 1 speacial character."},
		},
		{
			name:         "Too short (less than 8 chars)",
			password:     "P@ss1",
			wantErrCount: 1,
			wantErrMsgs:  []string{"Password must be at least 8 characters long."},
		},
		{
			name:         "Missing number",
			password:     "P@ssword",
			wantErrCount: 1,
			wantErrMsgs:  []string{"Password must contain at least 1 number."},
		},
		{
			name:         "Missing all requirements",
			password:     "",
			wantErrCount: 4,
			wantErrMsgs: []string{
				"Password must contain at least 1 upper character.",
				"Password must contain at least 1 speacial character.",
				"Password must be at least 8 characters long.",
				"Password must contain at least 1 number.",
			},
		},
		{
			name:         "Valid with boundary length",
			password:     "A1!bcdef",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := credvalidator.ValidatePassword(tt.password)

			if len(errs) != tt.wantErrCount {
				t.Fatalf("ValidatePassword(%q) returned %d errors, want %d. Errors: %v",
					tt.password, len(errs), tt.wantErrCount, errs)
			}

			for _, wantMsg := range tt.wantErrMsgs {
				found := false
				for _, err := range errs {
					if err.Error() == wantMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ValidatePassword(%q) missing expected error message: %q", tt.password, wantMsg)
				}
			}
		})
	}
}
