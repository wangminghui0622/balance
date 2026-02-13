package utils

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	IDInitialShopOwner int64 = 19906070668
	IDInitialOperator  int64 = 58608109796
	IDInitialPlatform  int64 = 91609051906

	IDIncrementShopOwnerMin int64 = 100
	IDIncrementShopOwnerMax int64 = 500

	IDIncrementOperatorMin int64 = 30
	IDIncrementOperatorMax int64 = 50

	IDIncrementPlatformMin int64 = 10
	IDIncrementPlatformMax int64 = 20

	FallbackIDPrefix  int64 = 8
	FallbackRandomMax int64 = 999
)

const generateIDScript = `
local current = redis.call('GET', KEYS[1])
if current == false then
	redis.call('SET', KEYS[1], ARGV[1])
	return tonumber(ARGV[1])
else
	local newID = tonumber(current) + tonumber(ARGV[2])
	redis.call('SET', KEYS[1], newID)
	return newID
end
`

// IDGenerator ID生成器
type IDGenerator struct {
	client        *redis.Client
	generateIDSha string
	shaMu         sync.Mutex
}

// NewIDGenerator 创建ID生成器
func NewIDGenerator(client *redis.Client) *IDGenerator {
	gen := &IDGenerator{client: client}
	sha, err := client.ScriptLoad(context.Background(), generateIDScript).Result()
	if err == nil {
		gen.generateIDSha = sha
	}
	return gen
}

func generateTimestampID() int64 {
	timestamp := time.Now().UnixMilli()
	random := rng.Int63n(FallbackRandomMax + 1)
	return FallbackIDPrefix + timestamp*1000 + random
}

func (g *IDGenerator) generateIDWithLua(ctx context.Context, key string, initialValue int64, increment int64) (int64, error) {
	var result interface{}
	var err error

	g.shaMu.Lock()
	cachedSha := g.generateIDSha
	g.shaMu.Unlock()

	if cachedSha != "" {
		result, err = g.client.EvalSha(ctx, cachedSha, []string{key}, initialValue, increment).Result()
		if err != nil && err.Error() == "NOSCRIPT No matching script. Use EVAL." {
			g.shaMu.Lock()
			g.generateIDSha = ""
			g.shaMu.Unlock()
		}
		if err == nil {
			id, ok := result.(int64)
			if !ok {
				return 0, fmt.Errorf("unexpected result type: %T", result)
			}
			return id, nil
		}
	}

	result, err = g.client.Eval(ctx, generateIDScript, []string{key}, initialValue, increment).Result()
	if err != nil {
		return generateTimestampID(), nil
	}

	g.shaMu.Lock()
	if g.generateIDSha == "" {
		sha, shaErr := g.client.ScriptLoad(ctx, generateIDScript).Result()
		if shaErr == nil {
			g.generateIDSha = sha
		}
	}
	g.shaMu.Unlock()

	id, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return id, nil
}

// GenerateShopOwnerID 生成店主ID
func (g *IDGenerator) GenerateShopOwnerID(ctx context.Context) (int64, error) {
	key := "id:generator:shopowner"
	rangeSize := IDIncrementShopOwnerMax - IDIncrementShopOwnerMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementShopOwnerMin
	return g.generateIDWithLua(ctx, key, IDInitialShopOwner, increment)
}

// GenerateOperatorID 生成运营ID
func (g *IDGenerator) GenerateOperatorID(ctx context.Context) (int64, error) {
	key := "id:generator:operator"
	rangeSize := IDIncrementOperatorMax - IDIncrementOperatorMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementOperatorMin
	return g.generateIDWithLua(ctx, key, IDInitialOperator, increment)
}

// GeneratePlatformID 生成平台ID
func (g *IDGenerator) GeneratePlatformID(ctx context.Context) (int64, error) {
	key := "id:generator:platform"
	rangeSize := IDIncrementPlatformMax - IDIncrementPlatformMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementPlatformMin
	return g.generateIDWithLua(ctx, key, IDInitialPlatform, increment)
}

// GenerateUserNo 生成用户编号
func GenerateUserNo(id int64) string {
	return fmt.Sprintf("U%011d", id)
}

// ==================== 业务ID生成（13位，增量30-90） ====================

const (
	// 业务ID初始值（13位，按业务类型区分前缀）
	IDInitialOrder            int64 = 1000000000000 // 订单相关 1xxx
	IDInitialOrderItem        int64 = 1100000000000
	IDInitialOrderAddress     int64 = 1200000000000
	IDInitialOrderEscrow      int64 = 1300000000000
	IDInitialOrderEscrowItem  int64 = 1400000000000
	IDInitialOrderSettlement  int64 = 1500000000000
	IDInitialShipmentRecord   int64 = 1600000000000
	IDInitialShipment         int64 = 1700000000000
	
	IDInitialFinanceIncome    int64 = 2000000000000 // 财务相关 2xxx
	IDInitialAccountTx        int64 = 2100000000000
	IDInitialWithdrawApp      int64 = 2200000000000
	IDInitialRechargeApp      int64 = 2300000000000
	
	IDInitialShop             int64 = 3000000000000 // 店铺相关 3xxx
	IDInitialShopAuth         int64 = 3100000000000
	IDInitialShopOperator     int64 = 3200000000000
	IDInitialShopSyncRecord   int64 = 3300000000000
	IDInitialProfitShareCfg   int64 = 3400000000000
	IDInitialLogisticsChannel int64 = 3500000000000
	IDInitialCollectionAcct   int64 = 3600000000000
	
	IDInitialPrepaymentAcct   int64 = 4000000000000 // 账户相关 4xxx
	IDInitialDepositAcct      int64 = 4100000000000
	IDInitialOperatorAcct     int64 = 4200000000000
	IDInitialShopOwnerCommAcct int64 = 4300000000000
	IDInitialPlatformCommAcct int64 = 4400000000000
	IDInitialPenaltyBonusAcct int64 = 4500000000000
	IDInitialEscrowAcct       int64 = 4600000000000
	
	IDInitialOperationLog     int64 = 5000000000000 // 日志相关 5xxx
	
	// 业务ID增量范围
	IDIncrementMin int64 = 30
	IDIncrementMax int64 = 90
)

