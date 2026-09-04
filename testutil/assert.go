package testutil

import (
	"encoding/json"
	"testing"
)

// AssertJSONNotError asserts that no error occurred.
func AssertJSONNotError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// AssertJSONError asserts that an error occurred.
func AssertJSONError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertJSONEqual asserts that two JSON values are equal.
func AssertJSONEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	expBytes, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected: %v", err)
	}
	actBytes, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("failed to marshal actual: %v", err)
	}
	if string(expBytes) != string(actBytes) {
		t.Fatalf("JSON mismatch:\nexpected: %s\nactual:   %s", expBytes, actBytes)
	}
}

// AssertJSONContains asserts that a JSON response contains a key.
func AssertJSONContains(t *testing.T, resp map[string]any, key string) {
	t.Helper()
	if _, ok := resp[key]; !ok {
		t.Fatalf("expected key '%s' in response", key)
	}
}

// AssertResultCount asserts that the echo server received the expected count.
func AssertResultCount(t *testing.T, got, expected int) {
	t.Helper()
	if got != expected {
		t.Fatalf("expected %d results, got %d", expected, got)
	}
}
