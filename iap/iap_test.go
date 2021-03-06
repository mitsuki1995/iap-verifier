package iap

import (
	"encoding/json"
	"github.com/mitsuki1995/iap-verifier/ios"
	"io/ioutil"
	"testing"
)

func readString(filename string) string {
	if b, err := ioutil.ReadFile(filename); err != nil {
		return ""
	} else {
		return string(b)
	}
}

func TestIOSVerifier(t *testing.T) {

	// File ReceiptData: base64 encoded string
	receiptData := readString("ReceiptData1")
	password := readString("password")
	v := ios.NewVerifier(password, true)
	infos, err := v.Verify(receiptData, false)
	if err != nil {
		t.Error(err)
	} else {
		b, _ := json.MarshalIndent(infos, "", "  ")
		t.Log(string(b))
	}
}
