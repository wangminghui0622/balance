<template>
  <div class="stores-page">
    <!-- 授权店铺BANNER -->
    <el-card class="auth-banner">
      <div class="banner-content">
        <div class="banner-left">
          <h3>授权店铺BANNER</h3>
          <p>授权你的店铺,开启合作之旅。</p>
        </div>
        <el-button type="default" size="large" class="auth-btn" @click="handleQuickAuth">
          马上授权
        </el-button>
      </div>
    </el-card>

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
        <el-form-item label="选择授权状态">
          <el-select v-model="filterForm.authStatus" placeholder="请选择" clearable class="filter-select-small">
            <el-option label="已授权" value="authorized" />
            <el-option label="未授权" value="unauthorized" />
            <el-option label="已过期" value="expired" />
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
    <el-card class="stores-list-card">
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
                <el-tag size="small" :type="getAuthStatusType(store.authStatus)">
                  {{ getAuthStatusText(store.authStatus) }}
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
                <div class="detail-label">店铺账户</div>
                <div class="detail-value">{{ store.account }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">主体类型</div>
                <div class="detail-value">{{ store.entityType }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">主体姓名</div>
                <div class="detail-value">{{ store.entityName }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">店铺健康</div>
                <div class="detail-value">{{ store.health }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">授权到期时间</div>
                <div class="detail-value">{{ store.expireTime }}</div>
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
            <el-tag size="small" :type="getAuthStatusType(currentStore?.authStatus || '')">
              {{ getAuthStatusText(currentStore?.authStatus || '') }}
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
              <span class="detail-label">店铺区域</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">店铺账号</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">主体类型</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">主体名称</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">所属店主</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">店铺健康</span>
              <span class="detail-value">{{ currentStore?.storeId }}</span>
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
              <div class="kpi-label">未结算佣金(NT$)</div>
              <div class="kpi-value">608.50</div>
              <div class="kpi-sub">托管中的订单佣金</div>
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
              <div class="kpi-label">已结算佣金(NT$)</div>
              <div class="kpi-value">24,543.00</div>
              <div class="kpi-sub">已结算的订单佣金</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 近期收益区域 -->
      <div class="manage-income-section">
        <div class="income-header">
          <span class="income-title">
            <span class="title-bar"></span>
            <span>近期收益</span>
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
              <div class="stat-label">今日销售(NT$)</div>
              <div class="stat-value">8,420.00</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">订单数量</div>
              <div class="stat-value">45</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">预估佣金(NT$)</div>
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
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { shopeeApi } from '@share/api/shopee'
import { STORAGE_KEYS, ROUTE_PATH, USER_TYPE_NUM, HTTP_STATUS } from '@share/constants'

interface Store {
  avatar: string
  name: string
  storeId: string
  shopId: number // Shopee Shop ID
  account: string
  entityType: string // 个人/企业
  entityName: string
  health: string
  expireTime: string
  shopStatus: string // normal/paused/closed
  authStatus: string // authorized/unauthorized/expired
  operationStatus: string // operating/paused
  authLoading?: boolean
}

const filterForm = reactive({
  keyword: '',
  shopStatus: '',
  authStatus: '',
  operationStatus: ''
})

const exportReport = ref(false)
const loading = ref(false)
const storeList = ref<Store[]>([])

// 检查是否为店主类型（userType=1）
const isShopOwner = () => {
  const userType = localStorage.getItem(STORAGE_KEYS.USER_TYPE)
  return userType === USER_TYPE_NUM.SHOPOWNER.toString()
}

// 获取店铺列表
const fetchShopList = async () => {
  // 只有店主类型才能调用此接口
  if (!isShopOwner()) {
    ElMessage.warning('此功能仅限店主类型用户使用')
    return
  }

  const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
  if (!token) {
    ElMessage.warning('请先登录')
    return
  }

  loading.value = true
  try {
    const res = await shopeeApi.getShopList()
    
    if (res.code === HTTP_STATUS.OK) {
      const shopList = res.data?.list || []
      if (shopList.length > 0) {
        // 使用完整的店铺数据
        storeList.value = shopList.map((shop: any) => {
          let authStatus = 'unauthorized'
          if (shop.authStatus === 1) authStatus = 'authorized'
          else if (shop.authStatus === 2) authStatus = 'expired'
          
          return {
            avatar: shop.shopSlug || '',
            name: shop.shopName || `店铺 ${shop.shopId}`,
            storeId: shop.shopIdStr || shop.shopId?.toString() || '',
            shopId: shop.shopId,
            account: shop.contactEmail || `shop_${shop.shopId}@example.com`,
            entityType: shop.isCbShop ? '企业' : '个人',
            entityName: shop.profile?.response?.shopName || `商家${shop.shopId}`,
            health: shop.ratingStar ? `${shop.ratingStar.toFixed(2)}分` : '待评估',
            expireTime: shop.expireTime || '未授权',
            shopStatus: shop.status === 1 ? 'normal' : 'paused',
            authStatus: authStatus,
            operationStatus: shop.status === 1 ? 'operating' : 'paused',
            authLoading: false
          }
        })
      } else {
        storeList.value = []
      }
    } else {
      ElMessage.error(res.message || '获取店铺列表失败')
    }
  } catch (err: any) {
    console.error('获取店铺列表失败:', err)
    ElMessage.error(err?.response?.data?.message || err?.message || '获取店铺列表失败')
  } finally {
    loading.value = false
  }
}

const route = useRoute()

// 处理授权成功后的绑定
const handleAuthSuccess = async () => {
  const authResult = route.query.auth
  const shopIdStr = route.query.shop_id as string
  
  if (authResult === 'success' && shopIdStr) {
    const shopId = parseInt(shopIdStr, 10)
    if (shopId > 0) {
      try {
        // 调用绑定接口将店铺绑定到当前用户
        const res = await shopeeApi.bindShop(shopId)
        if (res.code === HTTP_STATUS.OK) {
          ElMessage.success('店铺授权成功！')
        }
      } catch (err: any) {
        console.error('绑定店铺失败:', err)
        // 绑定失败不影响页面显示，可能店铺已经绑定
      }
    }
    // 清除 URL 参数
    router.replace({ path: route.path })
  } else if (authResult === 'failed') {
    const errorMsg = route.query.error as string
    ElMessage.error(errorMsg || '授权失败')
    router.replace({ path: route.path })
  }
}

// 组件挂载时获取店铺列表
onMounted(async () => {
  await handleAuthSuccess()
  fetchShopList()
})

const getShopStatusType = (status: string) => {
  const map: Record<string, string> = {
    normal: 'success',
    paused: 'warning',
    closed: 'danger'
  }
  return map[status] || 'info'
}

const getShopStatusText = (status: string) => {
  const map: Record<string, string> = {
    normal: '店铺状态',
    paused: '暂停',
    closed: '关闭'
  }
  return map[status] || status
}

const getAuthStatusType = (status: string) => {
  const map: Record<string, string> = {
    authorized: 'success',
    unauthorized: 'warning',
    expired: 'danger'
  }
  return map[status] || 'info'
}

const getAuthStatusText = (status: string) => {
  const map: Record<string, string> = {
    authorized: '授权状态',
    unauthorized: '未授权',
    expired: '已过期'
  }
  return map[status] || status
}

const getOperationStatusType = (status: string) => {
  const map: Record<string, string> = {
    operating: 'success',
    paused: 'warning'
  }
  return map[status] || 'info'
}

const getOperationStatusText = (status: string) => {
  const map: Record<string, string> = {
    operating: '运营状态',
    paused: '暂停'
  }
  return map[status] || status
}

const router = useRouter()

const handleQuickAuth = () => {
  // 检测登录态
  const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
  const userId = localStorage.getItem(STORAGE_KEYS.USER_ID)
  
  if (!token || !userId) {
    ElMessage.warning('请先登录后再进行授权操作')
    setTimeout(() => {
      router.push(ROUTE_PATH.LOGIN)
    }, 1500)
    return
  }

  ElMessage.info('快速授权功能开发中...')
}

const handleQuery = () => {
  ElMessage.success('查询功能开发中...')
}

const handleReset = () => {
  filterForm.keyword = ''
  filterForm.shopStatus = ''
  filterForm.authStatus = ''
  filterForm.operationStatus = ''
  ElMessage.success('已重置筛选条件')
}

const storeDetailVisible = ref(false)
const currentStore = ref<Store | null>(null)
const detailActiveTab = ref('shop')

const handleMore = (store: Store) => {
  currentStore.value = store
  detailActiveTab.value = 'shop'
  storeDetailVisible.value = true
}

const storeManageVisible = ref(false)
const currentManageStore = ref<Store | null>(null)

const manageChartOption = ref({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'line',
      snap: true,
      lineStyle: {
        color: '#ff6a3a',
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
      return `<div style="font-size:12px;color:#909399;">销售额(NT$)</div><div style="font-size:16px;font-weight:600;color:#ff6a3a;">${Number(value).toLocaleString('zh-CN', { minimumFractionDigits: 2 })}</div>`
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
        color: '#909399'
      }
    },
    axisTick: {
      lineStyle: {
        color: 'rgba(255, 106, 58, 0.25)'
      }
    },
    axisLabel: {
      color: '#909399'
    }
  },
  yAxis: {
    type: 'value',
    name: '销售(NT$)',
    nameLocation: 'end',
    nameGap: 10,
    nameTextStyle: {
      color: '#909399',
      fontSize: 12,
      fontWeight: 500
    },
    axisLine: {
      show: true,
      lineStyle: {
        color: '#dcdfe6'
      }
    },
    axisTick: {
      show: false
    },
    axisLabel: {
      show: false,
      color: '#909399'
    },
    splitLine: {
      lineStyle: {
        color: 'rgba(255, 106, 58, 0.08)'
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
        color: '#ff6a3a',
        width: 2
      },
      itemStyle: {
        color: '#ff6a3a',
        borderColor: '#fff',
        borderWidth: 2
      },
      emphasis: {
        scale: true,
        itemStyle: {
          color: '#ff6a3a',
          borderColor: '#fff',
          borderWidth: 3,
          shadowColor: 'rgba(255, 106, 58, 0.5)',
          shadowBlur: 10
        }
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(255, 106, 58, 0.35)' },
            { offset: 1, color: 'rgba(255, 106, 58, 0.05)' }
          ]
        }
      }
    }
  ]
})

