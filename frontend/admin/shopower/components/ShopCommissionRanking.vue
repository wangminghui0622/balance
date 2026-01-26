<template>
  <el-card class="ranking-card">
    <template #header>
      <div class="card-header">
        <span class="header-title">
          <span class="title-bar"></span>
          <span>近7日佣金排行榜</span>
        </span>
        <a href="#" class="more-link">
          更多
          <svg class="arrow-icon" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </a>
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
          <div class="commission">近7天佣金: NT${{ item.commission }}</div>
        </div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface RankingItem {
  storeName: string
  commission: string
}

const rankingList = ref<RankingItem[]>([
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    commission: '36050'
  },
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    commission: '28500'
  },
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    commission: '22100'
  },
  {
    storeName: '店铺名称示例文字占位符文字占位符',
    commission: '18900'
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

.more-link {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 400;
  line-height: 1;
  color: #909399;
  text-decoration: none;
  
  &:hover {
    color: #606266;
  }
  
  .arrow-icon {
    width: 12px;
    height: 12px;
  }
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
  max-width: 210px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.commission {
  font-size: 12px;
  color: #606266;
}
</style>

