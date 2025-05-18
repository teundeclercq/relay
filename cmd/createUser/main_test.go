package main_test

import (
	"testing"

	"github.com/pquerna/otp/totp"
)

func TestOtp(t *testing.T) {
	secret := "G4O4R6NUXSLENA45NUANWDIHYMSAX5CO"
	code := "891758" // replace with live code from your app

	valid := totp.Validate(code, secret)

	if !valid {
		t.Errorf("expected valid TOTP code for secret %q, got invalid", secret)
	}
}
