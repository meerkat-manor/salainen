package generate

import (
	"strings"
	"testing"
)

func TestAuthenticationSecretUnknownContentSet(t *testing.T) {

	params := &GenerateParameters{
		MinimumLength: 8,
		MaximumLength: 8,
		ContentSet:    "NOT_A_CONTENT_SET",
	}

	if _, err := AuthenticationSecret(params); err == nil {
		t.Errorf("expected an error for an unknown content set")
	}
}

func TestGenerateCredentialZeroLength(t *testing.T) {

	for _, credentialType := range []string{"DIGIT", "ALPHA", "ALPHASYMBOLSMIXED", "NOT_A_TYPE"} {
		value, err := GenerateCredential(credentialType, 0)
		if credentialType == "NOT_A_TYPE" {
			if err == nil {
				t.Errorf("expected an error for an unknown credential type")
			}
			continue
		}
		if err == nil {
			t.Errorf("expected error for %q: zero length, got '%s'", credentialType, value)
		}
	}
}

func TestAuthenticationSecretUsesDefaultContentSet(t *testing.T) {

	params := &GenerateParameters{MinimumLength: 32, MaximumLength: 32}
	value, err := AuthenticationSecret(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(value) != 32 {
		t.Fatalf("expected length 32, got %d", len(value))
	}
	for _, char := range value {
		if !strings.ContainsRune(alphasymbols_mixed, char) {
			t.Errorf("unexpected character %q outside default content set", char)
		}
	}
}
