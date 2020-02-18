package iap

import (
	"encoding/json"
	"github.com/mitsuki1995/iap-verifier/ios"
	"io/ioutil"
	"testing"
)

func TestIOSVerifier(t *testing.T) {

	// File ReceiptData: base64 encoded string
	if b, err := ioutil.ReadFile("ReceiptData"); err != nil {
		t.Fatal(err)
	} else {
		info, err := ios.Verify("b5e0eb1004684720a70f8dfd1bfe0d9e", string(b), false, false)
		if err != nil {
			t.Error(err)
		} else {
			b, _ := json.MarshalIndent(info, "", "  ")
			t.Log(string(b))
			t.Log(info.ReceiptInfo.ExpiresDate())
		}
	}
}
