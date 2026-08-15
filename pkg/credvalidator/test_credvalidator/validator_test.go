package credvalidator_test

import (
	"errors"
	"pkg/credvalidator"
	"strings"
	"testing"
)

func TestHasSpec(t *testing.T) {
	tests := []struct {
		name string
		pwd  string
		want error
	}{
		{"has special char", "abc@def", nil},
		{"no special char", "abcdef123", credvalidator.ErrPasswordNoSpecial},
		{"only special chars", "!@#$", nil},
		{"empty string", "", credvalidator.ErrPasswordNoSpecial},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := credvalidator.HasSpec(tt.pwd); got != tt.want {
				t.Errorf("HasSpec(%q) = %v, want %v", tt.pwd, got, tt.want)
			}
		})
	}
}

func TestIsShort(t *testing.T) {
	t.Run("returns error func when minLen below 8", func(t *testing.T) {
		fn := credvalidator.IsShort(7)
		if err := fn("anything, doesn't matter"); !errors.Is(err, credvalidator.ErrMinLenForFuncIsTooShort) {
			t.Errorf("IsLong(7)(...) = %v, want %v", err, credvalidator.ErrMinLenForFuncIsTooShort)
		}
	})

	t.Run("valid minLen enforces length", func(t *testing.T) {
		tests := []struct {
			name   string
			minLen int
			pwd    string
			want   error
		}{
			{"too short", 8, "short1@", credvalidator.ErrPasswordIsTooShort},
			{"exactly minLen", 8, "exactly8", nil},
			{"longer than minLen", 8, "wayyy longer than needed", nil},
			{"custom minLen too short", 12, "short1234@", credvalidator.ErrPasswordIsTooShort},
			{"custom minLen satisfied", 12, "longenough12", nil},
			{"empty string", 8, "", credvalidator.ErrPasswordIsTooShort},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fn := credvalidator.IsShort(tt.minLen)
				if got := fn(tt.pwd); got != tt.want {
					t.Errorf("IsLong(%d)(%q) = %v, want %v", tt.minLen, tt.pwd, got, tt.want)
				}
			})
		}
	})
}

func TestIsUpper(t *testing.T) {
	tests := []struct {
		name string
		pwd  string
		want error
	}{
		{"has upper", "abcDef", nil},
		{"no upper", "abcdef", credvalidator.ErrPasswordNoUpper},
		{"only upper", "ABC", nil},
		{"empty string", "", credvalidator.ErrPasswordNoUpper},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := credvalidator.IsUpper(tt.pwd); got != tt.want {
				t.Errorf("IsUpper(%q) = %v, want %v", tt.pwd, got, tt.want)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name string
		pwd  string
		want error
	}{
		{"has number", "abc1def", nil},
		{"no number", "abcdef", credvalidator.ErrPasswordNoNumber},
		{"only numbers", "12345", nil},
		{"empty string", "", credvalidator.ErrPasswordNoNumber},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := credvalidator.IsNumber(tt.pwd); got != tt.want {
				t.Errorf("IsNumber(%q) = %v, want %v", tt.pwd, got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{"valid email", "user@example.com", nil},
		{"missing at sign", "userexample.com", credvalidator.ErrEmailNotValid},
		{"missing domain", "user@", credvalidator.ErrEmailNotValid},
		{"empty string", "", credvalidator.ErrEmailNotValid},
		{"valid with name", "John Doe <john@example.com>", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := credvalidator.ValidateEmail(tt.email); got != tt.wantErr {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.wantErr)
			}
		})
	}
}

func TestCreatePasswordPolicies(t *testing.T) {
	alwaysFail := func(string) error { return errors.New("fail") }
	alwaysPass := func(string) error { return nil }

	t.Run("stores every provided policy", func(t *testing.T) {
		p := credvalidator.CreatePasswordPolicies(alwaysPass, alwaysFail, alwaysPass)
		if len(p.GetPolicices()) != 3 {
			t.Fatalf("expected 3 policies, got %d", len(p.GetPolicices()))
		}
	})

	t.Run("no policies given", func(t *testing.T) {
		p := credvalidator.CreatePasswordPolicies()
		if len(p.GetPolicices()) != 0 {
			t.Fatalf("expected 0 policies, got %d", len(p.GetPolicices()))
		}
	})

	t.Run("policies are actually usable after creation", func(t *testing.T) {
		p := credvalidator.CreatePasswordPolicies(alwaysFail)
		errs := p.ApplyPolicies("irrelevant")
		if len(errs) != 1 {
			t.Fatalf("expected 1 error from a single always-failing policy, got %d: %v", len(errs), errs)
		}
	})
}

func TestApplyPolicies(t *testing.T) {
	alwaysFail := func(string) error { return errors.New("fail") }
	alwaysPass := func(string) error { return nil }

	tests := []struct {
		name     string
		policies []func(string) error
		wantLen  int
	}{
		{"no policies", nil, 0},
		{"all pass", []func(string) error{alwaysPass, alwaysPass}, 0},
		{"all fail", []func(string) error{alwaysFail, alwaysFail}, 2},
		{"mixed", []func(string) error{alwaysPass, alwaysFail, alwaysPass, alwaysFail}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := credvalidator.CreatePasswordPolicies(tt.policies...)
			errs := p.ApplyPolicies("irrelevant")
			if len(errs) != tt.wantLen {
				t.Errorf("ApplyPolicies() returned %d errors, want %d", len(errs), tt.wantLen)
			}
		})
	}
}

