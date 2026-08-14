package server

import "fmt"

// validTransitions defines the allowed Work Unit status state machine.
//
//	available → allocated → in_production → faulted → available
var validTransitions = map[string][]string{
	StatusAvailable:    {StatusAllocated},
	StatusAllocated:    {StatusInProduction, StatusAvailable},
	StatusInProduction: {StatusFaulted, StatusAvailable},
	StatusFaulted:      {StatusAvailable},
}

// ValidateTransition checks if a status transition is allowed.
// Returns nil if valid, or an error describing why the transition is invalid.
func ValidateTransition(from, to string) error {
	allowed, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("unknown status: %s", from)
	}

	for _, s := range allowed {
		if s == to {
			return nil
		}
	}

	return fmt.Errorf("invalid transition: %s → %s", from, to)
}
