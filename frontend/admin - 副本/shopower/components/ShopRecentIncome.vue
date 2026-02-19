<template>
  <el-card class="income-card">
    <template #header>
      <div class="card-header">
        <span class="header-title">
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
    </template>
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
          <v-chart :option="chartOption" style="height: 200px" />
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const chartOption = ref({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'line',
      snap: true
    },
    formatter: (params: any) => {
      const first = Array.isArray(params) ? params[0] : params
      return String(first?.data ?? '')
    }
  },
  grid: {
    left: '5%',
    right: '15%',
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
        5500, 5100, 5800, 6200, 5900, 6500, 6100, 6800, 7200, 6900
      ],
      smooth: true,
      symbol: 'circle',
      symbolSize: 10,
      showSymbol: true,
      showAllSymbol: false,
      lineStyle: {
        color: '#ff6a3a',
        width: 2
      },
      itemStyle: {
        color: '#ff6a3a',
        opacity: 0
      },
      emphasis: {
        itemStyle: {
          opacity: 0
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
</script>

<style scoped lang="scss">
.income-card {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  font-size: 16px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-bar {
  width: 3px;
  height: 16px;
  background-color: #ff6a3a;
  border-radius: 2px;
}

.income-content {
  display: flex;
  gap: 16px;
}

.income-left {
  flex: 0 0 180px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.income-right {
  flex: 1;
}

.chart-container {
  width: 100%;
  margin-top: 28px;
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

// 第一个统计项（今日销售）使用更大的字体
.stat-item:first-child .stat-value {
  font-size: 28px;
  color: #ff6a3a;
}

.detail-button {
  color: #909399;
  display: flex;
  align-items: center;
  gap: 5px;
  
  &:hover {
    color: #606266;
  }
}

.arrow-icon {
  width: 12px;
  height: 12px;
}
</style>
