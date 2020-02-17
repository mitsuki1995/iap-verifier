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

func Verify(password string, receiptData string, excludeOldTransactions bool, isDebug bool) (*TransactionInfo, error) {
	return verifyReceipt(password, receiptData, excludeOldTransactions, isDebug, false)
}

// reversed: 如果为 true, 正式服务器会找苹果的沙盒服务器进行验证, 测试服务器会找苹果的正式服务器进行验证
func verifyReceipt(password string, receiptData string, excludeOldTransactions bool, isDebug bool, reversed bool) (*TransactionInfo, error) {

	url := iOSSandboxVerifyURL
	if isDebug == reversed {
		url = iOSProductionVerifyURL
	}

	response, err := common.PostJSON(url, map[string]interface{}{
		"password":                 password,
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
			return verifyReceipt(password, receiptData, excludeOldTransactions, isDebug, true)
		}

		// This receipt is from the production environment, but it was sent to the test environment for verification.
		if status == 21008 && !reversed { // 正式环境的收据提交到了测试服务器
			return verifyReceipt(password, receiptData, excludeOldTransactions, isDebug, true)
		}

		return nil, fmt.Errorf("invalid status: %d", status)
	}

	transactionInfo := findTransactionInfo(body.LatestReceiptInfo, body.PendingRenewalInfo)

	if transactionInfo == nil {
		return nil, common.TransactionNotFoundError
	}

	return transactionInfo, nil
}

// 查找尚未过期且过期时间最靠后的一条交易, 找不到返回 nil
func findTransactionInfo(latestReceiptInfo []*ReceiptInfo, pendingRenewalInfo []*RenewalInfo) *TransactionInfo {

	var payCounts = make(map[string]int) // key: original_transaction_id

	var activeReceiptInfo *ReceiptInfo
	for _, receiptInfo := range latestReceiptInfo {

		if len(receiptInfo.CancellationDateMS) > 0 {
			continue
		}

		if len(receiptInfo.ProductID) == 0 || len(receiptInfo.OriginalTransactionID) == 0 || len(receiptInfo.ExpiresDateMS) == 0 {
			continue
		}

		payCounts[receiptInfo.OriginalTransactionID] += 1

		if activeReceiptInfo == nil || receiptInfo.ExpiryTime().After(activeReceiptInfo.ExpiryTime()) {
			activeReceiptInfo = receiptInfo
		}
	}

	if activeReceiptInfo == nil {
		return nil
	}

	transactionInfo := &TransactionInfo{
		ActiveReceiptInfo: activeReceiptInfo,
	}

	for _, renewalInfo := range pendingRenewalInfo {

		if activeReceiptInfo.ProductID == renewalInfo.ProductID {
			transactionInfo.RenewalInfo = renewalInfo
			break
		}
	}

	return transactionInfo
}
