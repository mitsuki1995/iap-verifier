package ios

import (
	"github.com/mitsuki1995/iap-verifier/common"
	"time"
)

// https://developer.apple.com/documentation/appstorereceipts/responsebody/pending_renewal_info
// ExpirationIntent: https://developer.apple.com/documentation/appstorereceipts/expiration_intent
type RenewalInfo struct {
	AutoRenewProductID       string  `json:"auto_renew_product_id"`
	AutoRenewStatus          IntBool `json:"auto_renew_status"`
	ExpirationIntent         string  `json:"expiration_intent"`
	GracePeriodExpiresDateMS string  `json:"grace_period_expires_date_ms"`
	IsInBillingRetryPeriod   IntBool `json:"is_in_billing_retry_period"`
	OriginalTransactionID    string  `json:"original_transaction_id"`
	ProductID                string  `json:"product_id"`
}

func (r *RenewalInfo) GracePeriodExpiresDate() time.Time {
	return common.TimeMillisStringToTime(r.GracePeriodExpiresDateMS)
}
