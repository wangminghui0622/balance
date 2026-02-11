<template>
  <el-card class="orders-card">
    <template #header>
      <div class="card-header">
        <span>今日订单</span>
        <el-button type="text" size="small">更多</el-button>
      </div>
    </template>
    <div class="orders-content">
      <div class="orders-left">
        <div class="stat-item">
          <div class="stat-label">今日交易总额(NT$)</div>
          <div class="stat-value">8,420.00</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">订单数量</div>
          <div class="stat-value">450</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">取消单量</div>
          <div class="stat-value">54</div>
        </div>
      </div>
      <div class="orders-right">
        <div class="chart-container">
          <v-chart :option="chartOption" style="height: 200px; width: 100%" />
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const chartOption = ref({
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
    width: 'auto'
  },
  xAxis: {
    type: 'category',
    boundaryGap: false,
    data: Array.from({ length: 30 }, (_, i) => i + 1)
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      type: 'line',
      data: [
        200, 180, 220, 190, 210, 150, 180, 200, 250, 280, 270, 300, 320, 310,
        280, 260, 240, 270, 290, 310, 300, 280, 260, 250, 240, 230, 220, 210,
        200, 190
      ],
      smooth: true,
      lineStyle: {
        color: '#303133',
        width: 2
      },
      areaStyle: {
        color: {
          type: 'linear',
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: 'rgba(48, 49, 51, 0.3)' },
            { offset: 1, color: 'rgba(48, 49, 51, 0.1)' }
          ]
        }
      }
    }
  ]
})
</script>

<style scoped lang="scss">
.orders-card {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  font-size: 16px;
}

.orders-content {
  display: flex;
  gap: 16px;
  overflow-x: hidden;
}

.orders-left {
  flex: 0 0 180px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.orders-right {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.chart-container {
  width: 100%;
  overflow: hidden;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.stat-value {
  font-size: 20px;
  font-weight: bold;
  color: #303133;
}

// 第一个统计项（今日交易总额）使用更大的字体
.stat-item:first-child .stat-value {
  font-size: 28px;
}
</style>
