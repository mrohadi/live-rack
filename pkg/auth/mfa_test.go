package auth

import "testing"

func TestAmrIndicatesMFA(t *testing.T) {
	assert := func(got, want bool, name string) {
		if got != want {
			t.Errorf("%s: got %v want %v", name, got, want)
		}
	}
	assert(amrIndicatesMFA([]any{"pwd", "mfa"}), true, "mfa marker")
	assert(amrIndicatesMFA([]any{"pwd", "OTP"}), true, "case-insensitive otp")
	assert(amrIndicatesMFA([]any{"webauthn"}), true, "webauthn")
	assert(amrIndicatesMFA([]any{"pwd"}), false, "password only")
	assert(amrIndicatesMFA([]any{}), false, "empty")
	assert(amrIndicatesMFA([]any{123, "pwd"}), false, "non-string ignored")
}

func TestMfaFromClaims(t *testing.T) {
	if !mfaFromClaims(map[string]any{"amr": []any{"pwd", "mfa"}}) {
		t.Error("expected MFA true")
	}
	if mfaFromClaims(map[string]any{"amr": []any{"pwd"}}) {
		t.Error("expected MFA false")
	}
	if mfaFromClaims(map[string]any{}) {
		t.Error("missing amr -> false")
	}
}
