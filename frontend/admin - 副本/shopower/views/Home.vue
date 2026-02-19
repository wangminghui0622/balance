<template>
  <div class="shopowner-home">
    <!-- 顶部欢迎区域 -->
    <div class="welcome-section">
      <div class="welcome-info">
        <h2>早上好! 店主</h2>
        <el-row :gutter="12" class="welcome-status-row">
          <el-col class="welcome-left-col" :xs="24" :sm="24" :md="18" :lg="18" :xl="18">
            <p>今天店铺一切正常!有<span class="notice-count">4条</span>通知等您查询！</p>
          </el-col>
          <el-col class="welcome-right-col" :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
            <div class="notice-row">
              <el-alert
                class="important-notice"
                type="warning"
                show-icon
                :closable="false"
              >
                <template #title>
                  <span class="notice-title-prefix">重要通知：</span>
                  <span class="notice-title-text">通知示例文字占位符替换即可</span>
                </template>
                <template #icon>
                  <svg
                    class="notice-horn"
                    viewBox="0 0 24 24"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M4 10V14H7L11 17V7L7 10H4Z"
                      stroke="currentColor"
                      stroke-width="1.8"
                      stroke-linejoin="round"
                    />
                    <path
                      d="M14.5 9.8C15.2 10.5 15.6 11.2 15.6 12C15.6 12.8 15.2 13.5 14.5 14.2"
                      stroke="currentColor"
                      stroke-width="1.8"
                      stroke-linecap="round"
                    />
                    <path
                      d="M17 7.8C18.4 9.2 19.2 10.6 19.2 12C19.2 13.4 18.4 14.8 17 16.2"
                      stroke="currentColor"
                      stroke-width="1.8"
                      stroke-linecap="round"
                    />
                  </svg>
                </template>
              </el-alert>
            </div>
          </el-col>
        </el-row>
      </div>
    </div>

    <!-- 主要内容区域：左右两栏布局 -->
    <el-row :gutter="12" class="main-content-row">
      <!-- 左侧栏 -->
      <el-col class="main-left-col" :xs="24" :sm="24" :md="17" :lg="17" :xl="17">
        <div class="left-column">
          <!-- 上面：4个KPI统计卡片 -->
          <el-row :gutter="20" class="kpi-row">
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <ShopKPICard
                title="未结算订单金额(NT$)"
                :value="3246.0"
                subtitle="托管中订单: 6"
              />
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <ShopKPICard
                title="未结算佣金(NT$)"
                :value="608.5"
                subtitle="托管中的订单佣金"
              />
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <ShopKPICard
                title="已结算订单金额(NT$)"
                :value="9353636131.0"
                subtitle="已结算订单: 45"
              />
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <ShopKPICard
                title="已结算佣金(NT$)"
                :value="24543.0"
                subtitle="已结算的订单佣金"
              />
            </el-col>
          </el-row>

          <!-- 中间：近期收益 -->
          <div class="recent-income-section">
            <ShopRecentIncome />
          </div>

          <!-- 下面：最近订单 -->
          <div class="recent-orders-section">
            <ShopRecentOrders />
          </div>
        </div>
      </el-col>

      <!-- 右侧栏 -->
      <el-col class="main-right-col" :xs="24" :sm="24" :md="7" :lg="7" :xl="7">
        <div class="right-column">
          <!-- 上面：预存款余额 -->
          <div class="balance-section">
            <ShopPrestoreBalance />
          </div>

          <!-- 中间：我的店铺 -->
          <div class="stores-section">
            <ShopMyStores ref="shopMyStoresRef" />
          </div>

          <!-- 下面：近7日佣金排行榜 -->
          <div class="ranking-section">
            <ShopCommissionRanking />
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onActivated } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ShopKPICard from '../components/ShopKPICard.vue'
import ShopRecentIncome from '../components/ShopRecentIncome.vue'
import ShopPrestoreBalance from '../components/ShopPrestoreBalance.vue'
import ShopMyStores from '../components/ShopMyStores.vue'
import ShopRecentOrders from '../components/ShopRecentOrders.vue'
import ShopCommissionRanking from '../components/ShopCommissionRanking.vue'

