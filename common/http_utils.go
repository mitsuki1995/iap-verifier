package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/kataras/golog"
	"net/http"
	"time"
)

var (
	TransactionNotFoundError = errors.New("transaction not found")
)

// 以json为body的post请求
// 可优化点: 重用TCP连接
func PostJSON(url string, body interface{}) (*http.Response, error) {
	bytesData, err := json.Marshal(body)
	if err != nil {
		golog.Warn("PostJSON, Marshal error = ", err.Error())
		return nil, err
	}

	reader := bytes.NewReader(bytesData)

	request, err := http.NewRequest("POST", url, reader)
	if request != nil {
		//noinspection GoUnhandledErrorResult
		defer request.Body.Close()
	}
	if err != nil {
		golog.Warn("PostJSON, NewRequest error = ", err.Error())
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := http.Client{
		Timeout: time.Second * 60, // 从请求开始到响应消息体被完全接收的时间限制。
	}
	response, err := client.Do(request) // Do 方法发送请求，返回 HTTP 回复
	if err != nil {
		golog.Warn("PostJSON, Do error = ", err.Error())
		return nil, err
	}

	return response, nil
}
