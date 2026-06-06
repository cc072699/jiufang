package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertEqualOrder(t *testing.T) {
	// Test the exact same pattern as in sql_validator_test.go
	wantHasErrors := true
	hasErrors := false // This should fail

	assert.Equal(t, wantHasErrors, hasErrors, "This should show: expected: true, actual: false")
}

func TestAssertEqualOrderReverse(t *testing.T) {
	// Test the reverse pattern
	wantHasErrors := false
	hasErrors := true // This should fail

	assert.Equal(t, wantHasErrors, hasErrors, "This should show: expected: false, actual: true")
}