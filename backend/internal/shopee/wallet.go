package shopee

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// WalletTransactionResponse 钱包交易记录响应
type WalletTransactionResponse struct {
	BaseResponse
	Response struct {
		More            bool          `json:"more"`
		TransactionList []Transaction `json:"transaction_list"`
	} `json:"response"`
}

// Transaction 交易记录
type Transaction struct {
	TransactionID      int64   `json:"transaction_id"`
	Status             string  `json:"status"`
	WalletType         string  `json:"wallet_type"`
	TransactionType    string  `json:"transaction_type"`
	Amount             float64 `json:"amount"`
	CurrentBalance     float64 `json:"current_balance"`
	CreateTime         int64   `json:"create_time"`
	Reason             string  `json:"reason"`
	OrderSN            string  `json:"order_sn"`
	RefundSN           string  `json:"refund_sn"`
	WithdrawalType     string  `json:"withdrawal_type"`
	TransactionFee     float64 `json:"transaction_fee"`
	Description        string  `json:"description"`
	BuyerName          string  `json:"buyer_name"`
	WithdrawalID       int64   `json:"withdrawal_id"`
	TransactionTabType string  `json:"transaction_tab_type"`
	MoneyFlow          string  `json:"money_flow"`
}

// GetWalletTransactionList 获取钱包交易记录列表
func (c *Client) GetWalletTransactionList(accessToken string, shopID uint64, pageNo, pageSize int, walletType string) (*WalletTransactionResponse, error) {
	path := "/api/v2/payment/get_wallet_transaction_list"
	params := url.Values{}
	params.Set("page_no", strconv.Itoa(pageNo))
	params.Set("page_size", strconv.Itoa(pageSize))
	if walletType != "" {
		params.Set("wallet_type", walletType)
	}

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result WalletTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取钱包交易记录失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
