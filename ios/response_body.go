package ios

// https://developer.apple.com/documentation/appstorereceipts/responsebody
type ResponseBody struct {
	Environment        Environment    `json:"environment"` // Sandbox, Production
	IsRetryable        bool           `json:"is_retryable"`
	LatestReceipt      string         `json:"latest_receipt"`
	LatestReceiptInfo  []*ReceiptInfo `json:"latest_receipt_info"`
	PendingRenewalInfo []*RenewalInfo `json:"pending_renewal_info"`
	Status             int            `json:"status"` // 0 is ok
	//Receipt // 这个是收据的解码版，如果只是找交易信息的话，从LatestReceiptInfo里找会好一些
}
