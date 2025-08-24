package claudecode

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	expected := "0.1.0"
	if Version != expected {
		t.Errorf("Version = %q, want %q", Version, expected)
	}
}
