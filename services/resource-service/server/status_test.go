package server

import "testing"

// TestValidateTransition_Valid verifies that all allowed status transitions pass.
func TestValidateTransition_Valid(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
	}{
		{"available to allocated", StatusAvailable, StatusAllocated},
		{"allocated to in_production", StatusAllocated, StatusInProduction},
		{"allocated to available", StatusAllocated, StatusAvailable},
		{"in_production to faulted", StatusInProduction, StatusFaulted},
		{"in_production to available", StatusInProduction, StatusAvailable},
		{"faulted to available", StatusFaulted, StatusAvailable},
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
		{"available to in_production", StatusAvailable, StatusInProduction},
		{"available to faulted", StatusAvailable, StatusFaulted},
		{"faulted to allocated", StatusFaulted, StatusAllocated},
		{"faulted to in_production", StatusFaulted, StatusInProduction},
		{"in_production to allocated", StatusInProduction, StatusAllocated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateTransition(tt.from, tt.to); err == nil {
				t.Errorf("ValidateTransition(%s, %s) = nil, want error", tt.from, tt.to)
			}
		})
	}
}
