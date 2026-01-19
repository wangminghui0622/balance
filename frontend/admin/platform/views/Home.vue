<template>
  <div class="platform-home">
    <!-- A：上半部分，平台用户数据 + 平台警示用户（左右结构） -->
    <el-row :gutter="20" class="row row-top">
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <PlatformUserData />
      </el-col>
      <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
        <PlatformAlertUsers />
      </el-col>
    </el-row>

    <!-- B：下半部分 -->
    <!-- B1：第一行，今日订单和店主充值预警左右排列 -->
    <el-row :gutter="20" class="row row-bottom-1">
      <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="16">
        <TodayOrders />
      </el-col>
      <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
        <ShopOwnerRechargeAlert />
      </el-col>
    </el-row>

    <!-- B2：第二行，运营业绩预警、运营违规预警、店主异常预警水平排列 -->
    <el-row :gutter="20" class="row row-bottom-2">
      <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
        <OperationalPerformanceAlert />
      </el-col>
      <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
        <OperationalViolationAlert />
      </el-col>
      <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
        <ShopOwnerAnomalyAlert />
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import PlatformUserData from '../components/PlatformUserData.vue'
import PlatformAlertUsers from '../components/PlatformAlertUsers.vue'
import TodayOrders from '../components/TodayOrders.vue'
import ShopOwnerRechargeAlert from '../components/ShopOwnerRechargeAlert.vue'
import OperationalPerformanceAlert from '../components/OperationalPerformanceAlert.vue'
import OperationalViolationAlert from '../components/OperationalViolationAlert.vue'
import ShopOwnerAnomalyAlert from '../components/ShopOwnerAnomalyAlert.vue'
</script>

<style scoped lang="scss">
.platform-home {
  .row {
    margin-bottom: 20px;
  }

  .row-bottom-1 {
    margin-bottom: 16px;
    display: flex;
    align-items: stretch; // 确保子元素等高

    :deep(.el-col) {
      display: flex;
      flex-direction: column;
    }
  }

  .row-bottom-2 {
    margin-bottom: 0;
  }
}
</style>
