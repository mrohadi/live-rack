package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewriteTOTPIssuer(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "replaces ZITADEL issuer",
			input: "otpauth://totp/ZITADEL:user%40email.com?issuer=ZITADEL&secret=ABCD",
			want:  "otpauth://totp/live-rack:user@email.com?issuer=live-rack&secret=ABCD",
		},
		{
			name:  "no issuer param left unchanged",
			input: "otpauth://totp/ZITADEL:user%40email.com?secret=ABCD",
			want:  "otpauth://totp/live-rack:user@email.com?secret=ABCD",
		},
		{
			name:  "invalid URI returned as-is",
			input: "not-a-uri",
			want:  "not-a-uri",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rewriteTOTPIssuer(tc.input, "live-rack")
			assert.Equal(t, tc.want, got)
		})
	}
}
