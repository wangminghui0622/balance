<template>
  <div class="stores-page">
    <!-- 广告位BANNER -->
    <div class="banner-section">
      <div class="banner-title">广告位BANNER</div>
      <div class="banner-desc">字体占位符替换即可字体占位符替换即可</div>
    </div>

    <!-- 筛选和搜索区域 -->
    <div class="filter-section">
      <div class="filter-left">
        <el-input
          v-model="filterForm.keyword"
          placeholder="店铺名称/店铺编号"
          clearable
          class="filter-input"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select v-model="filterForm.shopStatus" placeholder="选择店铺状态" clearable class="filter-select">
          <el-option label="正常" value="normal" />
          <el-option label="暂停" value="paused" />
          <el-option label="关闭" value="closed" />
        </el-select>
        <el-select v-model="filterForm.authStatus" placeholder="选择授权状态" clearable class="filter-select">
          <el-option label="已授权" value="authorized" />
          <el-option label="未授权" value="unauthorized" />
        </el-select>
        <el-select v-model="filterForm.operationStatus" placeholder="选择运营状态" clearable class="filter-select">
          <el-option label="运营中" value="operating" />
          <el-option label="暂停" value="paused" />
        </el-select>
      </div>
      <div class="filter-right">
        <el-button type="primary" @click="handleQuery">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 店铺列表 -->
    <el-card class="stores-list-card" v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>店铺列表（{{ storeList.length }}）</span>
          <el-checkbox>导出报表</el-checkbox>
        </div>
      </template>
      <div class="stores-list">
        <div
          v-for="(store, index) in storeList"
          :key="index"
          class="store-item"
        >
          <!-- 左侧：头像+店铺名称+状态标签 -->
          <div class="store-left">
            <el-avatar :size="56" shape="square" :src="store.avatar" class="store-avatar" />
            <div class="store-info">
              <div class="store-name">{{ store.name }}</div>
              <div class="store-tags">
                <span class="tag-text">店铺状态</span>
                <span class="tag-text">授权状态</span>
                <span class="tag-text">运营状态</span>
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
                <div class="detail-label">店铺账户</div>
                <div class="detail-value">{{ store.shopAccount || '1234567890@gmail.com' }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">主体类型</div>
                <div class="detail-value">{{ store.entityType || '个人' }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">店铺健康</div>
                <div class="detail-value">{{ store.healthStatus || '占位符' }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">合作时间/天数</div>
                <div class="detail-value">{{ store.createTime }}</div>
              </div>
            </div>
          </div>

          <!-- 右侧：操作按钮 -->
          <div class="store-right">
            <el-button size="small" @click="handleMore(store)">更多</el-button>
            <el-button size="small" @click="handleManage(store)">经营</el-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <el-empty v-if="storeList.length === 0" description="暂无店铺数据" />
    </el-card>

    <!-- 店铺详情对话框 -->
    <el-dialog
      v-model="storeDetailVisible"
      width="500px"
      class="store-detail-dialog"
      :show-close="false"
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="dialog-header">
          <div class="header-left">
            <div class="dialog-store-name">{{ currentStore?.name }}</div>
            <div class="dialog-store-tags">
              <span class="tag-text">店铺状态</span>
              <span class="tag-text">授权状态</span>
              <span class="tag-text">运营状态</span>
            </div>
          </div>
          <el-avatar :size="56" shape="square" :src="currentStore?.avatar" class="dialog-avatar" />
        </div>
      </template>

      <div class="detail-tabs-wrapper">
        <div class="tab-buttons">
          <span 
            :class="['tab-btn', detailActiveTab === 'shop' ? 'active' : '']" 
            @click="detailActiveTab = 'shop'"
          >店铺信息</span>
          <span 
            :class="['tab-btn', detailActiveTab === 'cooperation' ? 'active' : '']" 
            @click="detailActiveTab = 'cooperation'"
          >合作信息</span>
        </div>

        <div v-if="detailActiveTab === 'shop'" class="detail-list">
          <div class="detail-row">
            <span class="detail-label">店铺编号：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">店铺区域：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">店铺类型：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">主体类型：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">主体名称：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">所属店主：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">店铺健康：</span>
            <span class="detail-value">S{{ currentStore?.storeId }}</span>
          </div>
        </div>

        <div v-else class="detail-list">
          <div class="detail-row">
            <span class="detail-label">所属店主：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">合作运营：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">合作编号：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">创建时间：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">合作开始时间：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">合作结束时间：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">合作剩余天数：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">履约方式：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">店主佣金比例：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">平台佣金比例：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">运营回款比例：</span>
            <span class="detail-value">示例文字占位符</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">备注信息：</span>
            <span class="detail-value">--</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">电子合约：</span>
            <span class="detail-value link">点击下载></span>
          </div>
          <div class="detail-row">
            <span class="detail-label">风险提示：</span>
            <span class="detail-value tip">平台作为第三机构，负责提供技术服务。不对损失承担责任，合作双方责任以双方约定为准。</span>
          </div>
        </div>
      </div>

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

      <!-- 近期回款区域 -->
      <div class="manage-income-section">
        <div class="income-header">
          <span class="income-title">
            <span class="title-bar"></span>
            <span>近期回款</span>
          </span>
          <el-button type="primary" link size="small" class="detail-button">
            查看详情
            <svg class="arrow-icon" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </el-button>
        </div>
        <div class="income-content">
          <div class="income-left">
            <div class="stat-item">
              <div class="stat-label">今日回款(NT$)</div>
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
            <div class="chart-container">
              <v-chart :option="manageChartOption" style="height: 200px; width: 100%;" autoresize />
            </div>
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
  shopAccount: string
  entityType: string
  healthStatus: string
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

const manageChartOption = ref({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'line',
      snap: true,
      lineStyle: {
        color: '#303133',
        width: 1,
        type: 'dashed'
      }
    },
    backgroundColor: '#fff',
    borderColor: '#e4e7ed',
    borderWidth: 1,
    padding: [8, 12],
    textStyle: {
      color: '#303133',
      fontSize: 12
    },
    formatter: (params: any) => {
      const first = Array.isArray(params) ? params[0] : params
      const value = first?.data ?? ''
      return `<div style="font-size:12px;color:#909399;">回款(NT$)</div><div style="font-size:16px;font-weight:600;color:#303133;">${Number(value).toLocaleString('zh-CN', { minimumFractionDigits: 2 })}</div>`
    }
  },
  grid: {
    left: '3%',
    right: '10%',
    top: '15%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: Array.from({ length: 30 }, (_, i) => i + 1),
    axisLine: {
      lineStyle: {
        color: '#e4e7ed'
      }
    },
    axisTick: {
      show: false
    },
    axisLabel: {
      color: '#909399'
    }
  },
  yAxis: {
    type: 'value',
    axisLine: {
      show: false
    },
    axisTick: {
      show: false
    },
    axisLabel: {
      show: false
    },
    splitLine: {
      lineStyle: {
        color: '#f0f0f0'
      }
    }
  },
  series: [
    {
      type: 'line',
      data: [
        1200, 1800, 1500, 2200, 1900, 2500, 2100, 2800, 3200, 2900,
        3500, 3100, 3800, 4200, 3900, 4500, 4100, 4800, 5200, 4900,
        5500, 5100, 5800, 6200, 5900, 6500, 6100, 6800, 7200, 8420
      ],
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      showSymbol: false,
      lineStyle: {
        color: '#303133',
        width: 2
      },
      itemStyle: {
        color: '#303133',
        borderColor: '#fff',
        borderWidth: 2
      },
      emphasis: {
        scale: true,
        itemStyle: {
          color: '#303133',
          borderColor: '#fff',
          borderWidth: 3,
          shadowColor: 'rgba(0, 0, 0, 0.3)',
          shadowBlur: 10
        }
      },
      areaStyle: {
        color: 'transparent'
      }
    }
  ]
})

const filterForm = reactive({
  keyword: '',
  shopStatus: '',
  authStatus: '',
  operationStatus: ''
})

const storeList = ref<Store[]>([])
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

function handleQuery() {
  fetchStores()
}

function handleReset() {
  filterForm.keyword = ''
  filterForm.shopStatus = ''
  filterForm.authStatus = ''
  filterForm.operationStatus = ''
  fetchStores()
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
    if (res.code === 0 && res.data && res.data.list.length > 0) {
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
        avatar: item.avatar || '',
        name: item.shopName,
        storeId: item.shopIdStr || String(item.shopId),
        shopAccount: item.shopAccount || '1234567890@gmail.com',
        entityType: item.entityType || '个人',
        healthStatus: item.healthStatus || '占位符',
        shopStatus: item.status === 1 ? 'normal' : (item.status === 0 ? 'closed' : 'paused'),
        operationStatus: item.authStatus === 1 ? 'operating' : 'paused',
        orderCount: item.totalOrders || 0,
        totalAmount: item.totalSales || 0,
        paymentAmount: 0,
        createTime: item.createdAt
      }))
      pagination.total = res.data.total
    } else {
      loadMockData()
    }
  } catch (error) {
    console.error('获取店铺列表失败', error)
    loadMockData()
  } finally {
    loading.value = false
  }
}

