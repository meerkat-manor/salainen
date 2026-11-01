package generate

import (
	"strings"
	"testing"
)

func TestAuthenticationSecretDefaults(t *testing.T) {

	value, err := AuthenticationSecret(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(value) < 10 || len(value) > 20 {
		t.Errorf("default length out of range: %d", len(value))
	}
}

func TestAuthenticationSecretParameters(t *testing.T) {

	tests := []struct {
		name       string
		parameters *GenerateParameters
		wantLen    int
	}{
		{
			name:       "explicit alpha",
			parameters: &GenerateParameters{MinimumLength: 5, MaximumLength: 12, ContentSet: "ALPHA"},
			wantLen:    12,
		},
		{
			name:       "empty content set defaults",
			parameters: &GenerateParameters{MinimumLength: 3, MaximumLength: 7},
			wantLen:    7,
		},
		{
			name:       "maximum length boundary",
			parameters: &GenerateParameters{MinimumLength: 50, MaximumLength: 50, ContentSet: "ALPHALOWER"},
			wantLen:    50,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, err := AuthenticationSecret(test.parameters)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(value) != test.wantLen {
				t.Errorf("expected length %d, got %d", test.wantLen, len(value))
			}
		})
	}
}

func TestAuthenticationSecretInvalidParameters(t *testing.T) {

	tests := []struct {
		name       string
		parameters *GenerateParameters
	}{
		{name: "zero minimum", parameters: &GenerateParameters{MinimumLength: 0, MaximumLength: 10}},
		{name: "zero maximum", parameters: &GenerateParameters{MinimumLength: 1, MaximumLength: 0}},
		{name: "maximum too large", parameters: &GenerateParameters{MinimumLength: 1, MaximumLength: 51}},
		{name: "maximum below minimum", parameters: &GenerateParameters{MinimumLength: 20, MaximumLength: 10}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := AuthenticationSecret(test.parameters); err == nil {
				t.Errorf("expected error")
			}
		})
	}
}

func TestGenerateCredentialTypes(t *testing.T) {

	tests := []struct {
		credentialType string
		length         int
		allowed        string
	}{
		{credentialType: "DIGIT", length: 1, allowed: "0123456789"},
		{credentialType: "ALPHAUPPER", length: 16, allowed: upperchar},
		{credentialType: "ALPHALOWER", length: 16, allowed: lowerchar},
		{credentialType: "ALPHANUMERIC", length: 16, allowed: alphanumeric},
		{credentialType: "ALPHANUMERICMIXED", length: 16, allowed: alphanumeric_mixed},
		{credentialType: "ALPHASYMBOLS", length: 16, allowed: alphasymbols},
		{credentialType: "ALPHASYMBOLSMIXED", length: 16, allowed: alphasymbols_mixed},
		{credentialType: "MIXEDALPHA", length: 16, allowed: mixedchar},
		{credentialType: "ALPHA", length: 16, allowed: mixedchar},
		{credentialType: "alpha", length: 16, allowed: mixedchar},
	}

	for _, test := range tests {
		t.Run(test.credentialType, func(t *testing.T) {
			value, err := GenerateCredential(test.credentialType, test.length)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(value) != test.length {
				t.Errorf("expected length %d, got %d", test.length, len(value))
			}
			for _, char := range value {
				if !strings.ContainsRune(test.allowed, char) {
					t.Errorf("unexpected character %q in generated value", char)
				}
			}
		})
	}
}

func TestGenerateCredentialUnknownType(t *testing.T) {

	if _, err := GenerateCredential("NOT_A_TYPE", 10); err == nil {
		t.Errorf("expected error for unknown credential type")
	}
}
