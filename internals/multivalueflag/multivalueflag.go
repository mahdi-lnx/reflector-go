// multivalueflag pacakge implements a cli flag that can have multiple values
package multivalueflag

import (
	"fmt"
	"strings"
)

// MultiValueFlag defines a flag that can have multiple values
type MultiValueFlag[T ~string] struct {
	// DefaultValues defines the the default values of the flag
	//
	// DefaultValues will never be mutated
	DefaultValues []T

	// newValues defines the parsed flag values provided by the user
	newValues []T

	// FlagName defines the flag's name
	FlagName string

	// defaultValueDiscarded is an internal state flag to indicate weather
	// defaultValues should be discarded
	defaultValueDiscarded bool
}

// String returns the default value of the MultiValueFlag as
// a comma seperated value string
func (mvf *MultiValueFlag[T]) String() string {
	defaultValues := make([]string, len(mvf.DefaultValues))

	for i, value := range mvf.DefaultValues {
		defaultValues[i] = string(value)
	}

	return strings.Join(defaultValues, ", ")
}

// Set sets a new flag value
func (mvf *MultiValueFlag[T]) Set(arg string) error {
	arg = strings.TrimSpace(arg)

	// [strings.ToLower] is used here because the "enum" constants in [mirrorlist] are all lowercased
	arg = strings.ToLower(arg)

	// for this function we will set mvf.newValues to a nil slice
	// if we encounter a standalone (e.g., --flagName all) or encounter
	// a comma seperated value that contains all (e.g., --flagName someValue,all)

	if !mvf.defaultValueDiscarded {
		mvf.defaultValueDiscarded = true
	}

	if arg == "all" {
		// skip the loop
		mvf.newValues = nil
		return nil
	}

	for value := range strings.SplitSeq(arg, ",") {
		value = strings.TrimSpace(value)

		if value == "all" {
			// update newValues to an nil slice
			mvf.newValues = nil
			continue
		}

		if value == "" {
			return fmt.Errorf("malformed %s list", mvf.FlagName)
		}

		mvf.newValues = append(mvf.newValues, T(value))
	}

	return nil
}

// GetFinalFlagValues returns the final values for the MultiValueFlag
func (mvf *MultiValueFlag[T]) GetFinalFlagValues() []T {
	if mvf.defaultValueDiscarded {
		return mvf.newValues
	}

	return mvf.DefaultValues
}
