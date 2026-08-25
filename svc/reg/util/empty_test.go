package util_test

import (
	"reg/util"
	"testing"
)

func TestAnyBlank(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want bool
	}{
		{"no args", nil, false},
		{"empty slice", []string{}, false},
		{"single non-empty", []string{"a"}, false},
		{"single empty", []string{""}, true},
		{"single whitespace", []string{"   "}, true},
		{"tab and newline", []string{"\t\n"}, true},
		{"all non-empty", []string{"a", "b", "c"}, false},
		{"one empty among many", []string{"a", "", "c"}, true},
		{"one whitespace among many", []string{"a", "   ", "c"}, true},
		{"empty in the middle", []string{"x", "y", "", "z"}, true},
		{"trailing/leading spaces but not blank", []string{"  a  "}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := util.AnyBlank(tt.in...)
			if got != tt.want {
				t.Errorf("AnyBlank(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
