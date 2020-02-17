package ios

type TransactionInfo struct {
	ReceiptInfo ReceiptInfo
	RenewalInfo RenewalInfo
	PayCount    int
}

func (t *TransactionInfo) hasRenewalInfo() bool {
	return len(t.RenewalInfo.ProductID) > 0
}
