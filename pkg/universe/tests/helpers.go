package tests

import (
	"testing"
)

func assertNoError(t *testing.T, err error, name string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", name, err)
	}
	t.Logf("%s: no error", name)
}

func assertEqual[T comparable](t *testing.T, expected, actual T, name string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", name, expected, actual)
	}
	t.Logf("%s: expected %v, got %v", name, expected, actual)
}