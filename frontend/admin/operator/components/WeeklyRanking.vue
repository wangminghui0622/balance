<template>
  <el-card class="ranking-card">
    <template #header>
      <div class="card-header">
        <span>本周回款排行榜</span>
        <el-button type="text" size="small">更多</el-button>
      </div>
    </template>
    <div class="ranking-list">
      <div
        v-for="(item, index) in rankingList"
        :key="index"
        class="ranking-item"
      >
        <div class="rank-number" :class="getRankClass(index)">
          {{ index + 1 }}
        </div>
        <div class="store-info">
          <div class="store-name">{{ item.storeName }}</div>
          <div class="revenue">近7天收益: NT${{ item.revenue }}</div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface RankingItem {
  storeName: string
  revenue: string
}

const rankingList = ref<RankingItem[]>([
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    revenue: '36050'
  },
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    revenue: '28500'
  },
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    revenue: '22100'
  },
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    revenue: '18900'
  }
])

const getRankClass = (index: number) => {
  if (index === 0) return 'rank-first'
  if (index === 1) return 'rank-second'
  if (index === 2) return 'rank-third'
  return ''
}
</script>

<style scoped lang="scss">
.ranking-card {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
  font-size: 16px;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  transition: background-color 0.3s;

  &:hover {
    background-color: #e4e7ed;
  }
}

.rank-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
  background-color: #909399;
  color: #fff;
  flex-shrink: 0;

  &.rank-first {
    background-color: #f56c6c;
  }

  &.rank-second {
    background-color: #e6a23c;
  }

  &.rank-third {
    background-color: #67c23a;
  }
}

.store-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.store-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  line-height: 1.4;
}

.revenue {
  font-size: 12px;
  color: #606266;
}
</style>
