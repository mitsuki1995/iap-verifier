package ios

import (
	"encoding/json"
	"strconv"
	"time"
)

// https://developer.apple.com/documentation/appstorereceipts/responsebody/latest_receipt_info
type ReceiptInfo struct {
	CancellationDateMS    string `json:"cancellation_date_ms"`
	CancellationReason    string `json:"cancellation_reason"`
	ExpiresDateMS         string `json:"expires_date_ms"`
	IsInIntroOfferPeriod  string `json:"is_in_intro_offer_period"`
	IsTrialPeriod         string `json:"is_trial_period"`
	OriginalTransactionID string `json:"original_transaction_id"`
	ProductID             string `json:"product_id"`
	PurchaseDateMS        string `json:"purchase_date_ms"`
	TransactionID         string `json:"transaction_id"`
}

func (r *ReceiptInfo) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (r *ReceiptInfo) ExpiryTime() time.Time {
	expiresDateMS, err := strconv.ParseFloat(r.ExpiresDateMS, 64)
	if err != nil {
		return time.Unix(0, 0)
	}
	return time.Unix(int64(expiresDateMS) / 1e3, (int64(expiresDateMS) % 1e3) * 1e6)
}

// https://developer.apple.com/documentation/appstorereceipts/responsebody/pending_renewal_info
type RenewalInfo struct {
	AutoRenewProductID       string `json:"auto_renew_product_id"`
	AutoRenewStatus          string `json:"auto_renew_status"` // true | false
	ExpirationIntent         string `json:"expiration_intent"` // 1, 2, 3, 4, 5
	GracePeriodExpiresDateMS string `json:"grace_period_expires_date_ms"`
	IsInBillingRetryPeriod   string `json:"is_in_billing_retry_period"`
	OriginalTransactionID    string `json:"original_transaction_id"`
	ProductID                string `json:"product_id"`
}

// https://developer.apple.com/documentation/appstorereceipts/responsebody
type ResponseBody struct {
	Environment        string         `json:"environment"`
	IsRetryable        bool           `json:"is_retryable"`
	LatestReceipt      string         `json:"latest_receipt"`
	LatestReceiptInfo  []*ReceiptInfo `json:"latest_receipt_info"`
	PendingRenewalInfo []*RenewalInfo `json:"pending_renewal_info"`
	Status             int            `json:"status"`
	//Receipt // 这个是收据的解码版，如果只是找交易信息的话，从LatestReceiptInfo里找会好一些
}

type TransactionInfo struct {
	ActiveReceiptInfo *ReceiptInfo
	RenewalInfo       *RenewalInfo
}
