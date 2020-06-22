// +build integration

package auth

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMethodMatch(t *testing.T) {
	rv, err := NewVerifier(MustLoadVerifierConfig().RSAPublicKeyPath, ExcludeMethods(
		`/test\.uaa\.sms\.v.+PublicSMS/.+`,
		`/test\.uaa\.user\.v.+PublicUser/ConfirmEmail`,
		`/.+Internal.+/.+`),
	)
	require.NoError(t, err)
	r, ok := rv.(*verifier)
	require.True(t, ok)

	td := []struct {
		method string
		res    bool
	}{
		{
			"/test.uaa.sms.v1alpha.PublicSMS/Send",
			true,
		},
		{
			"/test.uaa.sms.v1alpha.PublicSMS/Check",
			true,
		},
		{
			"/test.uaa.totp.v1alpha.InternalTOTP/Status",
			true,
		},
		{
			"/test.uaa.totp.v1alpha.InternalTOTP/Check",
			true,
		},
		{
			"/test.uaa.user.v1alpha.PublicUser/ConfirmEmail",
			true,
		},
	}

	for _, d := range td {
		t.Run(d.method, func(t *testing.T) {
			assert.Equal(t, d.res, r.matchMethod(d.method))
		})
	}
}
