package credvalidator_test

import (
	"errors"
	"pkg/credvalidator"
	"slices"
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

func TestIsLong(t *testing.T) {
	t.Run("returns error func when minLen below 8", func(t *testing.T) {
		fn := credvalidator.IsLong(7)
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
				fn := credvalidator.IsLong(tt.minLen)
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

	// Regression test: CreatePasswordPolicies used to build its slice with
	// make([]func(string) error, 0, len(f)) and then copy(funcs, f), which
	// copies zero elements because len(funcs) == 0. That silently produced
	// an empty policy set no matter how many funcs were passed in.
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
		{"valid password", "Abcdef1@", nil},
		{"missing upper", "abcdef1@", []error{credvalidator.ErrPasswordNoUpper}},
		{"missing special", "Abcdef12", []error{credvalidator.ErrPasswordNoSpecial}},
		{"missing number", "Abcdefg@", []error{credvalidator.ErrPasswordNoNumber}},
		{"too short", "Ab1@", []error{credvalidator.ErrPasswordIsTooShort}},
		{
			"empty string",
			"",
			[]error{credvalidator.ErrPasswordNoSpecial, credvalidator.ErrPasswordIsTooShort,
				credvalidator.ErrPasswordNoNumber, credvalidator.ErrPasswordNoUpper},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := credvalidator.DefaultPasswordPolicy.ApplyPolicies(tt.pwd)
			if len(errs) != len(tt.wantErrs) {
				t.Fatalf("ApplyPolicies(%q) returned %d errors, want %d (errs: %v)",
					tt.pwd, len(errs), len(tt.wantErrs), errs)
			}
			for _, want := range tt.wantErrs {
				found := slices.Contains(errs, want)
				if !found {
					t.Errorf("ApplyPolicies(%q) missing expected error %v, got: %v", tt.pwd, want, errs)
				}
			}
		})
	}
}
