package multivalueflag

import (
	"slices"
	"testing"
)

func createMultiValueFlag() MultiValueFlag[string] {
	return MultiValueFlag[string]{
		FlagName:      "some name",
		DefaultValues: []string{"value1", "value2", "value3"},
	}
}

func TestMultiValueFlagString(t *testing.T) {
	t.Parallel()
	mvf := createMultiValueFlag()

	defaultValue := mvf.String()

	expectedValue := "value1, value2, value3"

	if defaultValue != expectedValue {
		t.Errorf("expected '%s', got '%s'", expectedValue, defaultValue)
	}
}

func TestMultiValueFlagSet(t *testing.T) {

	t.Run("test valid values", func(t *testing.T) {
		tests := []struct {
			name         string
			rawValues    string
			parsedValues []string
		}{
			{
				name:         "accept valid values",
				rawValues:    "value4, value5",
				parsedValues: []string{"value4", "value5"},
			},
			{
				name:         "reset default via a multi value",
				rawValues:    "value4,all",
				parsedValues: nil,
			},
			{
				name:         "reset default via 'all'",
				rawValues:    "all",
				parsedValues: nil,
			},
		}

		for _, test := range tests {

			t.Run(test.name, func(t *testing.T) {
				mvf := createMultiValueFlag()
				err := mvf.Set(test.rawValues)

				if err != nil {
					t.Fatalf("expected err to be nil, got %v", err)
				}

				if !slices.Equal(mvf.newValues, test.parsedValues) {
					t.Fatalf("expected %v to be equal to %v", mvf.newValues, test.parsedValues)
				}

				defaultValues := createMultiValueFlag().DefaultValues

				if !slices.Equal(mvf.DefaultValues, defaultValues) {
					t.Fatalf("expected default values %v to not be changed, but got: %v", defaultValues, mvf.DefaultValues)
				}
			})
		}
	})

	t.Run("test invalid values", func(t *testing.T) {
		tests := []struct {
			name      string
			rawValues string
		}{
			{
				name:      "reject malformed values",
				rawValues: "value4, value5,",
			},
			{
				name:      "reject malformed reset",
				rawValues: "all,",
			},
		}

		for _, test := range tests {

			t.Run(test.name, func(t *testing.T) {
				mvf := createMultiValueFlag()
				err := mvf.Set(test.rawValues)

				if err == nil {
					t.Fatal("expected err to not be nil")
				}

				defaultValues := createMultiValueFlag().DefaultValues

				if !slices.Equal(mvf.DefaultValues, defaultValues) {
					t.Fatalf("expected default values %v to not be changed, but got: %v", defaultValues, mvf.DefaultValues)
				}

			})
		}
	})
}
