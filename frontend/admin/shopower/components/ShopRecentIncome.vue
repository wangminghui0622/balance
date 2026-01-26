<template>
  <el-card class="income-card">
    <template #header>
      <div class="card-header">
        <span class="header-title">
          <span class="title-bar"></span>
          <span>近期收益</span>
        </span>
        <el-button type="text" size="small">更多</el-button>
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

// 获取当前日期信息
const now = new Date()
const currentYear = now.getFullYear()
const currentMonth = now.getMonth() + 1
const currentDay = now.getDate()

// 获取当前月份的天数
const getDaysInMonth = (year: number, month: number) => {
  return new Date(year, month, 0).getDate()
}

const daysInCurrentMonth = getDaysInMonth(currentYear, currentMonth)

// 生成X轴数据（1到当前月份的最后一天）
const xAxisData = Array.from({ length: daysInCurrentMonth }, (_, i) => i + 1)

// 生成Y轴数据（只到今天为止有数据，之后为null）
const generateSeriesData = () => {
  const data: (number | null)[] = []
  for (let i = 1; i <= daysInCurrentMonth; i++) {
    if (i <= currentDay) {
      // 今天及之前的日期有数据（示例数据）
      data.push(1200 + Math.random() * 5000 + i * 200)
    } else {
      // 今天之后的日期没有数据
      data.push(null)
    }
  }
  return data
}

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
    data: xAxisData,
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
      data: generateSeriesData(),
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
      },
      markLine: {
        symbol: 'none',
        data: [
          {
            xAxis: currentDay,
            lineStyle: {
              color: 'rgba(144, 147, 153, 0.3)',
              width: 1,
              type: 'solid'
            },
            label: {
              show: false
            }
          }
        ]
      }
    }
  ]
})
</script>

<style scoped lang="scss">
.income-card {
  height: 100%;
  border-radius: 8px;
  overflow: hidden;
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
</style>
