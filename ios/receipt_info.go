package ios

import (
	"encoding/json"
	"github.com/mitsuki1995/iap-verifier/common"
	"time"
)

const (
	CancellationReasonApp  CancellationReason = "1"
	CancellationReasonUser CancellationReason = "0"
)

type CancellationReason string

// https://developer.apple.com/documentation/appstorereceipts/responsebody/latest_receipt_info
type ReceiptInfo struct {
	CancellationDateMS    string             `json:"cancellation_date_ms"`
	CancellationReason    CancellationReason `json:"cancellation_reason"` // 1, 0
	ExpiresDateMS         string             `json:"expires_date_ms"`
	IsInIntroOfferPeriod  StringBool         `json:"is_in_intro_offer_period"`
	IsTrialPeriod         StringBool         `json:"is_trial_period"`
	OriginalTransactionID string             `json:"original_transaction_id"`
	ProductID             string             `json:"product_id"`
	PurchaseDateMS        string             `json:"purchase_date_ms"`
	TransactionID         string             `json:"transaction_id"`
}

func (r *ReceiptInfo) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (r *ReceiptInfo) CancellationDate() time.Time {
	return common.TimeMillisStringToTime(r.CancellationDateMS)
}

func (r *ReceiptInfo) ExpiresDate() time.Time {
	return common.TimeMillisStringToTime(r.ExpiresDateMS)
}

func (r *ReceiptInfo) PurchaseDate() time.Time {
	return common.TimeMillisStringToTime(r.PurchaseDateMS)
}
