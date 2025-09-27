package datetime

import (
	"testing"
	"time"
)

// testTime creates a test time in local timezone for consistent testing
func testTime(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.Local)
}

// testTimeUTC creates a test time in UTC - kept for backward compatibility where needed
func testTimeUTC(year int, month time.Month, day, hour, min, sec int) time.Time {
	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

// assertTimeEqual compares two time.Time values with a helper for better test output
func assertTimeEqual(t *testing.T, expected, actual time.Time, msg string) {
	t.Helper()
	if !expected.Equal(actual) {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// assertStringEqual compares two strings with a helper for better test output
func assertStringEqual(t *testing.T, expected, actual, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %s, got %s", msg, expected, actual)
	}
}

// assertIntEqual compares two integers with a helper for better test output
func assertIntEqual(t *testing.T, expected, actual int, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %d, got %d", msg, expected, actual)
	}
}

// assertBoolEqual compares two booleans with a helper for better test output
func assertBoolEqual(t *testing.T, expected, actual bool, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %t, got %t", msg, expected, actual)
	}
}

// assertNoError checks that error is nil with a helper for better test output
func assertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: unexpected error: %v", msg, err)
	}
}

// assertError checks that error is not nil with a helper for better test output
func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("%s: expected error but got none", msg)
	}
}
