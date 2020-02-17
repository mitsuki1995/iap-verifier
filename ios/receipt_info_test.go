package ios

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"
)

func TestMarshallingReceiptInfo(t *testing.T) {
	receiptInfo := &ReceiptInfo{
		CancellationDateMS:    strconv.Itoa(int(time.Now().Unix() * 1000)),
		CancellationReason:    CancellationReasonApp,
		ExpiresDateMS:         strconv.Itoa(int(time.Now().Unix() * 1000)),
		IsInIntroOfferPeriod:  "false",
		IsTrialPeriod:         "true",
		OriginalTransactionID: "123456",
		ProductID:             "abcd",
		PurchaseDateMS:        strconv.Itoa(int(time.Now().Unix() * 1000)),
		TransactionID:         "123456",
	}
	b, err := json.Marshal(receiptInfo)
	if err != nil {
		t.Error(err)
	}
	_ = json.Unmarshal(b, receiptInfo)
	t.Log(receiptInfo.String())
}
