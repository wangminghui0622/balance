<template>
  <div class="stores-page">
    <!-- 筛选和搜索区域 -->
    <el-card class="filter-card">
      <el-form :model="filterForm" :inline="true" class="filter-form">
        <el-form-item label="店铺名称/店铺编号">
          <el-input
            v-model="filterForm.keyword"
            placeholder="请输入店铺名称或编号"
            clearable
            class="filter-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="选择店铺状态">
          <el-select v-model="filterForm.shopStatus" placeholder="请选择" clearable class="filter-select-small">
            <el-option label="正常" value="normal" />
            <el-option label="暂停" value="paused" />
            <el-option label="关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择运营状态">
          <el-select v-model="filterForm.operationStatus" placeholder="请选择" clearable class="filter-select-small">
            <el-option label="运营中" value="operating" />
            <el-option label="暂停" value="paused" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 店铺列表 -->
    <el-card class="stores-list-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>店铺列表 ({{ storeList.length }})</span>
          <el-button type="default" size="small" @click="handleExportReport">
            导出报表
          </el-button>
        </div>
      </template>
      <div class="stores-list">
        <div
          v-for="(store, index) in storeList"
          :key="index"
          class="store-item"
        >
          <!-- 左侧A区域 -->
          <div class="store-left">
            <el-avatar :size="72" shape="square" :src="store.avatar" class="store-avatar" />
            <div class="store-info">
              <div class="store-name">{{ store.name }}</div>
              <div class="store-tags">
                <el-tag size="small" :type="getShopStatusType(store.shopStatus)">
                  {{ getShopStatusText(store.shopStatus) }}
                </el-tag>
                <el-tag size="small" :type="getOperationStatusType(store.operationStatus)">
                  {{ getOperationStatusText(store.operationStatus) }}
                </el-tag>
              </div>
            </div>
          </div>

          <!-- 中间B区域 -->
          <div class="store-middle">
            <div class="store-details">
              <div class="detail-item">
                <div class="detail-label">店铺编号</div>
                <div class="detail-value">{{ store.storeId }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">店主名称</div>
                <div class="detail-value">{{ store.ownerName }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">订单数量</div>
                <div class="detail-value">{{ store.orderCount }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">累计金额</div>
                <div class="detail-value">NT${{ formatAmount(store.totalAmount) }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">累计回款</div>
                <div class="detail-value">NT${{ formatAmount(store.paymentAmount) }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">绑定时间</div>
                <div class="detail-value">{{ store.createTime }}</div>
              </div>
            </div>
          </div>

          <!-- 右侧C区域 -->
          <div class="store-right">
            <el-button size="small" @click="handleMore(store)" class="action-btn">更多</el-button>
            <el-button type="success" size="small" @click="handleManage(store)" class="action-btn">经营</el-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <el-empty v-if="storeList.length === 0" description="暂无店铺数据" />
    </el-card>

    <!-- 店铺详情对话框 -->
    <el-dialog
      v-model="storeDetailVisible"
      :title="currentStore?.name || '店铺详情'"
      width="500px"
      class="store-detail-dialog"
      :close-on-click-modal="false"
    >
      <div class="dialog-header">
        <el-avatar :size="48" shape="square" :src="currentStore?.avatar" class="dialog-avatar" />
        <div class="dialog-store-info">
          <div class="dialog-store-name">{{ currentStore?.name }}</div>
          <div class="dialog-store-tags">
            <el-tag size="small" :type="getShopStatusType(currentStore?.shopStatus || '')">
              {{ getShopStatusText(currentStore?.shopStatus || '') }}
            </el-tag>
            <el-tag size="small" :type="getOperationStatusType(currentStore?.operationStatus || '')">
              {{ getOperationStatusText(currentStore?.operationStatus || '') }}
            </el-tag>
          </div>
        </div>
      </div>

      <el-tabs v-model="detailActiveTab" class="detail-tabs">
        <el-tab-pane label="店铺信息" name="shop">
          <div class="detail-list">
            <div class="detail-row">
              <span class="detail-label">店铺编号</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">店主名称</span>
              <span class="detail-value">{{ currentStore?.ownerName }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">订单数量</span>
              <span class="detail-value">{{ currentStore?.orderCount }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">累计金额</span>
              <span class="detail-value">NT${{ formatAmount(currentStore?.totalAmount || 0) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">累计回款</span>
              <span class="detail-value">NT${{ formatAmount(currentStore?.paymentAmount || 0) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">绑定时间</span>
              <span class="detail-value">{{ currentStore?.createTime }}</span>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="合作信息" name="cooperation">
          <div class="detail-list">
            <div class="detail-row">
              <span class="detail-label">合作信息</span>
              <span class="detail-value">暂无数据</span>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button type="primary" @click="storeDetailVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 店铺经营数据对话框 -->
    <el-dialog
      v-model="storeManageVisible"
      :title="(currentManageStore?.name || '店铺') + '经营数据'"
      width="900px"
      class="store-manage-dialog"
      :close-on-click-modal="false"
    >
      <!-- KPI卡片区域 -->
      <div class="manage-kpi-section">
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="kpi-card">
              <div class="kpi-label">未结算订单金额(NT$)</div>
              <div class="kpi-value">3,246.00</div>
              <div class="kpi-sub">托管中订单：6</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="kpi-card">
              <div class="kpi-label">未结算回款(NT$)</div>
              <div class="kpi-value">608.50</div>
              <div class="kpi-sub">托管中的订单回款</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="kpi-card">
              <div class="kpi-label">已结算订单金额(NT$)</div>
              <div class="kpi-value">1,353,636.00</div>
              <div class="kpi-sub">已结算订单：45</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="kpi-card">
              <div class="kpi-label">已结算回款(NT$)</div>
              <div class="kpi-value">24,543.00</div>
              <div class="kpi-sub">已结算的订单回款</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 近期收益区域 -->
      <div class="manage-income-section">
        <div class="income-header">
          <span class="income-title">
            <span class="title-bar"></span>
            <span>近期回款</span>
          </span>
          <el-button type="primary" link size="small" class="detail-button">
            查看详情
          </el-button>
        </div>
        <div class="income-content">
          <div class="income-left">
            <div class="stat-item">
              <div class="stat-label">今日销售(NT$)</div>
              <div class="stat-value">8,420.00</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">订单数量</div>
              <div class="stat-value">45</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">预估回款(NT$)</div>
              <div class="stat-value">540</div>
            </div>
          </div>
          <div class="income-right">
            <div class="chart-placeholder">图表区域</div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { operatorShopeeApi } from '@share/api/shopee'

interface Store {
  id: number
  shopId: number
  shopIdStr: string
  shopName: string
  adminId: number
  ownerName: string
  region: string
  status: number
  authStatus: number
  totalOrders: number
  totalSales: number
  createdAt: string
  // 兼容模板使用的别名
  avatar: string
  name: string
  storeId: string
  shopStatus: string
  operationStatus: string
  orderCount: number
  totalAmount: number
  paymentAmount: number
  createTime: string
}

const loading = ref(false)
const storeDetailVisible = ref(false)
const storeManageVisible = ref(false)
const detailActiveTab = ref('shop')
const currentStore = ref<Store | null>(null)
const currentManageStore = ref<Store | null>(null)

const filterForm = reactive({
  keyword: '',
  shopStatus: '',
  operationStatus: ''
})

const storeList = ref<Store[]>([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

function formatAmount(value: number): string {
  return value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function getShopStatusType(status: string): string {
  const map: Record<string, string> = {
    normal: 'success',
    paused: 'warning',
    closed: 'danger'
  }
  return map[status] || 'info'
}

function getShopStatusText(status: string): string {
  const map: Record<string, string> = {
    normal: '正常',
    paused: '暂停',
    closed: '关闭'
  }
  return map[status] || '未知'
}

function getOperationStatusType(status: string): string {
  const map: Record<string, string> = {
    operating: 'success',
    paused: 'warning'
  }
  return map[status] || 'info'
}

function getOperationStatusText(status: string): string {
  const map: Record<string, string> = {
    operating: '运营中',
    paused: '暂停'
  }
  return map[status] || '未知'
}

function handleQuery() {
  fetchStores()
}

function handleReset() {
  filterForm.keyword = ''
  filterForm.shopStatus = ''
  filterForm.operationStatus = ''
  fetchStores()
}

function handleExportReport() {
  console.log('导出报表')
}

function handleMore(store: Store) {
  currentStore.value = store
  storeDetailVisible.value = true
}

function handleManage(store: Store) {
  currentManageStore.value = store
  storeManageVisible.value = true
}

async function fetchStores() {
  loading.value = true
  try {
    const res = await operatorShopeeApi.getShopList({
      keyword: filterForm.keyword,
      status: filterForm.shopStatus,
      page: pagination.page,
      page_size: pagination.pageSize
    })
    if (res.code === 0 && res.data) {
      storeList.value = res.data.list.map((item: any) => ({
        id: item.id,
        shopId: item.shopId,
        shopIdStr: item.shopIdStr || String(item.shopId),
        shopName: item.shopName,
        adminId: item.adminId,
        ownerName: item.ownerName || '-',
        region: item.region,
        status: item.status,
        authStatus: item.authStatus,
        totalOrders: item.totalOrders || 0,
        totalSales: item.totalSales || 0,
        createdAt: item.createdAt,
        // 兼容模板使用的别名
        avatar: item.avatar || '',
        name: item.shopName,
        storeId: item.shopIdStr || String(item.shopId),
        shopStatus: item.status === 1 ? 'normal' : (item.status === 0 ? 'closed' : 'paused'),
        operationStatus: item.authStatus === 1 ? 'operating' : 'paused',
        orderCount: item.totalOrders || 0,
        totalAmount: item.totalSales || 0,
        paymentAmount: 0,
        createTime: item.createdAt
      }))
      pagination.total = res.data.total
    }
  } catch (error) {
    console.error('获取店铺列表失败', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStores()
})
</script>

<style scoped lang="scss">
.stores-page {
  .filter-card {
    margin-bottom: 20px;
  }

  .filter-input {
    width: 220px;
  }

  .filter-select-small {
    width: 120px;
  }

  :deep(.el-input__wrapper),
  :deep(.el-select .el-select__wrapper) {
    border-radius: 30px !important;
  }
}

.stores-list-card {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 500;
    font-size: 16px;
  }

  .stores-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .store-item {
    display: flex;
    align-items: stretch;
    gap: 20px;
    padding: 16px;
    border: 1px solid #e4e7ed;
    border-radius: 8px;
    transition: all 0.3s;

    &:hover {
      border-color: #409eff;
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    }

    .store-left {
      display: flex;
      align-items: center;
      gap: 12px;
      flex: 1;
      min-width: 0;

      .store-avatar {
        flex-shrink: 0;
        background-color: #f5f7fa;
        border-radius: 8px;
      }

      .store-info {
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: 10px;
        flex: 1;
        min-width: 0;

        .store-name {
          font-size: 16px;
          font-weight: 500;
          color: #000000;
          line-height: 1.4;
        }

        .store-tags {
          display: flex;
          gap: 8px;
          flex-wrap: wrap;
        }
      }
    }

    .store-middle {
      flex: 2;
      min-width: 0;
      display: flex;
      align-items: center;

      .store-details {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 8px 32px;
        width: 100%;

        .detail-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          font-size: 14px;
          line-height: 1.5;

          .detail-label {
            color: #606266;
            margin-bottom: 2px;
            white-space: nowrap;
            font-size: 13px;
            font-weight: 500;
          }

          .detail-value {
            color: #909399;
            text-align: center;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            width: 100%;
          }
        }
      }
    }

    .store-right {
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: flex-end;
      gap: 12px;
      min-width: 80px;

      .action-btn {
        min-width: 80px;
      }
    }
  }
}

.store-detail-dialog {
  .dialog-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;

    .dialog-store-info {
      .dialog-store-name {
        font-size: 16px;
        font-weight: 500;
        margin-bottom: 8px;
      }

      .dialog-store-tags {
        display: flex;
        gap: 8px;
      }
    }
  }

  .detail-list {
    .detail-row {
      display: flex;
      justify-content: space-between;
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;

      .detail-label {
        color: #606266;
      }

      .detail-value {
        color: #303133;
      }
    }
  }
}

.store-manage-dialog {
  .manage-kpi-section {
    margin-bottom: 24px;

    .kpi-card {
      background: #f8f9fa;
      border-radius: 8px;
      padding: 16px;
      text-align: center;

      .kpi-label {
        font-size: 12px;
        color: #909399;
        margin-bottom: 8px;
      }

      .kpi-value {
        font-size: 20px;
        font-weight: 600;
        color: #303133;
        margin-bottom: 4px;
      }

      .kpi-sub {
        font-size: 12px;
        color: #c0c4cc;
      }
    }
  }

  .manage-income-section {
    .income-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .income-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 16px;
        font-weight: 500;

        .title-bar {
          width: 4px;
          height: 16px;
          background: #ff6a3a;
          border-radius: 2px;
        }
      }
    }

    .income-content {
      display: flex;
      gap: 24px;

      .income-left {
        flex: 1;

        .stat-item {
          margin-bottom: 16px;

          .stat-label {
            font-size: 12px;
            color: #909399;
            margin-bottom: 4px;
          }

          .stat-value {
            font-size: 24px;
            font-weight: 600;
            color: #303133;
          }
        }
      }

      .income-right {
        flex: 2;

        .chart-placeholder {
          height: 200px;
          background: #f5f7fa;
          border-radius: 8px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: #909399;
        }
      }
    }
  }
}

@media (max-width: 1200px) {
  .store-item {
    .store-middle {
      .store-details {
        grid-template-columns: repeat(2, 1fr);
      }
    }
  }
}

@media (max-width: 768px) {
  .store-item {
    flex-direction: column;

    .store-right {
      width: 100%;
      flex-direction: row;
      justify-content: flex-end;
    }

    .store-middle {
      .store-details {
        grid-template-columns: 1fr;
      }
    }
  }
}
</style>
