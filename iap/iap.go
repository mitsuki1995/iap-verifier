package iap

import (
	"github.com/mitsuki1995/iap-verifier/common"
)

func IsTransactionNotFoundError(err error) bool {
	return err == common.TransactionNotFoundError
}