const handleManage = (store: Store) => {
  currentManageStore.value = store
  storeManageVisible.value = true
}

const handleExportReport = () => {
  if (exportReport.value) {
    ElMessage.success('正在导出报表...')
  } else {
    ElMessage.info('请先勾选"导出报表"')
  }
}
</script>

<style scoped lang="scss">
.stores-page {
  padding: 20px;
}

.auth-banner {
  margin-bottom: 20px;
  background: linear-gradient(135deg, #ff6a3a 0%, #ff8c5a 100%);
  border: none;

  :deep(.el-card__body) {
    padding: 30px;
  }

  .banner-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    color: white;

    .banner-left {
      h3 {
        margin: 0 0 10px 0;
        font-size: 24px;
        font-weight: 600;
      }

      p {
        margin: 0;
        font-size: 16px;
        opacity: 0.9;
      }
    }

    .auth-btn {
      border: 1px solid rgba(255, 255, 255, 0.9);
      background: transparent;
      color: rgba(255, 255, 255, 0.95);
      border-radius: 6px;

      &:hover {
        border-color: rgba(255, 255, 255, 1);
        background-color: rgba(255, 255, 255, 0.08);
      }
    }
  }
}

.filter-card {
  margin-bottom: 20px;

  .filter-form {
    margin-bottom: 0;
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    justify-content: flex-start;
    gap: 16px;

    .el-form-item {
      flex-shrink: 0;
    }

    .el-form-item:last-child {
      margin-left: auto;
      margin-right: -3px;
      flex-shrink: 0;
    }
  }

  .filter-input {
    width: 25ch;
  }

  .filter-select-small {
    width: 13ch;
  }

  /* 强制下拉框和输入框圆角 */
  :deep(.el-input__wrapper),
  :deep(.el-input__inner),
  :deep(.el-select .el-select__wrapper) {
    border-radius: 30px !important;
    overflow: hidden;
  }
}

