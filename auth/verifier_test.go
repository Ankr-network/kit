// +build integration

package auth

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMethodMatch(t *testing.T) {
	rv, err := NewVerifier(ExcludeMethods(
		`/ankr\.uaa\.sms\.v.+PublicSMS/.+`,
		`/ankr\.uaa\.user\.v.+PublicUser/ConfirmEmail`,
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
			"/ankr.uaa.sms.v1alpha.PublicSMS/Send",
			true,
		},
		{
			"/ankr.uaa.sms.v1alpha.PublicSMS/Check",
			true,
		},
		{
			"/ankr.uaa.totp.v1alpha.InternalTOTP/Status",
			true,
		},
		{
			"/ankr.uaa.totp.v1alpha.InternalTOTP/Check",
			true,
		},
		{
			"/ankr.uaa.user.v1alpha.PublicUser/ConfirmEmail",
			true,
		},
	}

	for _, d := range td {
		t.Run(d.method, func(t *testing.T) {
			assert.Equal(t, d.res, r.matchMethod(d.method))
		})
	}
}
