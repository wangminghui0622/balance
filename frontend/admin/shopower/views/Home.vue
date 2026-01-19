<template>
  <div class="shopowner-home">
    <!-- 顶部欢迎和通知区域 -->
    <div class="welcome-section">
      <div class="welcome-info">
        <h2>早上好! 店主</h2>
        <p>今天店铺一切正常! 有4条通知等您查阅。</p>
      </div>
      <el-alert
        v-if="showNotification"
        :closable="true"
        type="info"
        class="notification-banner"
        @close="showNotification = false"
      >
        <template #default>
          <span>重要通知,待处理信息占位符文字占位符文字</span>
        </template>
      </el-alert>
    </div>

    <!-- 主要内容区域：左右两栏布局 -->
    <el-row :gutter="20" class="main-content-row">
      <!-- 左侧栏 -->
      <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="16">
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
                :value="1353636.0"
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
      <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
        <div class="right-column">
          <!-- 上面：预存款余额 -->
          <div class="balance-section">
            <ShopPrestoreBalance />
          </div>

          <!-- 中间：我的店铺 -->
          <div class="stores-section">
            <ShopMyStores />
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
import { ref } from 'vue'
import ShopKPICard from '../components/ShopKPICard.vue'
import ShopRecentIncome from '../components/ShopRecentIncome.vue'
import ShopPrestoreBalance from '../components/ShopPrestoreBalance.vue'
import ShopMyStores from '../components/ShopMyStores.vue'
import ShopRecentOrders from '../components/ShopRecentOrders.vue'
import ShopCommissionRanking from '../components/ShopCommissionRanking.vue'

const showNotification = ref(true)
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

  .notification-banner {
    margin-bottom: 20px;
  }

  .main-content-row {
    margin-bottom: 20px;
  }

  .left-column {
    display: flex;
    flex-direction: column;
    gap: 20px;
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
    gap: 20px;
  }

  .balance-section,
  .stores-section,
  .ranking-section {
    // 这些区域在 right-column 的 gap 中已经有间距了
  }
}
</style>
