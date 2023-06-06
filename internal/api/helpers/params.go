package helpers

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

// CommaDelimitedOptionalValue takes an array of strings and returns a pointer
// to a string containing all values concatenated with commas, or nil if the array is empty.
func CommaDelimitedOptionalValue(vals []string) *string {
	if len(vals) == 0 {
		return nil
	}

	return aws.String(strings.Join(vals, ", "))
}