function loadMockData() {
  storeList.value = Array.from({ length: 6 }, (_, i) => ({
    id: i + 1,
    shopId: 51234567890 + i,
    shopIdStr: String(51234567890 + i),
    shopName: '店铺名称示例文字占位符文字占位符',
    adminId: 1,
    ownerName: '张三',
    region: 'TW',
    status: 1,
    authStatus: 1,
    totalOrders: 100,
    totalSales: 10000,
    createdAt: '2026-12-12 23:59:59',
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: String(51234567890 + i),
    shopAccount: '1234567890@gmail.com',
    entityType: i % 2 === 0 ? '个人' : '企业',
    healthStatus: '占位符',
    shopStatus: 'normal',
    operationStatus: 'operating',
    orderCount: 100,
    totalAmount: 10000,
    paymentAmount: 5000,
    createTime: '2026-12-12 23:59:59'
  }))
  pagination.total = 6
}

onMounted(() => {
  // 直接加载模拟数据
  loadMockData()
  loading.value = false
})
</script>

<style scoped lang="scss">
.stores-page {
  .banner-section {
    background: linear-gradient(135deg, #ffd93d 0%, #ffb347 100%);
    padding: 24px 32px;
    border-radius: 8px;
    margin-bottom: 20px;

    .banner-title {
      font-size: 20px;
      font-weight: 600;
      color: #303133;
      margin-bottom: 8px;
    }

    .banner-desc {
      font-size: 14px;
      color: #606266;
    }
  }

  .filter-section {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding: 16px 20px;
    background: #fff;
    border-radius: 8px;

    .filter-left {
      display: flex;
      gap: 12px;
      align-items: center;
    }

    .filter-right {
      display: flex;
      gap: 8px;
    }

    .filter-input {
      width: 180px;
    }

    .filter-select {
      width: 140px;
    }
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
    gap: 16px;
  }

  .store-item {
    display: flex;
    align-items: center;
    padding: 16px 20px;
    border: 1px solid #e4e7ed;
    border-radius: 8px;
    transition: all 0.3s;

    &:hover {
      border-color: #409eff;
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    }

    // 左侧A区域
    .store-left {
      display: flex;
      align-items: center;
      gap: 12px;
      width: 280px;
      flex-shrink: 0;

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
        min-width: 0;

        .store-name {
          font-size: 16px;
          font-weight: 500;
          color: #000000;
          line-height: 1.4;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .store-tags {
          display: flex;
          gap: 8px;
          flex-wrap: wrap;

          .tag-text {
            font-size: 12px;
            color: #f90;
          }
        }
      }
    }

    // 中间B区域
    .store-middle {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 0 20px;

      .store-details {
        display: flex;
        justify-content: space-between;
        width: 100%;
        max-width: 600px;

        .detail-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          font-size: 14px;
          line-height: 1.5;

          .detail-label {
            color: #606266;
            margin-bottom: 4px;
            white-space: nowrap;
            font-size: 13px;
            font-weight: 500;
            text-align: center;
          }

          .detail-value {
            color: #909399;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            text-align: center;
          }
        }
      }
    }

    // 右侧C区域
    .store-right {
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: flex-end;
      gap: 8px;
      flex-shrink: 0;
      min-width: 80px;
    }
  }
}

