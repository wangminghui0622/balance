package shopee

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// EscrowDetailResponse 订单结算明细响应
type EscrowDetailResponse struct {
	BaseResponse
	Response struct {
		OrderSN     string `json:"order_sn"`
		BuyerUserID int64  `json:"buyer_user_id"`
		OrderIncome struct {
			EscrowAmount                    float64 `json:"escrow_amount"`
			BuyerTotalAmount                float64 `json:"buyer_total_amount"`
			OriginalPrice                   float64 `json:"original_price"`
			SellerDiscount                  float64 `json:"seller_discount"`
			ShopeeDiscount                  float64 `json:"shopee_discount"`
			VoucherFromSeller               float64 `json:"voucher_from_seller"`
			VoucherFromShopee               float64 `json:"voucher_from_shopee"`
			Coins                           float64 `json:"coins"`
			BuyerPaidShippingFee            float64 `json:"buyer_paid_shipping_fee"`
			BuyerTransactionFee             float64 `json:"buyer_transaction_fee"`
			CrossBorderTax                  float64 `json:"cross_border_tax"`
			PaymentPromotion                float64 `json:"payment_promotion"`
			CommissionFee                   float64 `json:"commission_fee"`
			ServiceFee                      float64 `json:"service_fee"`
			SellerTransactionFee            float64 `json:"seller_transaction_fee"`
			SellerLostCompensation          float64 `json:"seller_lost_compensation"`
			SellerCoinCashBack              float64 `json:"seller_coin_cash_back"`
			EscrowTax                       float64 `json:"escrow_tax"`
			FinalShippingFee                float64 `json:"final_shipping_fee"`
			ActualShippingFee               float64 `json:"actual_shipping_fee"`
			ShippingFeeDiscountFrom3PL      float64 `json:"shopee_shipping_rebate"`
			SellerShippingDiscount          float64 `json:"seller_shipping_discount"`
			EstimatedShippingFee            float64 `json:"estimated_shipping_fee"`
			SellerVoucherCode               []string `json:"seller_voucher_code"`
			DrcAdjustableRefund             float64 `json:"drc_adjustable_refund"`
			CostOfGoodsSold                 float64 `json:"cost_of_goods_sold"`
			OriginalCostOfGoodsSold         float64 `json:"original_cost_of_goods_sold"`
			OriginalShopeeDiscount          float64 `json:"original_shopee_discount"`
			SellerReturnRefund              float64 `json:"seller_return_refund"`
			ItemsCount                      int     `json:"items_count"`
			ReverseShippingFee              float64 `json:"reverse_shipping_fee"`
			FinalProductProtection          float64 `json:"final_product_protection"`
			CreditCardPromotion             float64 `json:"credit_card_promotion"`
			CreditCardTransactionFee        float64 `json:"credit_card_transaction_fee"`
		} `json:"order_income"`
		Items []struct {
			ItemID                    int64   `json:"item_id"`
			ItemName                  string  `json:"item_name"`
			ItemSKU                   string  `json:"item_sku"`
			ModelID                   int64   `json:"model_id"`
			ModelName                 string  `json:"model_name"`
			ModelSKU                  string  `json:"model_sku"`
			OriginalPrice             float64 `json:"original_price"`
			DiscountedPrice           float64 `json:"discounted_price"`
			SellerDiscount            float64 `json:"seller_discount"`
			ShopeeDiscount            float64 `json:"shopee_discount"`
			DiscountFromCoin          float64 `json:"discount_from_coin"`
			DiscountFromVoucher       float64 `json:"discount_from_voucher"`
			DiscountFromVoucherSeller float64 `json:"discount_from_voucher_seller"`
			DiscountFromVoucherShopee float64 `json:"discount_from_voucher_shopee"`
			ActivityType              string  `json:"activity_type"`
			ActivityID                int64   `json:"activity_id"`
			QuantityPurchased         int     `json:"quantity_purchased"`
		} `json:"items"`
	} `json:"response"`
}

// GetEscrowDetail 获取订单结算明细
func (c *Client) GetEscrowDetail(accessToken string, shopID uint64, orderSN string) (*EscrowDetailResponse, error) {
	path := "/api/v2/payment/get_escrow_detail"
	params := url.Values{}
	params.Set("order_sn", orderSN)

	respBody, err := c.Get(path, params, accessToken, shopID)
	if err != nil {
		return nil, err
	}

	var result EscrowDetailResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf("获取结算明细失败: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
