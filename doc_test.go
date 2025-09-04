package claudecode

import (
	"regexp"
	"testing"
)

func TestVersionBasic(t *testing.T) {
	assertDocVersion(t, Version, "0.1.0")
}

func TestVersionValidation(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantValid   bool
		description string
	}{
		{
			name:        "valid_semantic_version",
			version:     "0.1.0",
			wantValid:   true,
			description: "Standard semantic version format",
		},
		{
			name:        "empty_version",
			version:     "",
			wantValid:   false,
			description: "Empty version should be invalid",
		},
		{
			name:        "invalid_format",
			version:     "v1.0",
			wantValid:   false,
			description: "Non-semantic version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.version == Version {
				assertDocValidSemVer(t, tt.version, tt.wantValid, tt.description)
			}
		})
	}
}

func TestVersionImmutability(t *testing.T) {
	originalVersion := Version
	assertDocVersion(t, Version, originalVersion)

	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestPackageDocumentation(t *testing.T) {
	if Version == "" {
		t.Error("Package version should be defined for documentation purposes")
	}

	assertDocValidSemVer(t, Version, true, "Package version should follow semantic versioning")
}

func assertDocVersion(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("Version = %q, want %q", got, want)
	}
}

func assertDocValidSemVer(t *testing.T, version string, wantValid bool, description string) {
	t.Helper()

	if version == "" {
		if wantValid {
			t.Errorf("%s: empty version should be invalid", description)
		}
		return
	}

	semVerPattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	matched, err := regexp.MatchString(semVerPattern, version)
	if err != nil {
		t.Fatalf("%s: failed to validate semantic version pattern: %v", description, err)
	}

	if matched != wantValid {
		if wantValid {
			t.Errorf("%s: version %q should be valid semantic version", description, version)
		} else {
			t.Errorf("%s: version %q should be invalid semantic version", description, version)
		}
	}
}