// generateBusinessID 生成业务ID的通用方法
func (g *IDGenerator) generateBusinessID(ctx context.Context, key string, initialValue int64) (int64, error) {
	rangeSize := IDIncrementMax - IDIncrementMin + 1
	increment := rng.Int63n(rangeSize) + IDIncrementMin
	return g.generateIDWithLua(ctx, key, initialValue, increment)
}

// ==================== 订单相关ID ====================

func (g *IDGenerator) GenerateOrderID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:order", IDInitialOrder)
}

func (g *IDGenerator) GenerateOrderItemID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:order_item", IDInitialOrderItem)
}

func (g *IDGenerator) GenerateOrderAddressID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:order_address", IDInitialOrderAddress)
}

func (g *IDGenerator) GenerateOrderEscrowID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:order_escrow", IDInitialOrderEscrow)
}

func (g *IDGenerator) GenerateOrderEscrowItemID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:order_escrow_item", IDInitialOrderEscrowItem)
}

func (g *IDGenerator) GenerateOrderSettlementID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:order_settlement", IDInitialOrderSettlement)
}

func (g *IDGenerator) GenerateShipmentRecordID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shipment_record", IDInitialShipmentRecord)
}

func (g *IDGenerator) GenerateShipmentID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shipment", IDInitialShipment)
}

// ==================== 财务相关ID ====================

func (g *IDGenerator) GenerateFinanceIncomeID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:finance_income", IDInitialFinanceIncome)
}

func (g *IDGenerator) GenerateAccountTransactionID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:account_tx", IDInitialAccountTx)
}

func (g *IDGenerator) GenerateWithdrawApplicationID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:withdraw_app", IDInitialWithdrawApp)
}

func (g *IDGenerator) GenerateRechargeApplicationID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:recharge_app", IDInitialRechargeApp)
}

// ==================== 店铺相关ID ====================

func (g *IDGenerator) GenerateShopID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shop", IDInitialShop)
}

func (g *IDGenerator) GenerateShopAuthID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shop_auth", IDInitialShopAuth)
}

func (g *IDGenerator) GenerateShopOperatorRelationID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shop_operator", IDInitialShopOperator)
}

func (g *IDGenerator) GenerateShopSyncRecordID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shop_sync", IDInitialShopSyncRecord)
}

func (g *IDGenerator) GenerateProfitShareConfigID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:profit_share_cfg", IDInitialProfitShareCfg)
}

func (g *IDGenerator) GenerateLogisticsChannelID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:logistics_channel", IDInitialLogisticsChannel)
}

func (g *IDGenerator) GenerateCollectionAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:collection_acct", IDInitialCollectionAcct)
}

// ==================== 账户相关ID ====================

func (g *IDGenerator) GeneratePrepaymentAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:prepayment_acct", IDInitialPrepaymentAcct)
}

func (g *IDGenerator) GenerateDepositAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:deposit_acct", IDInitialDepositAcct)
}

func (g *IDGenerator) GenerateOperatorAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:operator_acct", IDInitialOperatorAcct)
}

func (g *IDGenerator) GenerateShopOwnerCommissionAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:shopowner_comm_acct", IDInitialShopOwnerCommAcct)
}

func (g *IDGenerator) GeneratePlatformCommissionAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:platform_comm_acct", IDInitialPlatformCommAcct)
}

func (g *IDGenerator) GeneratePenaltyBonusAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:penalty_bonus_acct", IDInitialPenaltyBonusAcct)
}

func (g *IDGenerator) GenerateEscrowAccountID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:escrow_acct", IDInitialEscrowAcct)
}

// ==================== 日志相关ID ====================

func (g *IDGenerator) GenerateOperationLogID(ctx context.Context) (int64, error) {
	return g.generateBusinessID(ctx, "id:gen:operation_log", IDInitialOperationLog)
}

// ==================== 批量生成ID ====================

// GenerateOrderIDs 批量生成订单ID
func (g *IDGenerator) GenerateOrderIDs(ctx context.Context, count int) ([]int64, error) {
	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		id, err := g.GenerateOrderID(ctx)
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}

// GenerateOrderItemIDs 批量生成订单商品ID
func (g *IDGenerator) GenerateOrderItemIDs(ctx context.Context, count int) ([]int64, error) {
	ids := make([]int64, count)
	for i := 0; i < count; i++ {
		id, err := g.GenerateOrderItemID(ctx)
		if err != nil {
			return nil, err
		}
		ids[i] = id
	}
	return ids, nil
}
