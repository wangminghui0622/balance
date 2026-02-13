<template>
  <div class="settlement-page">
    <!-- 订单概览 -->
    <div class="order-summary">
      <div class="summary-header">
        <span class="summary-title">订单概览</span>
        <el-button type="primary" link>订单统计</el-button>
      </div>
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-label">未结算</div>
          <div class="card-value">
            <span class="value-count">{{ summaryData.unsettledCount }}单</span>
            <span class="value-separator">=</span>
            <span class="value-amount">NT${{ summaryData.unsettledAmount }}</span>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-label">已结算</div>
          <div class="card-value">
            <span class="value-count">{{ summaryData.settledCount }}单</span>
            <span class="value-separator">=</span>
            <span class="value-amount">NT${{ summaryData.settledAmount }}</span>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-label">账款调整</div>
          <div class="card-value">
            <span class="value-count">{{ summaryData.adjustCount }}单</span>
            <span class="value-separator">=</span>
            <span class="value-amount">NT${{ summaryData.adjustAmount }}</span>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-label">全部订单 ⓘ</div>
          <div class="card-value">
            <span class="value-count">{{ summaryData.totalCount }}单</span>
            <span class="value-separator">=</span>
            <span class="value-amount">NT${{ summaryData.totalAmount }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <div class="filter-row">
        <el-select v-model="filterForm.owner" placeholder="店主" clearable style="width: 100px">
          <el-option label="全部" value="" />
          <el-option label="店主A" value="ownerA" />
          <el-option label="店主B" value="ownerB" />
        </el-select>
        <el-select v-model="filterForm.operator" placeholder="运营" clearable style="width: 100px">
          <el-option label="全部" value="" />
          <el-option label="运营A" value="operatorA" />
          <el-option label="运营B" value="operatorB" />
        </el-select>
        <el-input
          v-model="filterForm.storeKeyword"
          placeholder="店铺名称/店铺编号"
          :prefix-icon="Search"
          clearable
          style="width: 200px"
        />
        <el-select v-model="filterForm.payStatus" placeholder="选择付款状态" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="已付款" value="paid" />
          <el-option label="未付款" value="unpaid" />
        </el-select>
        <el-select v-model="filterForm.settleStatus" placeholder="选择结算状态" clearable style="width: 140px">
          <el-option label="全部" value="" />
          <el-option label="已结算" value="settled" />
          <el-option label="未结算" value="unsettled" />
        </el-select>
        <el-date-picker
          v-model="filterForm.startDate"
          type="date"
          placeholder="开始日期"
          value-format="YYYY-MM-DD"
          style="width: 140px"
        />
        <span style="color: #909399;">-</span>
        <el-date-picker
          v-model="filterForm.endDate"
          type="date"
          placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 140px"
        />
        <div class="filter-actions">
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </div>
      </div>
      <div class="filter-row">
        <el-input
          v-model="filterForm.orderNo"
          placeholder="订单编号"
          :prefix-icon="Search"
          clearable
          style="width: 200px"
        />
      </div>
    </div>

    <!-- Tab和订单列表 -->
    <el-card class="order-card" shadow="never">
      <div class="order-header">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="全部订单" name="all" />
          <el-tab-pane label="未结算" name="unsettled" />
          <el-tab-pane label="已结算" name="settled" />
          <el-tab-pane label="账款调整" name="adjust" />
        </el-tabs>
        <div class="header-actions">
          <el-checkbox v-model="showProductInfo">商品信息</el-checkbox>
          <el-button :icon="Download">导出报表</el-button>
          <el-checkbox v-model="selectAll" @change="handleSelectAll">全选</el-checkbox>
          <el-button @click="handleAdjust">账款调整</el-button>
          <el-button @click="handleManualSettle">手工结算</el-button>
        </div>
      </div>

      <!-- 订单列表 -->
      <div class="order-list" v-loading="loading">
        <div class="order-item" v-for="(order, index) in orderList" :key="index">
          <div class="order-header-row">
            <div class="order-left">
              <el-checkbox v-model="order.selected" />
              <span class="order-label">订单编号：</span>
              <span class="order-no">{{ order.orderNo }}</span>
              <span class="pay-status">{{ order.payStatus }}</span>
            </div>
            <div class="order-right">
              <div class="order-avatar">J</div>
              <span class="commission-label">{{ order.commissionLabel }}：</span>
              <span class="commission-value">NT${{ order.commission }}</span>
              <span class="order-amount-label">订单金额：</span>
              <span class="order-amount">NT${{ order.amount }}</span>
            </div>
          </div>
          <div class="order-info-row">
            <span>下单时间：{{ order.orderTime }}</span>
            <span>店铺编号：{{ order.storeNo }}</span>
            <span>店铺名称：{{ order.storeName }}</span>
            <span>虾皮订单号：{{ order.shopeeOrderNo }}</span>
            <span class="shopee-status">虾皮订单状态：{{ order.shopeeStatus }}</span>
          </div>
          <div class="order-products" v-if="showProductInfo">
            <div class="product-item" v-for="(product, pIndex) in order.products" :key="pIndex">
              <div class="product-image"></div>
              <div class="product-info">
                <div class="product-name">{{ product.name }}</div>
                <div class="product-specs">颜色：{{ product.color }} 尺寸：{{ product.size }}</div>
              </div>
              <div class="product-price">单价：NT${{ product.price }}</div>
              <div class="product-qty">数量：{{ product.qty }}</div>
              <div class="product-subtotal">小计：{{ product.subtotal }}</div>
            </div>
          </div>
          <div class="order-footer">
            <span class="shopee-amount">虾皮订单金额：NT${{ order.shopeeAmount }}</span>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 手工结算弹框 -->
    <el-dialog
      v-model="manualSettleDialogVisible"
      title="手工结算"
      width="1000px"
      :close-on-click-modal="false"
    >
      <el-table :data="settleList" style="width: 100%" border>
        <el-table-column prop="id" label="编号" min-width="80" />
        <el-table-column prop="orderNo" label="订单编号" min-width="150" />
        <el-table-column prop="escrowAmount" label="托管金额" min-width="120" />
        <el-table-column label="店主佣金" min-width="140">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.ownerCommission" style="width: 50px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.ownerOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="平台佣金" min-width="140">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.platformCommission" style="width: 50px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.platformOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="运营回款" min-width="140">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.operatorCommission" style="width: 50px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.operatorOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleSettleRow(row)">结算</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="manualSettleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleExecuteSettle">执行结算</el-button>
      </template>
    </el-dialog>

    <!-- 账款调整弹框 -->
    <el-dialog
      v-model="adjustDialogVisible"
      title="账款调整"
      width="1100px"
      :close-on-click-modal="false"
    >
      <el-table :data="adjustList" style="width: 100%" border>
        <el-table-column prop="id" label="编号" min-width="70" />
        <el-table-column prop="orderNo" label="订单编号" min-width="140" />
        <el-table-column label="调整金额" min-width="110">
          <template #default="{ row }">
            <el-input v-model="row.adjustAmount" style="width: 100px" size="small" />
          </template>
        </el-table-column>
        <el-table-column label="店主佣金" min-width="120">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.ownerCommission" style="width: 45px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.ownerOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="平台佣金" min-width="120">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.platformCommission" style="width: 45px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.platformOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="运营回款" min-width="120">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.operatorCommission" style="width: 45px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.operatorOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="店主预付款" min-width="120">
          <template #default="{ row }">
            <div class="commission-input">
              <el-input v-model="row.ownerPrepay" style="width: 45px" size="small" />
              <span>%</span>
              <span class="original">原：{{ row.ownerPrepayOriginal }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="70" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleAdjustRow(row)">调整</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="adjustDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleExecuteAdjust">执行调整</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search, Download } from '@element-plus/icons-vue'

interface Product {
  name: string
  color: string
  size: string
  price: string
  qty: number
  subtotal: string
}

interface OrderRecord {
  orderNo: string
  payStatus: string
  orderTime: string
  storeNo: string
  storeName: string
  shopeeOrderNo: string
  shopeeStatus: string
  commissionLabel: string
  commission: string
  amount: string
  shopeeAmount: string
  products: Product[]
  selected: boolean
}

const activeTab = ref('all')
const loading = ref(false)
const showProductInfo = ref(true)
const selectAll = ref(false)
const manualSettleDialogVisible = ref(false)
const adjustDialogVisible = ref(false)

interface SettleRecord {
  id: string
  orderNo: string
  escrowAmount: string
  ownerCommission: string
  ownerOriginal: string
  platformCommission: string
  platformOriginal: string
  operatorCommission: string
  operatorOriginal: string
}

const settleList = ref<SettleRecord[]>([
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', escrowAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60' }
])

interface AdjustRecord {
  id: string
  orderNo: string
  adjustAmount: string
  ownerCommission: string
  ownerOriginal: string
  platformCommission: string
  platformOriginal: string
  operatorCommission: string
  operatorOriginal: string
  ownerPrepay: string
  ownerPrepayOriginal: string
}

const adjustList = ref<AdjustRecord[]>([
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' },
  { id: 'TP12345', orderNo: 'X250904KQ2P078R', adjustAmount: 'NT$3,560.50', ownerCommission: '60', ownerOriginal: '60', platformCommission: '60', platformOriginal: '60', operatorCommission: '60', operatorOriginal: '60', ownerPrepay: '60', ownerPrepayOriginal: '60' }
])

const summaryData = reactive({
  unsettledCount: 101,
  unsettledAmount: '38,420.00',
  settledCount: 12,
  settledAmount: '456.00',
  adjustCount: 12,
  adjustAmount: '456.00',
  totalCount: 150,
  totalAmount: '34.00'
})

const filterForm = reactive({
  owner: '',
  operator: '',
  storeKeyword: '',
  payStatus: '',
  settleStatus: '',
  startDate: '2025-09-01',
  endDate: '2025-09-10',
  orderNo: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const orderList = ref<OrderRecord[]>([])

const fetchOrderList = async () => {
  loading.value = true
  try {
    orderList.value = [
      {
        orderNo: 'X250904KQ2P078R',
        payStatus: '付款状态',
        orderTime: '2025-12-10 23:59:59',
        storeNo: 'S1234567890',
        storeName: '示例文字占位符示例文...',
        shopeeOrderNo: '250904KQ2P078R',
        shopeeStatus: '待发货',
        commissionLabel: '未结算佣金',
        commission: '8.00',
        amount: '36.00',
        shopeeAmount: '46.00',
        products: [
          { name: '商品名称示例文字占位符替换即可文字占位符替换即可', color: 'xxx', size: 'xxx', price: '46.00', qty: 1, subtotal: '46.00' }
        ],
        selected: false
      },
      {
        orderNo: 'X250904KQ2P078R',
        payStatus: '付款状态',
        orderTime: '2025-12-10 23:59:59',
        storeNo: 'S1234567890',
        storeName: '示例文字占位符示例文...',
        shopeeOrderNo: '250904KQ2P078R',
        shopeeStatus: '已完成',
        commissionLabel: '已结算佣金',
        commission: '16.00',
        amount: '198.00',
        shopeeAmount: '208.00',
        products: [
          { name: '商品名称示例文字占位符替换即可文字占位符替换即可', color: 'xxx', size: 'xxx', price: '46.00', qty: 2, subtotal: '92.00' },
          { name: '商品名称示例文字占位符替换即可文字占位符替换即可', color: 'xxx', size: 'xxx', price: '116.00', qty: 1, subtotal: '116.00' }
        ],
        selected: false
      },
      {
        orderNo: 'X250904KQ2P078R',
        payStatus: '付款状态',
        orderTime: '2025-12-10 23:59:59',
        storeNo: 'S1234567890',
        storeName: '示例文字占位符示例文...',
        shopeeOrderNo: '250904KQ2P078R',
        shopeeStatus: '待发货',
        commissionLabel: '未结算佣金',
        commission: '8.00',
        amount: '36.00',
        shopeeAmount: '46.00',
        products: [
          { name: '商品名称示例文字占位符替换即可文字占位符替换即可', color: 'xxx', size: 'xxx', price: '46.00', qty: 1, subtotal: '46.00' }
        ],
        selected: false
      }
    ]
    pagination.total = 123
  } catch (err) {
    console.error('获取订单列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  pagination.page = 1
  fetchOrderList()
}

const handleSearch = () => {
  pagination.page = 1
  fetchOrderList()
}

const handleReset = () => {
  filterForm.owner = ''
  filterForm.operator = ''
  filterForm.storeKeyword = ''
  filterForm.payStatus = ''
  filterForm.settleStatus = ''
  filterForm.startDate = '2025-09-01'
  filterForm.endDate = '2025-09-10'
  filterForm.orderNo = ''
  pagination.page = 1
  fetchOrderList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchOrderList()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchOrderList()
}

const handleSelectAll = (val: boolean) => {
  orderList.value.forEach(order => {
    order.selected = val
  })
}

const handleManualSettle = () => {
  manualSettleDialogVisible.value = true
}

const handleSettleRow = (row: SettleRecord) => {
  console.log('结算单行:', row)
}

const handleExecuteSettle = () => {
  manualSettleDialogVisible.value = false
}

const handleAdjust = () => {
  adjustDialogVisible.value = true
}

const handleAdjustRow = (row: AdjustRecord) => {
  console.log('调整单行:', row)
}

const handleExecuteAdjust = () => {
  adjustDialogVisible.value = false
}

onMounted(() => {
  fetchOrderList()
})
</script>

<style scoped lang="scss">
.settlement-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100%;
}

.order-summary {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  border: 1px solid #ebeef5;

  .summary-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .summary-title {
      font-size: 14px;
      font-weight: 500;
      color: #303133;
    }
  }

  .summary-cards {
    display: flex;
    gap: 16px;
  }

  .summary-card {
    flex: 1;
    padding: 16px;
    border: 1px solid #ebeef5;
    border-radius: 8px;

    .card-label {
      font-size: 12px;
      color: #909399;
      margin-bottom: 8px;
    }

    .card-value {
      display: flex;
      align-items: baseline;
      gap: 8px;

      .value-count {
        font-size: 24px;
        font-weight: 600;
        color: #303133;
      }

      .value-separator {
        color: #909399;
      }

      .value-amount {
        font-size: 24px;
        font-weight: 600;
        color: #303133;
      }
    }
  }
}

.filter-section {
  background: #fff;
  border-radius: 8px;
  padding: 16px 20px;
  margin-bottom: 20px;
  border: 1px solid #ebeef5;

  .filter-row {
    display: flex;
    gap: 12px;
    align-items: center;
    margin-bottom: 12px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .filter-actions {
    margin-left: auto;
    display: flex;
    gap: 12px;
  }
}

.order-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  :deep(.el-tabs) {
    .el-tabs__header {
      margin-bottom: 0;
    }

    .el-tabs__nav-wrap::after {
      display: none;
    }
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 16px;
  }
}

.order-list {
  padding: 0 20px;
}

.order-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  margin: 16px 0;
  overflow: hidden;

  .order-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: #fafafa;
    border-bottom: 1px solid #ebeef5;

    .order-left {
      display: flex;
      align-items: center;
      gap: 8px;

      .order-label {
        color: #606266;
        font-size: 14px;
      }

      .order-no {
        font-weight: 500;
        color: #303133;
      }

      .pay-status {
        font-size: 12px;
        color: #909399;
        margin-left: 8px;
      }
    }

    .order-right {
      display: flex;
      align-items: center;
      gap: 8px;

      .order-avatar {
        width: 24px;
        height: 24px;
        border-radius: 50%;
        background: #67c23a;
        color: #fff;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
      }

      .commission-label {
        color: #606266;
        font-size: 14px;
      }

      .commission-value {
        color: #303133;
        font-weight: 500;
      }

      .order-amount-label {
        color: #606266;
        font-size: 14px;
        margin-left: 16px;
      }

      .order-amount {
        color: #303133;
        font-weight: 500;
      }
    }
  }

  .order-info-row {
    display: flex;
    gap: 24px;
    padding: 12px 16px;
    font-size: 12px;
    color: #909399;
    border-bottom: 1px solid #ebeef5;

    .shopee-status {
      margin-left: auto;
    }
  }

  .order-products {
    padding: 12px 16px;

    .product-item {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;

      &:last-child {
        border-bottom: none;
      }

      .product-image {
        width: 60px;
        height: 60px;
        border-radius: 4px;
        background: #e0e0e0;
      }

      .product-info {
        flex: 1;

        .product-name {
          font-size: 14px;
          color: #303133;
          margin-bottom: 4px;
        }

        .product-specs {
          font-size: 12px;
          color: #909399;
        }
      }

      .product-price,
      .product-qty,
      .product-subtotal {
        font-size: 14px;
        color: #303133;
        min-width: 100px;
      }
    }
  }

  .order-footer {
    padding: 12px 16px;
    text-align: right;
    border-top: 1px solid #ebeef5;

    .shopee-amount {
      font-size: 14px;
      color: #303133;
    }
  }
}

.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.commission-input {
  display: flex;
  align-items: center;
  gap: 4px;

  .original {
    font-size: 12px;
    color: #909399;
  }
}
</style>
