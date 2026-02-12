<template>
  <div class="report-page">
    <div class="page-header">
      <h2>汇总报表</h2>
    </div>

    <!-- 平台总账 -->
    <el-card class="section-card" shadow="never">
      <div class="section-title">平台总账</div>
      <div class="summary-row four-cols with-operators">
        <div class="summary-item">
          <div class="item-label">平台总金额</div>
          <div class="item-value">NT$5,450.00</div>
        </div>
        <span class="operator">=</span>
        <div class="summary-item">
          <div class="item-label">平台账户总额</div>
          <div class="item-value">NT$5,000.00</div>
        </div>
        <span class="operator">+</span>
        <div class="summary-item">
          <div class="item-label">店主账户总额</div>
          <div class="item-value">NT$5,000.00</div>
        </div>
        <span class="operator">-</span>
        <div class="summary-item">
          <div class="item-label">运营账户总额</div>
          <div class="item-value">NT$5,000.00</div>
        </div>
      </div>
    </el-card>

    <!-- 平台账户 和 店主账户 -->
    <div class="two-col-section">
      <el-card class="section-card half" shadow="never">
        <div class="section-header">
          <span class="section-title">平台账户</span>
          <el-link type="primary">详情</el-link>
        </div>
        <div class="summary-row three-cols">
          <div class="summary-item small">
            <div class="item-label with-dot platform">平台托管账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
          <div class="summary-item small">
            <div class="item-label">平台佣金账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
          <div class="summary-item small">
            <div class="item-label">平台罚补账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
        </div>
      </el-card>

      <el-card class="section-card half" shadow="never">
        <div class="section-header">
          <span class="section-title">店主账户</span>
          <el-link type="primary">详情</el-link>
        </div>
        <div class="summary-row four-cols">
          <div class="summary-item small">
            <div class="item-label with-dot owner1">店主订单未付款金额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
          <div class="summary-item small">
            <div class="item-label with-dot owner2">店主预付款账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
          <div class="summary-item small">
            <div class="item-label">店主佣金账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
          <div class="summary-item small">
            <div class="item-label with-dot owner3">店主保证金账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 运营账户 和 图表 -->
    <div class="two-col-section">
      <el-card class="section-card half" shadow="never">
        <div class="section-header">
          <span class="section-title">运营账户</span>
        </div>
        <div class="summary-row two-cols">
          <div class="summary-item small">
            <div class="item-label with-dot operator1">运营回款账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
          <div class="summary-item small">
            <div class="item-label with-dot operator2">运营保证金账户余额</div>
            <div class="item-value">NT$5,450.00</div>
          </div>
        </div>
        <div class="chart-container">
          <div class="chart-label">金额</div>
          <div class="line-chart" ref="lineChartRef"></div>
          <div class="chart-x-axis">
            <span v-for="i in 7" :key="i">{{ (i - 1) * 5 || 1 }}</span>
            <span>30 时间/天</span>
          </div>
        </div>
      </el-card>

      <el-card class="section-card half" shadow="never">
        <div class="chart-container bar">
          <div class="bar-chart" ref="barChartRef"></div>
          <div class="chart-x-axis">
            <span v-for="i in 7" :key="i">{{ (i - 1) * 5 || 1 }}</span>
            <span>30</span>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const lineChartRef = ref<HTMLElement>()
const barChartRef = ref<HTMLElement>()

onMounted(() => {
  drawLineChart()
  drawBarChart()
})

