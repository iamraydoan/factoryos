package server

import "testing"

// TestValidateTransition_Valid verifies that all allowed status transitions pass.
func TestValidateTransition_Valid(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{"available to allocated", "available", "allocated"},
		{"allocated to in_production", "allocated", "in_production"},
		{"allocated to available", "allocated", "available"},
		{"in_production to faulted", "in_production", "faulted"},
		{"in_production to available", "in_production", "available"},
		{"faulted to available", "faulted", "available"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransition(tt.from, tt.to); err != nil {
				t.Errorf("ValidateTransition(%s, %s) = %v, want nil", tt.from, tt.to, err)
			}
		})
	}
}

// TestValidateTransition_Invalid verifies that disallowed transitions return an error.
func TestValidateTransition_Invalid(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{"available to in_production", "available", "in_production"},
		{"available to faulted", "available", "faulted"},
		{"faulted to allocated", "faulted", "allocated"},
		{"faulted to in_production", "faulted", "in_production"},
		{"in_production to allocated", "in_production", "allocated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransition(tt.from, tt.to); err == nil {
				t.Errorf("ValidateTransition(%s, %s) = nil, want error", tt.from, tt.to)
			}
		})
	}
}
