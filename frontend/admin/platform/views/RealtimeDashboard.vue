<template>
  <div class="realtime-dashboard">
    <!-- 顶部导航栏 -->
    <div class="dashboard-header">
      <div class="header-left">
        <el-button :icon="ArrowLeft" circle size="small" @click="goBack" />
        <span class="user-name">Hector ▼</span>
      </div>
      <div class="header-center">
        <img src="" alt="XShopee" class="logo" />
        <span class="logo-text">XShopee</span>
      </div>
      <div class="header-right">
        <span class="sort-label">排序方式：店铺 ▼</span>
        <el-icon><Picture /></el-icon>
        <el-icon><FullScreen /></el-icon>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="dashboard-content">
      <!-- 左侧区域 -->
      <div class="left-section">
        <!-- 今日销售额卡片 -->
        <div class="sales-card">
          <div class="sales-main">
            <div class="sales-left">
              <div class="sales-header">
                <span class="sales-title">今日销售额</span>
              </div>
              <div class="sales-amount">NT$ {{ formatNumber(todaySales) }}</div>
            </div>
            <div class="sales-right">
              <div class="stat-item">
                <span class="stat-label">订单总数：</span>
                <span class="stat-value highlight">{{ totalOrders }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">活跃店铺：</span>
                <span class="stat-value highlight">{{ activeShops }}</span>
              </div>
            </div>
          </div>
          <div class="sales-footer">
            <span class="update-time">{{ currentTime }}</span>
            <div class="marquee-items">
              <div class="marquee-item">
                <el-icon><Shop /></el-icon>
                <span>跑马灯先占位 (+NT$ 600.00)</span>
              </div>
              <div class="marquee-item">
                <el-icon><Shop /></el-icon>
                <span>跑马灯先占位 (+NT$ 600.00)</span>
              </div>
              <div class="marquee-item">
                <el-icon><Shop /></el-icon>
                <span>跑马灯先占位</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 销售趋势图表 -->
        <div class="chart-card">
          <div class="chart-header">
            <span class="chart-title">销售趋势</span>
            <div class="chart-legend">
              <span class="legend-item today">● 今日</span>
              <span class="legend-item yesterday">● 昨日</span>
            </div>
          </div>
          <div class="chart-container" ref="chartRef"></div>
        </div>
      </div>

      <!-- 右侧区域 -->
      <div class="right-section">
        <!-- 昨日指标 -->
        <div class="stats-card">
          <div class="stats-title">昨日指标</div>
          <div class="stats-row">
            <div class="stats-item">
              <div class="stats-value">{{ formatNumber(yesterdaySales) }}</div>
              <div class="stats-label">订单销售额(NT$)</div>
            </div>
            <div class="stats-item">
              <div class="stats-value">{{ yesterdayOrders }}</div>
              <div class="stats-label">订单数量</div>
            </div>
            <div class="stats-item">
              <div class="stats-value">{{ yesterdayShops }}</div>
              <div class="stats-label">活跃店铺数</div>
            </div>
          </div>
        </div>

        <!-- 累计销售 -->
        <div class="stats-card">
          <div class="stats-title">累计销售</div>
          <div class="stats-row">
            <div class="stats-item">
              <div class="stats-value">{{ formatNumber(totalSalesAmount) }}</div>
              <div class="stats-label">订单销售额(NT$)</div>
            </div>
            <div class="stats-item">
              <div class="stats-value">{{ formatNumber(totalOrdersCount) }}</div>
              <div class="stats-label">订单数量</div>
            </div>
          </div>
        </div>

        <!-- 店铺销售排行榜 -->
        <div class="ranking-card">
          <div class="ranking-header">
            <span class="ranking-title">● 店铺销售排行榜</span>
          </div>
          <div class="ranking-list">
            <div 
              v-for="(shop, index) in shopRanking" 
              :key="index" 
              class="ranking-item"
            >
              <div class="ranking-left">
                <el-avatar :size="40" :src="shop.avatar" shape="square" />
                <div class="shop-info">
                  <div class="shop-name">{{ shop.name }}</div>
                  <div class="shop-id">店铺ID：{{ shop.id }}</div>
                </div>
              </div>
              <div class="ranking-right">
                <div class="shop-sales">NT${{ formatNumber(shop.sales) }}</div>
                <div class="shop-orders">{{ shop.orders }}单</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, Picture, FullScreen, Shop } from '@element-plus/icons-vue'
