package ios

import (
	"fmt"
	"github.com/mitsuki1995/iap-verifier/common"
	"github.com/bitly/go-simplejson"
	"github.com/kataras/golog"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	iOSSandboxVerifyURL    = "https://sandbox.itunes.apple.com/verifyReceipt"
	iOSProductionVerifyURL = "https://buy.itunes.apple.com/verifyReceipt"
)

type TransactionInfo struct {
	ActiveTransactionInfo
	PendingRenewalInfo
}

type ActiveTransactionInfo struct {
	ProductID             string
	OriginalTransactionID string
	StartTimeMS           float64
	ExpiryTimeMS          float64
	IsTrialPeriod         bool
}

type PendingRenewalInfo struct {
	AutoRenewProductID     string
	AutoRenewStatus        string
	ExpirationIntent       string
	IsInBillingRetryPeriod string
}

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
		return nil, fmt.Errorf("PostJSON error: %s", err.Error())
	}

	resultByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadResponseBody error: %s", err.Error())
	}

	resultJson, err := simplejson.NewJson(resultByte)
	if err != nil {
		return nil, fmt.Errorf("ParseResponseBody error: %s", err.Error())
	}

	status := resultJson.Get("status").MustInt(-1)

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

	receipt := resultJson.Get("receipt")

	// 通过票据解析出来的所有交易信息
	inAppTransactions := receipt.Get("in_app")

	// 自动续订的最新交易（即使使用老的票据也能够取出最新的交易）
	latestTransactions := resultJson.Get("latest_receipt_info")

	active := findActiveTransactionInfo(inAppTransactions, latestTransactions)

	if active != nil {

		pendingRenewalInfo := findPendingRenewalInfo(resultJson.Get("pending_renewal_info"), active.ProductID)

		return &TransactionInfo{
			ActiveTransactionInfo: *active,
			PendingRenewalInfo:    *pendingRenewalInfo,
		}, nil
	}

	return nil, common.TransactionNotFoundError
}

// 查找尚未过期且过期时间最靠后的一条交易, 找不到返回 nil
func findActiveTransactionInfo(someTransactions ...*simplejson.Json) *ActiveTransactionInfo {

	var result *ActiveTransactionInfo
	var nowMS = float64(time.Now().UnixNano() / 1e6)

	var originalTransactionIDs = make(map[string]bool)

	for _, transactions := range someTransactions {
		array, err := transactions.Array()
		if err != nil {
			golog.Warn("findActiveTransactionInfo, transactions is not an array")
		}
		for index, length := 0, len(array); index < length; index++ {
			transaction := transactions.GetIndex(index)

			// Treat a canceled receipt the same as if no purchase had ever been made.
			// 这里的`cancel`是退款的意思
			if _, canceled := transaction.CheckGet("cancellation_date"); canceled {
				golog.Info("findActiveTransactionInfo, found canceled transaction")
				continue
			}

			productID := transaction.Get("product_id").MustString("")
			if len(productID) == 0 {
				golog.Warn("findActiveTransactionInfo, product_id is empty")
				continue
			}

			originalTransactionID := transaction.Get("original_transaction_id").MustString("")
			if len(originalTransactionID) == 0 {
				golog.Warn("findActiveTransactionInfo, original_transaction_id is empty")
				continue
			}

			expiryTimeStr := transaction.Get("expires_date_ms").MustString("")
			if len(expiryTimeStr) == 0 {
				golog.Warn("findActiveTransactionInfo, expires_date_ms is empty")
				continue
			}

			expiryTimeMS, err := strconv.ParseFloat(expiryTimeStr, 64)
			if err != nil {
				golog.Warn("findActiveTransactionInfo, parse expires_date_ms error = ", err.Error())
				continue
			}

			if expiryTimeMS > nowMS {

				originalTransactionIDs[originalTransactionID] = true

				if result == nil || expiryTimeMS > result.ExpiryTimeMS {
					startTimeMS, _ := strconv.ParseFloat(transaction.Get("purchase_date_ms").MustString("0"), 64)
					result = &ActiveTransactionInfo{
						ProductID:             productID,
						OriginalTransactionID: originalTransactionID,
						StartTimeMS:           startTimeMS,
						ExpiryTimeMS:          expiryTimeMS,
						IsTrialPeriod:         transaction.Get("is_trial_period").MustBool(false),
					}
				}
			}
		}
	}

	if len(originalTransactionIDs) > 1 {
		golog.Warn("find ", len(originalTransactionIDs), " active transactionInfos, choose one: ", result)
	}

	return result
}

// 找不到返回空
func findPendingRenewalInfo(infos *simplejson.Json, productID string) *PendingRenewalInfo {
	length := len(infos.MustArray())
	for i := 0; i < length; i++ {
		info := infos.GetIndex(i)
		pid := info.Get("product_id").MustString()
		if pid == productID {

			return &PendingRenewalInfo{
				AutoRenewProductID:     info.Get("auto_renew_product_id").MustString(),
				AutoRenewStatus:        info.Get("auto_renew_status").MustString(),
				ExpirationIntent:       info.Get("expiration_intent").MustString(),
				IsInBillingRetryPeriod: info.Get("is_in_billing_retry_period").MustString(),
			}
		}
	}

	return nil
}
