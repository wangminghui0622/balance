<template>
  <div class="operator-home">
    <!-- 顶部欢迎区域 -->
    <div class="welcome-section">
      <div class="welcome-left">
        <div class="welcome-info">
          <h2>早上好！运营</h2>
          <p>今天店铺一切正常！<span class="notification-link">有4条通知等您查阅。</span></p>
        </div>
      </div>
      <div class="welcome-right">
        <div class="notification-banner">
          <span class="notification-text">重要通知，将发理信息占位符文字占位符文字</span>
          <el-button type="text" class="close-btn">×</el-button>
        </div>
      </div>
    </div>

    <!-- 主要内容区域：左右两栏布局 -->
    <el-row :gutter="20" class="main-content-row">
      <!-- 左侧栏 -->
      <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="16">
        <div class="left-column">
          <!-- 上面：4个KPI统计卡片 -->
          <el-row :gutter="20" class="kpi-row">
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <KPICard
                title="未结算订单金额(NT$)"
                :value="3246.00"
                subtitle="未结算订单: 6"
              />
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <KPICard
                title="未结算回款(NT$)"
                :value="608.50"
                subtitle="订单: 1"
              />
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <KPICard
                title="已结算订单金额(NT$)"
                :value="1353636.00"
                subtitle="已结算订单: 45"
              />
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="6" :xl="6">
              <KPICard
                title="已结算回款(NT$)"
                :value="24543.00"
                subtitle="订单: 13"
              />
            </el-col>
          </el-row>

          <!-- 中间：近期回款 -->
          <div class="recent-payments-section">
            <RecentPayments />
          </div>

          <!-- 下面：最近订单 -->
          <div class="recent-orders-section">
            <RecentOrders />
          </div>
        </div>
      </el-col>

      <!-- 右侧栏 -->
      <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
        <div class="right-column">
          <!-- 上面：回款余额 -->
          <div class="balance-section">
            <PaymentBalance />
          </div>

          <!-- 中间：我的店铺 -->
          <div class="stores-section">
            <MyStores />
          </div>

          <!-- 下面：本周回款排行榜 -->
          <div class="ranking-section">
            <WeeklyRanking />
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import KPICard from '../components/KPICard.vue'
import RecentPayments from '../components/RecentPayments.vue'
import PaymentBalance from '../components/PaymentBalance.vue'
import MyStores from '../components/MyStores.vue'
import RecentOrders from '../components/RecentOrders.vue'
import WeeklyRanking from '../components/WeeklyRanking.vue'
</script>

<style scoped lang="scss">
.operator-home {
  .welcome-section {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;
  }

  .welcome-left {
    flex: 1;
  }

  .welcome-right {
    flex-shrink: 0;
  }

  .welcome-info {
    h2 {
      font-size: 20px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 8px;
    }

    p {
      font-size: 14px;
      color: #909399;
      margin: 0;
    }

    .notification-link {
      color: #f90;
      cursor: pointer;
    }
  }

  .notification-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px 16px;
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 4px;

    .notification-text {
      font-size: 14px;
      color: #606266;
    }

    .close-btn {
      padding: 0;
      font-size: 18px;
      color: #909399;
    }
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

  .recent-payments-section,
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