const route = useRoute()
const router = useRouter()
const shopMyStoresRef = ref<InstanceType<typeof ShopMyStores> & { fetchShopList: (forceRefresh?: boolean) => Promise<void> } | null>(null)

// 刷新店铺列表的函数
const refreshShopList = async () => {
  await nextTick()
  if (shopMyStoresRef.value?.fetchShopList) {
    console.log('触发店铺列表强制刷新')
    // 传递 forceRefresh 参数，强制刷新并清空旧数据
    shopMyStoresRef.value.fetchShopList(true)
  }
  // 清除 refresh 参数（避免重复刷新）
  const newQuery = { ...route.query }
  delete newQuery.refresh
  router.replace({ query: newQuery })
}

// 监听路由参数，如果有 refresh 参数，则刷新店铺列表
watch(() => route.query.refresh, async (refresh) => {
  if (refresh === 'true') {
    await refreshShopList()
  }
}, { immediate: true })

// 组件挂载时检查是否需要刷新
onMounted(() => {
  if (route.query.refresh === 'true') {
    refreshShopList()
  }
})

// 组件激活时检查是否需要刷新（用于 keep-alive 场景）
onActivated(() => {
  if (route.query.refresh === 'true') {
    refreshShopList()
  }
})
</script>

<style scoped lang="scss">
.shopowner-home {
  .welcome-section {
    margin-bottom: 20px;
  }

  .welcome-info {
    margin-bottom: 16px;

    h2 {
      font-size: 20px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 8px;
    }

    p {
      font-size: 14px;
      color: #909399;
    }
  }

  .welcome-status-row {
    margin-top: 4px;

    p {
      margin: 0;
    }
  }

  .notice-count {
    color: #ff6a3a;
    font-weight: 600;
  }

  .notice-row {
    margin-top: 0;
  }

  .important-notice {
    width: 100%;
    max-width: none;
  }

  .notice-horn {
    width: 22px;
    height: 22px;
  }

  :deep(.important-notice.el-alert) {
    background: #ffffff;
    border: 1px solid rgba(255, 106, 58, 0.25);
    border-radius: 12px;
    overflow: hidden;
  }

  :deep(.important-notice .el-alert__description) {
    color: #303133;
  }

  .notice-title-prefix {
    color: #303133;
  }

  .notice-title-text {
    color: #909399;
  }

  :deep(.important-notice .el-alert__icon) {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    flex: 0 0 30px;
    line-height: 30px;
    border-radius: 0;
    background: transparent;
    color: #ff6a3a;
  }

  :deep(.important-notice .el-alert__icon svg) {
    width: 20px;
    height: 20px;
    display: block;
  }

  :deep(.important-notice .el-alert__content),
  :deep(.important-notice .el-alert__title) {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :deep(.important-notice .el-alert__content) {
    flex: 1;
  }

  .notification-banner {
    margin-bottom: 20px;
  }

  .main-content-row {
    margin-bottom: 20px;
  }

  .left-column {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .kpi-row {
    margin-bottom: 0;
  }

  .recent-income-section,
  .recent-orders-section {
    // 这些区域在 left-column 的 gap 中已经有间距了
  }

  .right-column {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .balance-section,
  .stores-section,
  .ranking-section {
    // 这些区域在 right-column 的 gap 中已经有间距了
  }

  .stores-section {
    margin-top: 4px;
  }

  .ranking-section {
    margin-top: 4px;
  }

  @media (min-width: 992px) {
    :deep(.welcome-left-col) {
      flex: 0 0 72.9166667%;
      max-width: 72.9166667%;
    }

    :deep(.welcome-right-col) {
      flex: 0 0 27.0833333%;
      max-width: 27.0833333%;
    }

    :deep(.main-left-col) {
      flex: 0 0 72.9166667%;
      max-width: 72.9166667%;
    }

    :deep(.main-right-col) {
      flex: 0 0 27.0833333%;
      max-width: 27.0833333%;
    }
  }
}
</style>
