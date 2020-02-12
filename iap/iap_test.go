package iap

import (
	"github.com/mitsuki1995/iap-verifier/ios"
	"testing"
)

func TestIOSVerifier(t *testing.T) {
	info, err := ios.Verify("a", "a", true, true)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}
}