/* 下拉弹出框圆角（全局样式） */
:global(.el-select__popper.el-popper) {
  border-radius: 8px !important;
  overflow: hidden;
}

:global(.el-select-dropdown) {
  border-radius: 8px !important;
  overflow: hidden;
}

:global(.el-select-dropdown__list) {
  border-radius: 8px !important;
}

/* 下拉选项内容居中 */
:global(.el-select-dropdown__item) {
  text-align: center;
  justify-content: center;
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
    padding: 8px 0;
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
      flex: 1;
      min-width: 0;
      margin-left: 10px;

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

    // 中间B区域
    .store-middle {
      flex: 1;
      min-width: 0;
      display: flex;
      align-items: center;
      margin-right: -230px;

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
            color: #c0c4cc;
            text-align: center;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            width: 100%;
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
      gap: 12px;
      flex: 1;
      min-width: 0;
      margin-right: 10px;

      .action-btn {
        min-width: 80px;
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

.store-detail-dialog {
  :deep(.el-dialog__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #e4e7ed;
    margin-right: 0;

    .el-dialog__title {
      font-size: 16px;
      font-weight: 500;
    }
  }

  :deep(.el-dialog__body) {
    padding: 0;
  }

  :deep(.el-dialog__footer) {
    padding: 12px 20px;
    border-top: 1px solid #e4e7ed;
  }

  .dialog-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 20px;
    background-color: #fff;
    border-bottom: 1px solid #f0f0f0;

    .dialog-avatar {
      flex-shrink: 0;
      background-color: #ff6a3a;
      border-radius: 8px;
    }

    .dialog-store-info {
      flex: 1;

      .dialog-store-name {
        font-size: 16px;
        font-weight: 500;
        color: #303133;
        margin-bottom: 8px;
      }

      .dialog-store-tags {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
      }
    }
  }

  .detail-tabs {
    :deep(.el-tabs__header) {
      margin: 0;
      padding: 0 20px;
      background-color: #fff;
    }

    :deep(.el-tabs__nav-wrap::after) {
      height: 1px;
    }

    :deep(.el-tabs__content) {
      padding: 0;
    }
  }

  .detail-list {
    padding: 16px 20px;
    background-color: #fff;

    .detail-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 0;
      border-bottom: 1px solid #f5f5f5;

      &:last-child {
        border-bottom: none;
      }

      .detail-label {
        font-size: 14px;
        color: #606266;
      }

      .detail-value {
        font-size: 14px;
        color: #303133;
      }
    }
  }
}

.store-manage-dialog {
  :deep(.el-dialog__header) {
    padding: 16px 20px;
    border-bottom: 1px solid #e4e7ed;
    margin-right: 0;

    .el-dialog__title {
      font-size: 16px;
      font-weight: 500;
    }
  }

  :deep(.el-dialog__body) {
    padding: 20px;
    background-color: #f5f7fa;
  }

  .manage-kpi-section {
    margin-bottom: 20px;

    .kpi-card {
      background-color: #fff;
      border-radius: 8px;
      padding: 16px;
      height: 100%;

      .kpi-label {
        font-size: 12px;
        color: #909399;
        margin-bottom: 8px;
      }

      .kpi-value {
        font-size: 24px;
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
    background-color: #fff;
    border-radius: 8px;
    padding: 16px;

    .income-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .income-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 14px;
        font-weight: 500;
        color: #303133;

        .title-bar {
          width: 4px;
          height: 14px;
          background-color: #ff6a3a;
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

          &:last-child {
            margin-bottom: 0;
          }

          .stat-label {
            font-size: 12px;
            color: #909399;
            margin-bottom: 4px;
          }

          .stat-value {
            font-size: 20px;
            font-weight: 600;
            color: #ff6a3a;
          }
        }
      }

      .income-right {
        flex: 1;
        min-width: 0;

        .chart-container {
          height: 200px;
          width: 100%;
        }
      }
    }
  }
}
</style>
