package utils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// CustomError is a custom error type for testing
type CustomError struct {
	msg string
}

func (e *CustomError) Error() string {
	return e.msg
}

// AnotherError is another custom error type for testing
type AnotherError struct {
	code int
}

func (e *AnotherError) Error() string {
	return fmt.Sprintf("another error: %d", e.code)
}

func TestMapErr_WithMatchingCustomError(t *testing.T) {
	originalErr := &CustomError{msg: "test error"}
	target, ok := MapErr[*CustomError](originalErr)

	assert.True(t, ok, "MapErr should return true when error matches")
	assert.NotNil(t, target, "target should not be nil")
	assert.Equal(t, originalErr, target, "target should be the same as original error")
	assert.Equal(t, "test error", target.msg, "error message should be preserved")
}

func TestMapErr_WithNonMatchingCustomError(t *testing.T) {
	originalErr := &AnotherError{code: 42}
	target, ok := MapErr[*CustomError](originalErr)

	assert.False(t, ok, "MapErr should return false when error doesn't match")
	assert.Nil(t, target, "target should be nil when error doesn't match")
}

func TestMapErr_WithWrappedError(t *testing.T) {
	originalErr := &CustomError{msg: "wrapped error"}
	wrappedErr := fmt.Errorf("wrapping: %w", originalErr)

	target, ok := MapErr[*CustomError](wrappedErr)

	assert.True(t, ok, "MapErr should unwrap and find the error")
	assert.NotNil(t, target, "target should not be nil")
	assert.Equal(t, originalErr, target, "target should be the original error")
	assert.Equal(t, "wrapped error", target.msg, "error message should be accessible")
}

func TestMapErr_WithNilError(t *testing.T) {
	target, ok := MapErr[*CustomError](nil)

	assert.False(t, ok, "MapErr should return false for nil error")
	assert.Nil(t, target, "target should be nil for nil error")
}

func TestMapErr_WithStandardError(t *testing.T) {
	standardErr := errors.New("standard error")
	target, ok := MapErr[*CustomError](standardErr)

	assert.False(t, ok, "MapErr should return false for standard error without custom type")
	assert.Nil(t, target, "target should be nil")
}

func TestMapErr_WithMultipleLevelsOfWrapping(t *testing.T) {
	originalErr := &CustomError{msg: "deep error"}
	level1 := fmt.Errorf("level 1: %w", originalErr)
	level2 := fmt.Errorf("level 2: %w", level1)
	level3 := fmt.Errorf("level 3: %w", level2)

	target, ok := MapErr[*CustomError](level3)

	assert.True(t, ok, "MapErr should find error in deep wrapper chain")
	assert.NotNil(t, target, "target should not be nil")
	assert.Equal(t, originalErr, target, "target should be the original error")
	assert.Equal(t, "deep error", target.msg, "error message should be preserved")
}

func TestMapErr_WithErrorInChainNotAtEnd(t *testing.T) {
	customErr := &CustomError{msg: "middle error"}
	wrappedOnce := fmt.Errorf("wrapping: %w", customErr)
	wrappedTwice := fmt.Errorf("another wrap: %w", wrappedOnce)

	target, ok := MapErr[*CustomError](wrappedTwice)

	assert.True(t, ok, "MapErr should find error in middle of chain")
	assert.NotNil(t, target, "target should not be nil")
	assert.Equal(t, customErr, target, "target should be the original error")
}

func TestMapErr_WithDifferentErrorTypesInChain(t *testing.T) {
	customErr := &CustomError{msg: "custom"}
	anotherErr := &AnotherError{code: 123}
	wrapped := fmt.Errorf("wrapping another: %w", anotherErr)

	target1, ok1 := MapErr[*CustomError](wrapped)
	assert.False(t, ok1, "should not find CustomError")
	assert.Nil(t, target1, "should not have target")

	target2, ok2 := MapErr[*AnotherError](wrapped)
	assert.True(t, ok2, "should find AnotherError")
	assert.NotNil(t, target2, "should have target")
	assert.Equal(t, anotherErr, target2, "target should be AnotherError")

	mixedErr := fmt.Errorf("wrapped custom: %w", customErr)
	target3, ok3 := MapErr[*CustomError](mixedErr)
	assert.True(t, ok3, "should find CustomError in mixed chain")
	assert.NotNil(t, target3, "should have target")
	assert.Equal(t, customErr, target3, "target should be CustomError")
}

func TestMapErr_WithEmptyCustomError(t *testing.T) {
	emptyErr := &CustomError{msg: ""}
	target, ok := MapErr[*CustomError](emptyErr)

	assert.True(t, ok, "MapErr should handle empty custom error")
	assert.NotNil(t, target, "target should not be nil even if message is empty")
	assert.Equal(t, "", target.msg, "error message should be empty")
}

func TestMapErr_PreservesErrorInformation(t *testing.T) {
	testMsg := "important error information"
	originalErr := &CustomError{msg: testMsg}

	target, ok := MapErr[*CustomError](originalErr)

	assert.True(t, ok, "MapErr should succeed")
	assert.NotNil(t, target, "target should not be nil")
	assert.Equal(t, testMsg, target.msg, "error message must be exactly preserved")
	assert.Equal(t, testMsg, target.Error(), "Error() method should return preserved message")
}