.store-detail-dialog {
  :deep(.el-dialog__header) {
    padding: 20px 20px 0;
    margin: 0;
  }

  :deep(.el-dialog__body) {
    padding: 0 20px 20px;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;

    .header-left {
      .dialog-store-name {
        font-size: 18px;
        font-weight: 500;
        color: #303133;
        margin-bottom: 8px;
      }

      .dialog-store-tags {
        display: flex;
        gap: 12px;

        .tag-text {
          font-size: 12px;
          color: #f90;
        }
      }
    }

    .dialog-avatar {
      background-color: #f5f7fa;
      border-radius: 4px;
    }
  }

  .detail-tabs-wrapper {
    .tab-buttons {
      display: flex;
      gap: 24px;
      margin-bottom: 20px;
      border-bottom: 1px solid #ebeef5;
      padding-bottom: 12px;

      .tab-btn {
        font-size: 14px;
        color: #909399;
        cursor: pointer;
        padding-bottom: 8px;
        border-bottom: 2px solid transparent;
        margin-bottom: -13px;

        &:hover {
          color: #303133;
        }

        &.active {
          color: #303133;
          font-weight: 500;
          border-bottom-color: #303133;
        }
      }
    }

    .detail-list {
      .detail-row {
        display: flex;
        padding: 12px 0;
        border-bottom: 1px solid #f5f5f5;

        &:last-child {
          border-bottom: none;
        }

        .detail-label {
          width: 100px;
          flex-shrink: 0;
          color: #909399;
          font-size: 14px;
          white-space: nowrap;
        }

        .detail-value {
          flex: 1;
          color: #303133;
          font-size: 14px;

          &.link {
            color: #409eff;
            cursor: pointer;
          }

          &.tip {
            color: #909399;
            font-size: 12px;
            line-height: 1.6;
          }
        }
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
          background: #303133;
          border-radius: 2px;
        }
      }

      .detail-button {
        display: flex;
        align-items: center;
        gap: 4px;

        .arrow-icon {
          width: 12px;
          height: 12px;
        }
      }
    }

    .income-content {
      display: flex;
      gap: 24px;

      .income-left {
        width: 120px;
        flex-shrink: 0;

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
        flex: 1;

        .chart-container {
          height: 200px;
          width: 100%;
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