import * as echarts from 'echarts'

const router = useRouter()
const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

// 数据
const todaySales = ref(26000.00)
const totalOrders = ref(1433)
const activeShops = ref(42)
const currentTime = ref('')

const yesterdaySales = ref(22522.00)
const yesterdayOrders = ref(1433)
const yesterdayShops = ref(45)

const totalSalesAmount = ref(2622522.00)
const totalOrdersCount = ref(289433)

const shopRanking = ref([
  { avatar: '', name: '店铺名称示例文字占位符占位...', id: '1234567890', sales: 245534.50, orders: 3521 },
  { avatar: '', name: '店铺名称示例文字占位符占位...', id: '1234567890', sales: 245534.50, orders: 2821 },
  { avatar: '', name: '店铺名称示例文字占位符占位...', id: '1234567890', sales: 245534.50, orders: 2621 },
  { avatar: '', name: '店铺名称示例文字占位符占位...', id: '1234567890', sales: 245534.50, orders: 2221 },
  { avatar: '', name: '店铺名称示例文字占位符占位...', id: '1234567890', sales: 245534.50, orders: 1821 },
  { avatar: '', name: '店铺名称示例文字占位符占位...', id: '1234567890', sales: 245534.50, orders: 1621 },
])

const formatNumber = (num: number) => {
  return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const updateTime = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  const seconds = String(now.getSeconds()).padStart(2, '0')
  const offset = -now.getTimezoneOffset() / 60
  const offsetStr = offset >= 0 ? `+${String(offset).padStart(2, '0')}` : String(offset).padStart(3, '0')
  currentTime.value = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}(GMT${offsetStr})`
}

const goBack = () => {
  router.back()
}

const initChart = () => {
  if (!chartRef.value) return
  
  chartInstance = echarts.init(chartRef.value)
  
  const option = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(0, 0, 0, 0.7)',
      textStyle: { color: '#fff' }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: Array.from({ length: 24 }, (_, i) => i),
      axisLine: { lineStyle: { color: '#e0e0e0' } },
      axisLabel: { color: '#999' }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: '#f0f0f0' } },
      axisLabel: { color: '#999' }
    },
    series: [
      {
        name: '今日',
        type: 'line',
        smooth: true,
        data: [300, 400, 350, 500, 600, 800, 1200, 1500, 1800, 2000, 2200, 2500, 2800, 3000, 2800, 2500, 2200, 2000, 1800, 1500, 1200, 800, 500, 300],
        lineStyle: { color: '#ff6a3a', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(255, 106, 58, 0.3)' },
            { offset: 1, color: 'rgba(255, 106, 58, 0.05)' }
          ])
        },
        itemStyle: { color: '#ff6a3a' }
      },
      {
        name: '昨日',
        type: 'line',
        smooth: true,
        data: [200, 300, 280, 400, 500, 700, 1000, 1300, 1600, 1800, 2000, 2300, 2600, 2800, 2600, 2300, 2000, 1800, 1600, 1300, 1000, 700, 400, 200],
        lineStyle: { color: '#999', width: 2 },
        itemStyle: { color: '#999' }
      }
    ]
  }
  
  chartInstance.setOption(option)
}

let timeInterval: number | null = null

onMounted(() => {
  updateTime()
  timeInterval = window.setInterval(updateTime, 1000)
  
  setTimeout(() => {
    initChart()
  }, 100)
  
  window.addEventListener('resize', () => {
    chartInstance?.resize()
  })
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
  chartInstance?.dispose()
})
</script>

<style scoped lang="scss">
.realtime-dashboard {
  min-height: 100vh;
  background: linear-gradient(180deg, #ff7040 0%, #ff9060 20%, #ffb088 40%, #ffd4c0 60%, #fff0e8 80%, #fff 100%);
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background: transparent;
  color: #fff;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;

    .user-name {
      font-size: 14px;
    }
  }

  .header-center {
    display: flex;
    align-items: center;
    gap: 8px;

    .logo {
      height: 24px;
    }

    .logo-text {
      font-size: 18px;
      font-weight: bold;
      color: #ff6a3a;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .sort-label {
      font-size: 14px;
    }

    .el-icon {
      font-size: 20px;
      cursor: pointer;
    }
  }
}

.dashboard-content {
  display: flex;
  gap: 16px;
  padding: 16px;
}

.left-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.right-section {
  width: 360px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sales-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  color: #333;

  .sales-main {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .sales-left {
    .sales-header {
      margin-bottom: 8px;

      .sales-title {
        font-size: 16px;
        color: #666;
      }
    }

    .sales-amount {
      font-size: 48px;
      font-weight: bold;
      color: #ff6a3a;
    }
  }

  .sales-right {
    display: flex;
    flex-direction: column;
    gap: 12px;
    text-align: right;

    .stat-item {
      .stat-label {
        font-size: 14px;
        color: #666;
      }

      .stat-value {
        font-size: 18px;
        font-weight: bold;

        &.highlight {
          color: #ff6a3a;
        }
      }
    }
  }

  .sales-footer {
    display: flex;
    align-items: center;
    gap: 16px;
    padding-top: 16px;
    border-top: 1px solid #eee;

    .update-time {
      font-size: 12px;
      color: #999;
    }

    .marquee-items {
      display: flex;
      gap: 16px;
      flex: 1;
      overflow: hidden;

      .marquee-item {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        white-space: nowrap;
        background: #fff5f0;
        color: #ff6a3a;
        padding: 4px 8px;
        border-radius: 4px;
      }
    }
  }
}

.chart-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  flex: 1;

  .chart-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    .chart-title {
      font-size: 16px;
      font-weight: 500;
    }

    .chart-legend {
      display: flex;
      gap: 16px;

      .legend-item {
        font-size: 12px;
        color: #666;

        &.today {
          color: #ff6a3a;
        }

        &.yesterday {
          color: #999;
        }
      }
    }
  }

  .chart-container {
    height: 300px;
  }
}

.stats-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;

  .stats-title {
    font-size: 14px;
    color: #666;
    margin-bottom: 12px;
  }

  .stats-row {
    display: flex;
    gap: 16px;

    .stats-item {
      flex: 1;

      .stats-value {
        font-size: 20px;
        font-weight: bold;
        color: #333;
      }

      .stats-label {
        font-size: 12px;
        color: #999;
        margin-top: 4px;
      }
    }
  }
}

.ranking-card {
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  flex: 1;
  overflow: hidden;

  .ranking-header {
    margin-bottom: 12px;

    .ranking-title {
      font-size: 14px;
      font-weight: 500;
      color: #ff6a3a;
    }
  }

  .ranking-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-height: 400px;
    overflow-y: auto;

    .ranking-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 0;
      border-bottom: 1px solid #f0f0f0;

      &:last-child {
        border-bottom: none;
      }

      .ranking-left {
        display: flex;
        align-items: center;
        gap: 12px;

        .shop-info {
          .shop-name {
            font-size: 14px;
            color: #333;
            max-width: 150px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .shop-id {
            font-size: 12px;
            color: #999;
          }
        }
      }

      .ranking-right {
        text-align: right;

        .shop-sales {
          font-size: 14px;
          font-weight: bold;
          color: #ff6a3a;
        }

        .shop-orders {
          font-size: 12px;
          color: #999;
        }
      }
    }
  }
}
</style>