const drawLineChart = () => {
  if (!lineChartRef.value) return
  const canvas = document.createElement('canvas')
  canvas.width = 400
  canvas.height = 150
  lineChartRef.value.appendChild(canvas)
  
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  ctx.strokeStyle = '#303133'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(20, 120)
  ctx.lineTo(60, 100)
  ctx.lineTo(100, 90)
  ctx.lineTo(140, 85)
  ctx.lineTo(180, 75)
  ctx.lineTo(220, 60)
  ctx.lineTo(260, 50)
  ctx.lineTo(300, 35)
  ctx.lineTo(340, 25)
  ctx.stroke()
  
  ctx.strokeStyle = '#909399'
  ctx.beginPath()
  ctx.moveTo(20, 130)
  ctx.lineTo(60, 115)
  ctx.lineTo(100, 105)
  ctx.lineTo(140, 100)
  ctx.lineTo(180, 90)
  ctx.lineTo(220, 80)
  ctx.lineTo(260, 70)
  ctx.lineTo(300, 55)
  ctx.lineTo(340, 45)
  ctx.stroke()
}

const drawBarChart = () => {
  if (!barChartRef.value) return
  const canvas = document.createElement('canvas')
  canvas.width = 450
  canvas.height = 250
  barChartRef.value.appendChild(canvas)
  
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  
  const barWidth = 8
  const gap = 4
  const groupGap = 20
  const startX = 30
  const baseY = 220
  
  const data = [
    [80, 60, 90, 70],
    [100, 80, 110, 85],
    [70, 50, 75, 60],
    [120, 100, 130, 110],
    [90, 70, 95, 75],
    [85, 65, 88, 68],
    [110, 90, 115, 95],
    [75, 55, 80, 62],
    [95, 75, 100, 80],
    [105, 85, 110, 88],
    [88, 68, 92, 72],
    [115, 95, 120, 100]
  ]
  
  const colors = ['#303133', '#606266', '#909399', '#C0C4CC']
  
  data.forEach((group, groupIndex) => {
    group.forEach((value, barIndex) => {
      const x = startX + groupIndex * (barWidth * 4 + gap * 3 + groupGap) + barIndex * (barWidth + gap)
      ctx.fillStyle = colors[barIndex]
      ctx.fillRect(x, baseY - value, barWidth, value)
    })
  })
}
</script>

<style scoped lang="scss">
.report-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100%;
}

.page-header {
  margin-bottom: 20px;
  
  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 500;
    color: #303133;
  }
}

.section-card {
  margin-bottom: 20px;
  
  &.half {
    flex: 1;
  }
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 16px;
}

.section-header .section-title {
  margin-bottom: 0;
}

.two-col-section {
  display: flex;
  gap: 20px;
}

.summary-row {
  display: flex;
  gap: 20px;
  align-items: center;
  
  &.four-cols .summary-item {
    flex: 1;
  }
  
  &.three-cols .summary-item {
    flex: 1;
  }
  
  &.two-cols .summary-item {
    flex: 1;
  }
  
  &.with-operators {
    gap: 10px;
    
    .operator {
      font-size: 20px;
      color: #909399;
      font-weight: 300;
    }
  }
}

.summary-item {
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 16px 20px;
  
  &.small {
    padding: 12px 16px;
  }
  
  .item-label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 8px;
    
    &.with-dot {
      display: flex;
      align-items: center;
      
      &::before {
        content: '';
        width: 8px;
        height: 8px;
        border-radius: 50%;
        margin-right: 6px;
      }
      
      &.platform::before {
        background: #303133;
      }
      
      &.owner1::before {
        background: #303133;
      }
      
      &.owner2::before {
        background: #606266;
      }
      
      &.owner3::before {
        background: #C0C4CC;
      }
      
      &.operator1::before {
        background: #303133;
      }
      
      &.operator2::before {
        background: #909399;
      }
    }
  }
  
  .item-value {
    font-size: 20px;
    font-weight: 500;
    color: #303133;
  }
}

.chart-container {
  margin-top: 20px;
  
  .chart-label {
    font-size: 12px;
    color: #909399;
    margin-bottom: 8px;
  }
  
  .line-chart,
  .bar-chart {
    width: 100%;
    min-height: 150px;
  }
  
  .bar-chart {
    min-height: 250px;
  }
  
  .chart-x-axis {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: #909399;
    margin-top: 8px;
    padding: 0 20px;
  }
  
  &.bar {
    margin-top: 0;
    padding-top: 20px;
  }
}
</style>
