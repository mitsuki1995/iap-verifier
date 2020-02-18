package ios

import (
	"encoding/json"
	"fmt"
	"github.com/mitsuki1995/iap-verifier/common"
	"io/ioutil"
)

const (
	iOSSandboxVerifyURL    = "https://sandbox.itunes.apple.com/verifyReceipt"
	iOSProductionVerifyURL = "https://buy.itunes.apple.com/verifyReceipt"
)

type Verifier struct {
	password string
	isDebug  bool
}

// "isDebug = true" means "send to sandbox url first"
func NewVerifier(password string, isDebug bool) *Verifier {
	return &Verifier{
		password: password,
		isDebug:  isDebug,
	}
}

func (v *Verifier) Verify(receiptData string, excludeOldTransactions bool) (map[string]*TransactionInfo, error) {
	return v.verifyReceipt(receiptData, excludeOldTransactions, false)
}

// reversed: 如果为 true, 正式服务器会找苹果的沙盒服务器进行验证, 测试服务器会找苹果的正式服务器进行验证
func (v *Verifier) verifyReceipt(receiptData string, excludeOldTransactions bool, reversed bool) (map[string]*TransactionInfo, error) {

	url := iOSSandboxVerifyURL
	if v.isDebug == reversed {
		url = iOSProductionVerifyURL
	}

	response, err := common.PostJSON(url, map[string]interface{}{
		"password":                 v.password,
		"receipt-data":             receiptData,
		"exclude-old-transactions": excludeOldTransactions,
	})
	if err != nil {
		return nil, fmt.Errorf("post JSON error: %s", err.Error())
	}

	resultByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response.Body error: %s", err.Error())
	}

	body := new(ResponseBody)
	if err := json.Unmarshal(resultByte, body); err != nil {
		return nil, fmt.Errorf("unmarshal body error: %s", err.Error())
	}

	status := body.Status

	// https://developer.apple.com/library/archive/releasenotes/General/ValidateAppStoreReceipt/Chapters/ValidateRemotely.html
	if status != 0 {

		// This receipt is from the test environment, but it was sent to the production environment for verification.
		if status == 21007 && !reversed { // 测试环境的收据提交到了正式服务器
			return v.verifyReceipt(receiptData, excludeOldTransactions, true)
		}

		// This receipt is from the production environment, but it was sent to the test environment for verification.
		if status == 21008 && !reversed { // 正式环境的收据提交到了测试服务器
			return v.verifyReceipt(receiptData, excludeOldTransactions, true)
		}

		return nil, fmt.Errorf("invalid status: %d", status)
	}

	return v.FindTransactionInfo(body.LatestReceiptInfo, body.PendingRenewalInfo), nil
}

// 查找尚未过期且过期时间最靠后的一条交易
// key: productID
func (v *Verifier) FindTransactionInfo(latestReceiptInfo []*ReceiptInfo, pendingRenewalInfo []*RenewalInfo) map[string]*TransactionInfo {

	var payCounts = make(map[string]int) // key: original_transaction_id

	transactionInfos := make(map[string]*TransactionInfo)
	for _, receiptInfo := range latestReceiptInfo {

		if len(receiptInfo.CancellationDateMS) > 0 {
			continue
		}

		if len(receiptInfo.ProductID) == 0 || len(receiptInfo.OriginalTransactionID) == 0 ||
			len(receiptInfo.ExpiresDateMS) == 0 || len(receiptInfo.PurchaseDateMS) == 0 {
			continue
		}

		if !receiptInfo.IsTrialPeriod.Bool() {
			payCounts[receiptInfo.OriginalTransactionID] += 1
		}

		transactionInfo := transactionInfos[receiptInfo.ProductID]
		if transactionInfo == nil || receiptInfo.ExpiresDate().After(transactionInfo.ReceiptInfo.ExpiresDate()) {
			transactionInfos[receiptInfo.ProductID] = &TransactionInfo{
				ReceiptInfo: *receiptInfo,
			}
		}
	}

	for _, renewalInfo := range pendingRenewalInfo {
		if transactionInfo, ok := transactionInfos[renewalInfo.ProductID]; ok {
			transactionInfo.RenewalInfo = *renewalInfo
		}
	}

	for _, transactionInfo := range transactionInfos {
		transactionInfo.PayCount = payCounts[transactionInfo.ReceiptInfo.OriginalTransactionID]
	}

	return transactionInfos
}