func TestDefaultPasswordPolicy(t *testing.T) {
	tests := []struct {
		name     string
		pwd      string
		wantErrs []error
	}{
		{
			name:     "valid password",
			pwd:      "Abcdef1@",
			wantErrs: nil,
		},
		{
			name: "missing uppercase",
			pwd:  "abcdef1@",
			wantErrs: []error{
				credvalidator.ErrPasswordNoUpper,
			},
		},
		{
			name: "missing special character",
			pwd:  "Abcdef12",
			wantErrs: []error{
				credvalidator.ErrPasswordNoSpecial,
			},
		},
		{
			name: "missing number",
			pwd:  "Abcdefg@",
			wantErrs: []error{
				credvalidator.ErrPasswordNoNumber,
			},
		},
		{
			name: "too short",
			pwd:  "Ab1@",
			wantErrs: []error{
				credvalidator.ErrPasswordIsTooShort,
			},
		},
		{
			name: "too long",
			pwd:  strings.Repeat("A", 127) + "1@",
			wantErrs: []error{
				credvalidator.ErrPasswordIsTooLong,
			},
		},
		{
			name:     "exactly maximum length",
			pwd:      strings.Repeat("A", 126) + "1@",
			wantErrs: nil,
		},
		{
			name: "empty password",
			pwd:  "",
			wantErrs: []error{
				credvalidator.ErrPasswordIsEmpty,
				credvalidator.ErrPasswordIsTooShort,
				credvalidator.ErrPasswordNoSpecial,
				credvalidator.ErrPasswordNoNumber,
				credvalidator.ErrPasswordNoUpper,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := credvalidator.DefaultPasswordPolicy.ApplyPolicies(tt.pwd)

			if len(errs) != len(tt.wantErrs) {
				t.Fatalf(
					"ApplyPolicies(%q) returned %d errors, want %d: %v",
					tt.pwd,
					len(errs),
					len(tt.wantErrs),
					errs,
				)
			}

			for i, want := range tt.wantErrs {
				if !errors.Is(errs[i], want) {
					t.Errorf(
						"ApplyPolicies(%q) error[%d] = %v, want %v",
						tt.pwd,
						i,
						errs[i],
						want,
					)
				}
			}
		})
	}
}
func TestIsLong(t *testing.T) {
	t.Run("returns error func when maxLen above 128", func(t *testing.T) {
		fn := credvalidator.IsLong(129)

		if err := fn("anything"); !errors.Is(err, credvalidator.ErrMaxLenForFuncIsTooLong) {
			t.Errorf(
				"IsLong(129)(...) = %v, want %v",
				err,
				credvalidator.ErrMaxLenForFuncIsTooLong,
			)
		}
	})
	t.Run("valid maxLen enforces length", func(t *testing.T) {
		tests := []struct {
			name   string
			maxLen int
			pwd    string
			want   error
		}{
			{"too long", 8, "123456789", credvalidator.ErrPasswordIsTooLong},
			{"exactly maxLen", 8, "12345678", nil},
			{"shorter than maxLen", 8, "1234567", nil},
			{"empty string", 8, "", nil},
			{"custom maxLen exceeded", 12, "1234567890123", credvalidator.ErrPasswordIsTooLong},
			{"custom maxLen satisfied", 12, "123456789012", nil},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				fn := credvalidator.IsLong(tt.maxLen)

				if got := fn(tt.pwd); got != tt.want {
					t.Errorf(
						"IsLong(%d)(%q) = %v, want %v",
						tt.maxLen,
						tt.pwd,
						got,
						tt.want,
					)
				}
			})
		}
	})
}
